package screens

import (
	"github.com/xmhrnx/giode"
	"github.com/xmhrnx/giode/components/input"

	"orchard/ui/components"
	"orchard/ui/env"
)

// FolderField is a path field with a button that opens the system's
// folder picker.
type FolderField struct {
	field  *input.Input
	browse *giode.ButtonElement
}

// NewFolderField creates the field. Create it once.
func NewFolderField(e *env.Env, title string) *FolderField {
	f := &FolderField{}
	f.field = giode.Input.Text("Choose a folder", components.FieldStyle())
	f.browse = components.Button("Browse", "folder", components.Secondary, func() {
		e.PickFolder(title, f.field.Text(), f.field.SetText)
	})
	return f
}

// Path is the folder entered.
func (f *FolderField) Path() string { return f.field.Text() }

// SetPath fills the field.
func (f *FolderField) SetPath(p string) { f.field.SetText(p) }

// Layout draws the field.
func (f *FolderField) Layout(gtx giode.Context) giode.Dimensions {
	return giode.Row(giode.Styles{Width: -1, Gap: 8, Align: giode.AlignCenter, Wrap: giode.WrapNoWrap},
		giode.Box(giode.Styles{FlexGrow: 1}, f.field), f.browse).Layout(gtx)
}
