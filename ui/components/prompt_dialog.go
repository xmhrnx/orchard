package components

import (
	"strings"

	"github.com/xmhrnx/giode"
	"github.com/xmhrnx/giode/components/dialog"
	"github.com/xmhrnx/giode/components/input"

	"orchard/ui/theme"
)

// PromptDialog asks for one line of text: a branch name, a tag, a
// commit message.
type PromptDialog struct {
	open         bool
	title, label string
	onSubmit     func(string)
	field        *input.Input
	dlg          *dialog.Dialog
	ok, cancel   *giode.ButtonElement
}

// NewPromptDialog creates the dialog. Create it once.
func NewPromptDialog() *PromptDialog {
	d := &PromptDialog{}
	d.field = giode.Input.Text("", FieldStyle()).Submit(func(string) { d.submit() })
	d.ok = Button("Save", "check", Primary, d.submit)
	d.cancel = Button("Cancel", "", Secondary, func() { d.open = false })
	d.dlg = giode.Dialog(&d.open, d.content, DialogPanel())
	return d
}

// Ask opens the dialog with a prefilled value; fn gets the trimmed text
// when the user saves something non-empty.
func (d *PromptDialog) Ask(title, label, placeholder, initial, action string, fn func(string)) {
	d.title, d.label, d.onSubmit = title, label, fn
	d.field.Placeholder(placeholder)
	d.field.SetText(initial)
	d.field.SelectAll()
	d.ok.Label(action)
	d.open = true
	d.field.Focus()
}

func (d *PromptDialog) submit() {
	text := strings.TrimSpace(d.field.Text())
	if text == "" {
		return
	}
	d.open = false
	d.onSubmit(text)
}

func (d *PromptDialog) content() giode.Element {
	return giode.Stack(giode.Styles{Gap: 16, Width: 440},
		giode.Text(d.title, theme.Heading),
		Field(d.label, "", d.field),
		giode.Row(giode.Styles{Width: -1, Gap: 8, Justify: giode.JustifyEnd}, d.cancel, d.ok),
	)
}

// Layout draws the dialog.
func (d *PromptDialog) Layout(gtx giode.Context) giode.Dimensions { return d.dlg.Layout(gtx) }
