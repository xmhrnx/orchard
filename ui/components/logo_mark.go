package components

import (
	"bytes"
	"image"
	"image/png"
	"sync"

	"github.com/xmhrnx/giode"
	"github.com/xmhrnx/giode/styles/properties"

	"orchard/assets"
	"orchard/ui/theme"
)

var (
	logoOnce sync.Once
	logoImg  image.Image
)

// LogoMark is the app icon at size dp: the git-branch tree bearing
// commit fruit.
func LogoMark(size int) giode.Element {
	logoOnce.Do(func() {
		if img, err := png.Decode(bytes.NewReader(assets.Logo)); err == nil {
			logoImg = img
		}
	})
	if logoImg == nil {
		return Glyph("branch", size, theme.Primary)
	}
	return giode.Image(logoImg, giode.Styles{Width: properties.Width(size), Height: properties.Height(size)})
}
