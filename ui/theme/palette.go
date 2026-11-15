// Package theme holds Orchard's look: palettes, fonts and the shared
// style values the components build on.
package theme

import "github.com/xmhrnx/giode"

// darkPalette is "night orchard": deep forest surfaces, jade growth
// and amber fruit.
var darkPalette = giode.Palette{
	Primary:    "#1F9D74",
	Secondary:  "#23B183",
	Background: "#080C0B",
	Surface:    "#111816",
	Text:       "#E7EFEC",
	Border:     "#22302B",
	OnPrimary:  "#FFFFFF",
	Muted:      "#8A9C95",
	Accent:     "#F5B54A",
	Success:    "#23B183",
	Warning:    "#F5B54A",
	Danger:     "#F87171",
	Info:       "#60A5FA",
}

// lightPalette is "morning orchard": parchment, deep leaves and honey.
var lightPalette = giode.Palette{
	Primary:    "#0E9F6E",
	Secondary:  "#0F766E",
	Background: "#F4F2EC",
	Surface:    "#FFFFFF",
	Text:       "#14201B",
	Border:     "#DCD8CC",
	OnPrimary:  "#FFFFFF",
	Muted:      "#66766F",
	Accent:     "#C77A0A",
	Success:    "#0E9F6E",
	Warning:    "#C77A0A",
	Danger:     "#DC2626",
	Info:       "#2563EB",
}

// Install registers Orchard's palettes for both modes.
func Install() {
	giode.SetPalette(giode.ModeDark, darkPalette)
	giode.SetPalette(giode.ModeLight, lightPalette)
}

// SetDark switches between the dark and light palettes.
func SetDark(dark bool) {
	if dark {
		giode.SetMode(giode.ModeDark)
	} else {
		giode.SetMode(giode.ModeLight)
	}
}

// IsDark reports whether the dark palette is active.
func IsDark() bool { return giode.CurrentMode() == giode.ModeDark }
