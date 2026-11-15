package components

import (
	"github.com/xmhrnx/giode"
	"github.com/xmhrnx/giode/components/dialog"

	"orchard/ui/theme"
)

// ConfirmDialog asks before doing something that can't be undone.
type ConfirmDialog struct {
	open          bool
	title, body   string
	danger        bool
	onConfirm     func()
	dlg           *dialog.Dialog
	confirm, warn *giode.ButtonElement
	cancel        *giode.ButtonElement
}

// NewConfirmDialog creates the dialog. Create it once and lay it out
// anywhere; it paints over the window while open.
func NewConfirmDialog() *ConfirmDialog {
	d := &ConfirmDialog{}
	run := func() {
		d.open = false
		if d.onConfirm != nil {
			d.onConfirm()
		}
	}
	d.confirm = Button("Confirm", "", Primary, run)
	d.warn = Button("Confirm", "", Destructive, run)
	d.cancel = Button("Cancel", "", Secondary, func() { d.open = false })
	d.dlg = giode.Dialog(&d.open, d.content, DialogPanel())
	return d
}

// Ask opens the dialog; fn runs if the user confirms.
func (d *ConfirmDialog) Ask(title, body, confirmLabel string, danger bool, fn func()) {
	d.title, d.body, d.danger, d.onConfirm = title, body, danger, fn
	d.confirm.Label(confirmLabel)
	d.warn.Label(confirmLabel)
	d.open = true
}

func (d *ConfirmDialog) content() giode.Element {
	icon, tint, soft := "alert", theme.Warning, theme.AccentSoft
	ok := d.confirm
	if d.danger {
		icon, tint, soft, ok = "trash", theme.Danger, theme.DangerSoft, d.warn
	}
	return giode.Stack(giode.Styles{Gap: 14, Width: 420},
		giode.Row(giode.Styles{Gap: 12, Align: giode.AlignCenter, Wrap: giode.WrapNoWrap},
			giode.Box(giode.Styles{Padding: giode.UniformInset(9), BorderRadius: 12, Background: soft}, Glyph(icon, 20, tint)),
			giode.Text(d.title, theme.Heading),
		),
		giode.Text(d.body, giode.Styles{FontSize: 13, Color: theme.Muted, LineHeight: 19}),
		giode.Row(giode.Styles{Width: -1, Gap: 8, Justify: giode.JustifyEnd}, d.cancel, ok),
	)
}

// Layout draws the dialog.
func (d *ConfirmDialog) Layout(gtx giode.Context) giode.Dimensions { return d.dlg.Layout(gtx) }

// DialogPanel is the panel style shared by Orchard's dialogs.
func DialogPanel() giode.Styles {
	return giode.Styles{
		Background: theme.Surface, BorderRadius: theme.RadiusLarge, Padding: giode.UniformInset(22),
		BorderWidth: 1, BorderColor: giode.Mix("theme.primary", "theme.border", 0.2),
		BoxShadow: giode.BoxShadow{Y: 24, Blur: 60, Color: "#00000088"},
	}
}
