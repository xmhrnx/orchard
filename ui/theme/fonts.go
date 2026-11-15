package theme

import (
	"io/fs"

	"github.com/xmhrnx/giode"

	"orchard/assets"
)

// Font families registered by LoadFonts.
const (
	Sans = "Inter"
	Mono = "JetBrains Mono"
)

// LoadFonts registers the bundled Inter and JetBrains Mono faces.
func LoadFonts() error {
	files := []struct {
		name   string
		family string
		weight giode.FontWeight
	}{
		{"Inter-Regular.ttf", Sans, giode.FontWeightNormal},
		{"Inter-Medium.ttf", Sans, giode.FontWeightMedium},
		{"Inter-SemiBold.ttf", Sans, giode.FontWeightSemiBold},
		{"Inter-Bold.ttf", Sans, giode.FontWeightBold},
		{"JetBrainsMono-Regular.ttf", Mono, giode.FontWeightNormal},
		{"JetBrainsMono-Medium.ttf", Mono, giode.FontWeightMedium},
	}
	var faces []giode.FontFile
	for _, f := range files {
		data, err := fs.ReadFile(assets.Fonts, "fonts/"+f.name)
		if err != nil {
			return err
		}
		faces = append(faces, giode.FontFile{Family: f.family, Weight: f.weight, Data: data})
	}
	return giode.RegisterFont(faces...)
}

// LoadIcons registers the UI glyphs by file name ("icons/pull.svg" is
// "pull").
func LoadIcons() error {
	return giode.RegisterIconsFS(assets.Icons, "icons")
}
