package components

import (
	"github.com/xmhrnx/giode"

	"orchard/ui/theme"
)

// EmptyState fills an empty list with a friendly explanation and the
// action that fills it; action may be nil.
func EmptyState(icon, title, body string, action giode.Element) giode.Element {
	kids := []giode.Element{
		giode.Box(giode.Styles{
			Padding: giode.UniformInset(18), BorderRadius: theme.RadiusPill,
			Background: theme.PrimarySoft, BorderWidth: 1, BorderColor: giode.Mix("theme.primary", "theme.surface", 0.3),
		}, Glyph(icon, 30, theme.Primary)),
		giode.Text(title, theme.Heading),
		giode.Text(body, giode.Styles{FontSize: 13, Color: theme.Muted, TextAlign: giode.TextAlignCenter, MaxWidth: 380}),
	}
	if action != nil {
		kids = append(kids, giode.Box(giode.Styles{Margin: giode.Inset{Top: 6}}, action))
	}
	return giode.Stack(giode.Styles{Width: -1, Align: giode.AlignCenter, Gap: 10, Padding: giode.SymmetricInset(24, 48)}, kids...)
}
