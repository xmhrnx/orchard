// Package assets embeds Orchard's fonts, glyphs and brand images.
package assets

import "embed"

// Icon is the app icon as a PNG.
//
//go:embed icon.png
var Icon []byte

// Logo is the small app mark shown in the sidebar.
//
//go:embed brand/icon-256.png
var Logo []byte

// Icons holds the UI glyphs as SVG files under "icons".
//
//go:embed icons
var Icons embed.FS

// Fonts holds the bundled typefaces under "fonts".
//
//go:embed fonts/*.ttf
var Fonts embed.FS
