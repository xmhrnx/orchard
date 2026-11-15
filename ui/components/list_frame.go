package components

import "github.com/xmhrnx/giode"

// ListFrame is the frame around a page's main list: it takes the rest
// of the page's height and its surface fills it, however short the
// list is.
func ListFrame(body giode.Element) giode.Element {
	return FilledFrame(giode.Styles{Width: -1, FlexGrow: 1, Padding: giode.UniformInset(0)}, body)
}
