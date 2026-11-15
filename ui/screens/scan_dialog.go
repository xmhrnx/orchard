package screens

import (
	"fmt"
	"path/filepath"

	"github.com/xmhrnx/giode"
	"github.com/xmhrnx/giode/components/dialog"
	"github.com/xmhrnx/giode/components/scroll"

	"orchard/ui/components"
	"orchard/ui/env"
	"orchard/ui/theme"
)

// ScanDialog lists the repositories found under a folder and adds the
// ones the user keeps checked.
type ScanDialog struct {
	env      *env.Env
	open     bool
	root     string
	scanning bool
	found    []string
	sel      *components.Selection
	dlg      *dialog.Dialog
	list     *scroll.Scroll
	add      *giode.ButtonElement
	cancel   *giode.ButtonElement
}

// NewScanDialog creates the dialog. Create it once.
func NewScanDialog(e *env.Env) *ScanDialog {
	d := &ScanDialog{env: e, sel: components.NewSelection()}
	d.list = giode.Scroll(d.rows, giode.Styles{ScrollBar: giode.ScrollBar{Width: 6, TrackColor: "transparent"}}).FitContent(true)
	d.add = components.Button("Add selected", "plus", components.Primary, d.addSelected)
	d.cancel = components.Button("Cancel", "", components.Secondary, func() { d.open = false })
	d.dlg = giode.Dialog(&d.open, d.content, components.DialogPanel())
	return d
}

// Start scans root and opens the dialog with what it finds.
func (d *ScanDialog) Start(root string) {
	d.root, d.scanning, d.found, d.open = root, true, nil, true
	d.sel.Clear()
	d.env.WS.Scan(root, func(found []string, err error) {
		d.scanning, d.found = false, found
		d.sel.SetAll(found, true)
		d.env.Fail("Scanning", err)
	})
}

func (d *ScanDialog) addSelected() {
	paths := d.sel.Keys()
	d.open = false
	d.env.WS.AddPaths(paths, func(added int, err error) {
		d.env.Done(fmt.Sprintf("Added %d repositories", added), err)
	})
}

func (d *ScanDialog) rows() giode.Element {
	rows := make([]giode.Element, 0, len(d.found))
	for _, p := range d.found {
		rel, err := filepath.Rel(d.root, p)
		if err != nil {
			rel = p
		}
		rows = append(rows, components.SelectRow("scan", p, d.sel, func() giode.Element {
			return giode.Stack(giode.Styles{Gap: 2},
				giode.Text(filepath.Base(p), theme.Strong),
				giode.Text(rel, theme.Ellipsis(theme.CodeDim)),
			)
		}))
	}
	return giode.Stack(giode.Styles{Width: -1, Gap: 2}, rows...)
}

func (d *ScanDialog) content() giode.Element {
	var body giode.Element
	switch {
	case d.scanning:
		body = giode.Row(giode.Styles{Gap: 10, Align: giode.AlignCenter, Padding: giode.SymmetricInset(0, 24)},
			giode.Spinner(giode.Styles{Width: 18, Height: 18, Color: theme.Primary}),
			giode.Text("Looking for repositories…", theme.Caption))
	case len(d.found) == 0:
		body = giode.Text("No new repositories here.", theme.Caption)
	default:
		body = giode.Box(giode.Styles{Width: -1, MaxHeight: 380}, d.list)
	}
	d.add.Disabled(d.sel.Count() == 0)
	return giode.Stack(giode.Styles{Gap: 14, Width: 560},
		giode.Text("Repositories found", theme.Heading),
		giode.Text(d.root, theme.Ellipsis(theme.CodeDim)),
		body,
		giode.Row(giode.Styles{Width: -1, Gap: 8, Justify: giode.JustifyEnd}, d.cancel, d.add),
	)
}

// Layout draws the dialog.
func (d *ScanDialog) Layout(gtx giode.Context) giode.Dimensions { return d.dlg.Layout(gtx) }
