package components

import (
	"github.com/xmhrnx/giode"

	"orchard/ui/theme"
)

// Kbd shows a keyboard shortcut written the platform's way.
func Kbd(spec string) giode.Element {
	return giode.Box(giode.Styles{
		Padding: giode.SymmetricInset(6, 2), BorderRadius: theme.RadiusSmall,
		BorderWidth: 1, BorderColor: theme.Hairline, Background: theme.Sunken,
	}, giode.Text(giode.FormatShortcut(spec), giode.Styles{FontSize: 11, FontFamily: theme.Mono, Color: theme.Muted}))
}
