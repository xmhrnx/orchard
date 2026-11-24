package components

import (
	"fmt"
	"time"
)

// Duration formats a span briefly: "0.8s", "42s", "3m 05s", "1h 12m".
func Duration(d time.Duration) string {
	switch {
	case d <= 0:
		return ""
	case d < 10*time.Second:
		return fmt.Sprintf("%.1fs", d.Seconds())
	case d < time.Minute:
		return fmt.Sprintf("%ds", int(d.Seconds()))
	case d < time.Hour:
		return fmt.Sprintf("%dm %02ds", int(d.Minutes()), int(d.Seconds())%60)
	}
	return fmt.Sprintf("%dh %02dm", int(d.Hours()), int(d.Minutes())%60)
}
