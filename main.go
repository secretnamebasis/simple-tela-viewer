package main

import (
	"crypto/rand"
	"os"
	"os/exec"
	"runtime"
	"time"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/layout"
	"fyne.io/fyne/v2/widget"
	"github.com/civilware/tela"
)

func main() {

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
	entry_scid := widget.NewEntry()
	entry_scid.SetPlaceHolder("scid")

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

		// get the endpoint
		endpoint := entry_endpoint.Text

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
