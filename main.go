// Command orchard is a desktop app for keeping every git repository on
// a computer up to date and in good shape.
package main

import (
	"image"
	"log"

	"github.com/xmhrnx/giode"

	"orchard/assets"
	"orchard/internal/core"
	"orchard/internal/store"
	"orchard/ui"
	"orchard/ui/theme"
)

func main() {
	giode.SecureStore.Service = "com.orchard.app"
	theme.Install()
	if err := theme.LoadFonts(); err != nil {
		log.Fatalf("loading fonts: %v", err)
	}
	if err := theme.LoadIcons(); err != nil {
		log.Fatalf("loading icons: %v", err)
	}
	st, err := store.Open()
	if err != nil {
		log.Fatalf("locating settings: %v", err)
	}

	var app *giode.App
	ws, loadErr := core.New(func(fn func()) { app.Post(fn) }, st, giode.SecureStore)
	theme.SetDark(ws.Settings().Theme != "light")

	var root *ui.Root
	app = giode.New(giode.Config{
		Title:      "Orchard",
		Size:       image.Pt(1360, 860),
		MinSize:    image.Pt(980, 640),
		Resizable:  true,
		Frameless:  true,
		Center:     true,
		Background: "theme.background",
		Icon:       assets.Icon,
		Font:       theme.Sans,
		SingleInstanceLock: &giode.SingleInstanceLock{
			UniqueID: "12447727-40df-497e-a756-77312bcc2ccf",
		},
		OnReady: func() {
			ws.Refresh()
			ws.StartAutoFetch(root.Env().ReportFailures)
			root.Env().Fail("Loading saved state", loadErr)
		},
		OnFileDrop: func(paths []string) {
			ws.AddPaths(paths, func(added int, err error) {
				root.Env().Done("Added dropped repositories", err)
			})
		},
		OnShutdown: ws.Close,
	})
	root = ui.New(app, ws)
	if err := app.Run(root.View); err != nil {
		log.Fatal(err)
	}
}
