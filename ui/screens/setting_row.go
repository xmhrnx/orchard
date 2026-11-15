package screens

import (
	"github.com/xmhrnx/giode"

	"orchard/ui/theme"
)

// SettingRow lays out one preference: what it is on the left, its
// control on the right.
func SettingRow(title, about string, control giode.Element) giode.Element {
	return giode.Row(giode.Styles{Width: -1, Gap: 24, Align: giode.AlignCenter, Wrap: giode.WrapNoWrap, Padding: giode.SymmetricInset(0, 4)},
		giode.Stack(giode.Styles{FlexGrow: 1, Gap: 3},
			giode.Text(title, theme.Strong),
			giode.Text(about, giode.Styles{FontSize: 12, Color: theme.Muted, LineHeight: 17}),
		),
		control,
	)
}
