package components

import (
	"github.com/xmhrnx/giode"

	"orchard/ui/theme"
)

// Section is a framed block with an overline title and optional
// trailing controls.
func Section(title string, trailing giode.Element, body ...giode.Element) giode.Element {
	head := []giode.Element{giode.Box(giode.Styles{FlexGrow: 1}, giode.Text(title, theme.Overline))}
	if trailing != nil {
		head = append(head, trailing)
	}
	kids := append([]giode.Element{
		giode.Row(giode.Styles{Width: -1, Align: giode.AlignCenter, Wrap: giode.WrapNoWrap}, head...),
	}, body...)
	return Frame(giode.Styles{Width: -1, Gap: 12, Padding: giode.UniformInset(18)}, kids...)
}
