package components

import (
	"github.com/xmhrnx/giode"

	"orchard/ui/theme"
)

// Pill is a small rounded status label with an optional glyph.
func Pill(icon, text string, fg, bg giode.Color) giode.Element {
	kids := make([]giode.Element, 0, 2)
	if icon != "" {
		kids = append(kids, Glyph(icon, 12, fg))
	}
	kids = append(kids, giode.Text(text, giode.Styles{FontSize: 11.5, FontWeight: giode.FontWeightMedium, Color: fg}))
	return giode.Row(giode.Styles{
		Gap: 4, Align: giode.AlignCenter, Wrap: giode.WrapNoWrap,
		Padding: giode.SymmetricInset(8, 3), BorderRadius: theme.RadiusPill, Background: bg,
	}, kids...)
}
