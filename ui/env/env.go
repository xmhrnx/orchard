// Package env is what every screen needs from the app: the window, the
// workspace, notifications and navigation.
package env

import (
	"context"
	"errors"
	"fmt"
	"os/exec"
	"runtime"

	"github.com/xmhrnx/giode"

	"orchard/internal/core"
)

// Env is shared by all screens. It is used on the UI goroutine only.
type Env struct {
	App    *giode.App
	WS     *core.Workspace
	Toasts *giode.ToasterElement
	Nav    *Navigator
}

// Fail shows an error toast; a nil error does nothing.
func (e *Env) Fail(what string, err error) {
	if err == nil {
		return
	}
	e.Toasts.Show(giode.ToastError, fmt.Sprintf("%s: %v", what, err))
}

// Done shows a success toast, or an error toast when err is set.
func (e *Env) Done(success string, err error) {
	if err != nil {
		e.Toasts.Show(giode.ToastError, err.Error())
		return
	}
	e.Toasts.Show(giode.ToastSuccess, success)
}

// Report shows the summary of a bulk operation.
func (e *Env) Report(r core.Result) {
	kind := giode.ToastSuccess
	if len(r.Failed) > 0 {
		kind = giode.ToastWarning
		if r.Done == 0 {
			kind = giode.ToastError
		}
	}
	e.Toasts.Show(kind, r.Summary())
}

// ReportFailures shows a bulk operation's summary only when something
// failed, for background work the user didn't start.
func (e *Env) ReportFailures(r core.Result) {
	if len(r.Failed) > 0 {
		e.Report(r)
	}
}

// PickFolder asks for a folder and calls done with it on the UI
// goroutine; cancelling calls nothing.
func (e *Env) PickFolder(title, start string, done func(string)) {
	go func() {
		dir, err := e.App.PickFolder(context.Background(), giode.FolderOptions{Title: title, Directory: start})
		e.App.Post(func() {
			switch {
			case errors.Is(err, giode.ErrCanceled):
			case err != nil:
				e.Fail("Choosing a folder", err)
			default:
				done(dir)
			}
		})
	}()
}

// OpenURL opens a web page in the browser.
func (e *Env) OpenURL(url string) {
	go func() {
		err := e.App.OpenURL(context.Background(), url)
		e.App.Post(func() { e.Fail("Opening "+url, err) })
	}()
}

// Reveal opens a folder in the system's file manager.
func (e *Env) Reveal(path string) {
	go func() {
		err := openWith(fileManager(), path)
		e.App.Post(func() { e.Fail("Opening folder", err) })
	}()
}

// Terminal opens a terminal window in a folder.
func (e *Env) Terminal(path string) {
	go func() {
		err := openTerminal(path)
		e.App.Post(func() { e.Fail("Opening terminal", err) })
	}()
}

// Copy puts text on the clipboard and confirms it.
func (e *Env) Copy(label, text string) {
	e.App.SetClipboard(text)
	e.Toasts.Show(giode.ToastInfo, label+" copied")
}

func fileManager() string {
	switch runtime.GOOS {
	case "windows":
		return "explorer"
	case "darwin":
		return "open"
	}
	return "xdg-open"
}

func openWith(program string, args ...string) error {
	return exec.Command(program, args...).Start()
}
