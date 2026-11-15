package components

import (
	"github.com/xmhrnx/giode"

	"orchard/ui/theme"
)

// StatTile is a headline number with a label, tinted by meaning.
func StatTile(icon, label, value string, tint, tintSoft giode.Color) giode.Element {
	return Frame(giode.Styles{FlexGrow: 1, MinWidth: 150, Padding: giode.UniformInset(16), Gap: 10, BorderRadius: 14},
		giode.Row(giode.Styles{Gap: 8, Align: giode.AlignCenter, Wrap: giode.WrapNoWrap},
			giode.Box(giode.Styles{Padding: giode.UniformInset(7), BorderRadius: 9, Background: tintSoft}, Glyph(icon, 15, tint)),
			giode.Text(label, theme.Ellipsis(theme.Caption)),
		),
		giode.Text(value, giode.Styles{FontSize: 26, FontWeight: giode.FontWeightBold, Color: theme.Text}),
	)
}
