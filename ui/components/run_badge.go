package components

import (
	"github.com/xmhrnx/giode"

	"orchard/internal/forge"
	"orchard/ui/theme"
)

// RunBadge is a CI run's state as a tinted round glyph; a running run
// spins.
func RunBadge(s forge.RunState) giode.Element {
	icon, fg := "circle", theme.Muted
	switch s {
	case forge.RunRunning:
		return giode.Box(giode.Styles{Padding: giode.UniformInset(6), BorderRadius: theme.RadiusPill, Background: theme.InfoSoft},
			giode.Spinner(giode.Styles{Width: 16, Height: 16, Color: theme.Info}))
	case forge.RunQueued:
		icon, fg = "history", theme.Muted
	case forge.RunWaiting:
		icon, fg = "alert", theme.Warning
	case forge.RunSucceeded:
		icon, fg = "check", "theme.success"
	case forge.RunFailed:
		icon, fg = "x", theme.Danger
	case forge.RunCancelled:
		icon, fg = "stop", theme.Muted
	case forge.RunSkipped:
		icon, fg = "minus", theme.Muted
	}
	return giode.Box(giode.Styles{Padding: giode.UniformInset(6), BorderRadius: theme.RadiusPill, Background: giode.Mix(string(fg), "theme.surface", 0.14)},
		Glyph(icon, 16, fg))
}
