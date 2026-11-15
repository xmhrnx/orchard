package components

import (
	"github.com/xmhrnx/giode"

	"orchard/ui/theme"
)

// Field labels a form control, with an optional hint underneath.
func Field(label, hint string, control giode.Element) giode.Element {
	kids := []giode.Element{giode.Text(label, theme.Tinted(theme.Strong, theme.Text)), control}
	if hint != "" {
		kids = append(kids, giode.Text(hint, theme.Caption))
	}
	return giode.Stack(giode.Styles{Width: -1, Gap: 6}, kids...)
}
