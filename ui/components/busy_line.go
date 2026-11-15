package components

import (
	"math"
	"time"

	"gioui.org/op"
	"github.com/xmhrnx/giode"

	"orchard/ui/theme"
)

// BusyLine is a thin indeterminate progress bar: a glowing emerald
// streak sweeping across, shown while background work runs.
func BusyLine(active bool) giode.Element {
	if !active {
		return giode.Box(giode.Styles{Width: -1, Height: 2})
	}
	return giode.Canvas(giode.Styles{Width: -1, Height: 2}, func(p *giode.Painter) {
		w := p.Width()
		t := float64(time.Now().UnixMilli()%1600) / 1600
		x := float32(easeInOut(t)) * (w + 160)
		p.Rect(x-160, 0, 160, 2, 1, giode.Mix("theme.primary", "theme.background", 0.45))
		p.Rect(x-60, 0, 60, 2, 1, theme.Primary)
		p.Context().Execute(op.InvalidateCmd{})
	})
}

func easeInOut(t float64) float64 { return 0.5 - math.Cos(t*math.Pi)/2 }
