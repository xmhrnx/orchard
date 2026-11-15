package components

import "github.com/xmhrnx/giode"

// FilledFrame is a Frame whose surface always fills the space its parent
// gives it, for frames sized by FlexGrow beside other content.
func FilledFrame(st giode.Styles, kids ...giode.Element) giode.Element {
	return frameBox(st, true, kids)
}
