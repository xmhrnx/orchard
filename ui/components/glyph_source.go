package components

import "github.com/xmhrnx/giode"

// GlyphSource returns a registered glyph as an icon source, for menus.
// Glyphs must be registered (theme.LoadIcons) before menus are built.
func GlyphSource(name string) giode.IconSource {
	if src, ok := giode.LookupIcon(name); ok {
		return src
	}
	return giode.IconPath{}
}
