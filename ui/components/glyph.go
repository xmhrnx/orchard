// Package components is Orchard's UI kit: one component per file, built
// on Giode elements and the theme.
package components

import (
	"github.com/xmhrnx/giode"
	"github.com/xmhrnx/giode/styles/properties"
)

// Glyph draws a registered UI icon at size dp in a colour; "" inherits
// the surrounding text colour.
func Glyph(name string, size int, color giode.Color) giode.Element {
	return giode.Icon(name, giode.Styles{Width: properties.Width(size), Height: properties.Height(size), Color: color})
}
