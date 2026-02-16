package main

import (
	"crypto/rand"
	"crypto/tls"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"os"
	"os/exec"
	"runtime"
	"strings"
	"time"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/layout"
	"fyne.io/fyne/v2/widget"
	"github.com/civilware/Gnomon/structures"
	"github.com/civilware/tela"
	"github.com/gorilla/websocket"
)

var conn *websocket.Conn
var ws = "127.0.0.1:9190"

func set_gnomon_conn() error {

	url := "ws://" + ws + "/ws"
	dialer := websocket.Dialer{TLSClientConfig: &tls.Config{
		InsecureSkipVerify: true, // allow self-signed certs
	}}
	var err error
	conn, _, err = dialer.Dial(url, nil)
	if err != nil {
		return err
	}

	return nil
}

type getAllSCIDsByClassParams struct {
	Class string `json:"class"`
}
type getAllSCIDsByClassResult struct {
	Result []any `json:"result"`
}

func getAllSCIDsByClass(params getAllSCIDsByClassParams) (getAllSCIDsByClassResult, error) {

	msg := map[string]any{
		"jsonrpc": "2.0",
		"method":  "GetAllSCIDsByClass",
		"id":      "1",
		"params":  params,
	}
	b, err := json.Marshal(msg)
	if err != nil {
		return getAllSCIDsByClassResult{}, errors.New("failed to marshal")
	}
	var r structures.JSONRpcResp
	if err := json.Unmarshal(postBytes(b), &r); err != nil {
		return getAllSCIDsByClassResult{}, errors.New("failed to unmarshal")
	}

	return getAllSCIDsByClassResult{r.Result.([]any)}, nil
}

type getAllClassesResult struct {
	Result []any `json:"result"`
}

func getAllClasses() (getAllClassesResult, error) {

	msg := map[string]any{
		"jsonrpc": "2.0",
		"method":  "GetAllClasses",
		"id":      "1",
	}
	b, err := json.Marshal(msg)
	if err != nil {
		return getAllClassesResult{}, errors.New("failed to marshal")
	}
	var r structures.JSONRpcResp
	if err := json.Unmarshal(postBytes(b), &r); err != nil {
		return getAllClassesResult{}, errors.New("failed to unmarshal")
	}

	return getAllClassesResult{r.Result.([]any)}, nil
}

type getAllSCIDsAndHeadersResult struct {
	Result map[string]any `json:"result"`
}

func getAllSCIDsAndHeaders() (getAllSCIDsAndHeadersResult, error) {

	msg := map[string]any{
		"jsonrpc": "2.0",
		"method":  "GetAllSCIDsAndHeaders",
		"id":      "1",
	}
	b, err := json.Marshal(msg)
	if err != nil {
		return getAllSCIDsAndHeadersResult{}, errors.New("failed to marshal")
	}
	var r structures.JSONRpcResp
	if err := json.Unmarshal(postBytes(b), &r); err != nil {
		return getAllSCIDsAndHeadersResult{}, errors.New("failed to unmarshal")
	}

	return getAllSCIDsAndHeadersResult{r.Result.(map[string]any)}, nil
}
func postBytes(b []byte) []byte {
	// fmt.Println(string(b))
	err := conn.WriteMessage(websocket.TextMessage, b)
	if err != nil {
		panic(err)
	}

	_, msg, err := conn.ReadMessage()
	if err != nil {
		panic(err)
	}
	// fmt.Println(string(msg))
	return msg
}

func main() {
	for _, each := range os.Args {
		if strings.Contains(each, "--ws-address") {
			ws = strings.Split(each, "=")[1]
		}
	}
	if err := set_gnomon_conn(); err != nil {
		log.Fatal(err)
	}
	r, err := getAllSCIDsAndHeaders()
	if err != nil {
		log.Fatal(err)
	}
	scids_and_headers := r.Result
	// for scid, header := range scids_and_headers {
	// 	fmt.Println(scid, header.(string))
	// }
	allClasses, err := getAllClasses()
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println(allClasses.Result...)
	re, err := getAllSCIDsByClass(getAllSCIDsByClassParams{Class: "TELA-INDEX-1"})
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println(re.Result)
	collection := map[string]string{}
	for _, each := range re.Result {
		// fmt.Println(each)
		header, ok := scids_and_headers[each.(string)]
		if !ok {
			continue
		}
		collection[each.(string)] = header.(string)
	}
	// make a simple app in fyne with some random id
	myApp := app.NewWithID(rand.Text())

	// title the window
	myWindow := myApp.NewWindow("simple-tela-viewer")

	// make the app kind of a small square
	myWindow.Resize(fyne.NewSize(300, 300))
	myWindow.CenterOnScreen()

	// make a way to submit a node endpint
	entry_endpoint := widget.NewEntry()
	entry_endpoint.SetPlaceHolder("127.0.0.1:10102")

	// make a way to submit a scid to serve
	search := widget.NewEntry()
	search.SetPlaceHolder("search")
	entry_scid := widget.NewEntry()
	entry_scid.SetPlaceHolder("scid")
	search.OnChanged = func(s string) {
		for scid, each := range collection {
			if strings.Contains(each, s) {
				fmt.Println(scid, each)
				entry_scid.SetText(scid)
			}
		}
	}

	// load them into a form for processing
	submit := widget.NewForm(
		widget.NewFormItem("node", entry_endpoint),
		widget.NewFormItem("tela", entry_scid),
	)

	// if they cancel, close the application
	submit.OnCancel = func() { os.Exit(0) }

	// and if they submit...
	submit.OnSubmit = func() {

		// get the scid
		scid := entry_scid.Text

		// clean up
		scid = strings.TrimSpace(scid)

		// get the endpoint
		endpoint := entry_endpoint.Text

		// clean up
		endpoint = strings.TrimSpace(endpoint)

		// allow for updates, else a panic can occur
		tela.AllowUpdates(true)

		// clone the tela to the tmp directory and serve the filesystem
		url, err := tela.ServeTELA(scid, endpoint)
		if err != nil {
			dialog.ShowError(err, myWindow)
			return
		}

		// open a browser on a separate thread
		go openBrowser(url)
	}
	entry_endpoint.OnSubmitted = func(s string) { submit.OnSubmit() }
	entry_scid.OnSubmitted = func(s string) { submit.OnSubmit() }

	// put the content in a box
	content := container.NewVBox(
		layout.NewSpacer(),
		submit,
		search,
		layout.NewSpacer(),
	)

	myWindow.SetContent(content)
	myWindow.ShowAndRun()

}

func openBrowser(url string) {
	time.Sleep(200 * time.Millisecond)
	var cmd *exec.Cmd
	switch runtime.GOOS {
	case "windows":
		cmd = exec.Command("rundll32", "url.dll,FileProtocolHandler", url)
	case "darwin":
		cmd = exec.Command("open", url)
	default:
		cmd = exec.Command("xdg-open", url)
	}
	err := cmd.Start()
	if err != nil {
		panic(err)
	}
}
