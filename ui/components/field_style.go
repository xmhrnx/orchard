package components

import (
	"github.com/xmhrnx/giode"

	"orchard/ui/theme"
)

// FieldStyle is the look of text fields and selects in forms.
func FieldStyle() giode.Styles {
	return giode.Styles{
		Background: theme.Sunken, Color: theme.Text, BorderColor: theme.Hairline, BorderWidth: 1,
		BorderRadius: theme.Radius, Padding: giode.SymmetricInset(12, 9), FontSize: 13.5,
	}
}
