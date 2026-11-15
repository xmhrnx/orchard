package components

import (
	"github.com/xmhrnx/giode"

	"orchard/ui/theme"
)

// PageHeader is the title block at the top of a page, with actions on
// the right.
func PageHeader(title, subtitle string, actions ...giode.Element) giode.Element {
	return giode.Row(giode.Styles{Width: -1, Align: giode.AlignCenter, Gap: 12, Wrap: giode.WrapNoWrap},
		giode.Stack(giode.Styles{FlexGrow: 1, Gap: 4},
			giode.Text(title, theme.Display),
			giode.Text(subtitle, theme.Ellipsis(theme.Caption)),
		),
		giode.Row(giode.Styles{Gap: 8, Align: giode.AlignCenter, Wrap: giode.WrapNoWrap}, actions...),
	)
}
