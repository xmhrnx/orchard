package components

import (
	"github.com/xmhrnx/giode"

	"orchard/internal/gitx"
	"orchard/ui/theme"
)

// RefBadge labels a commit with a branch, remote branch, tag or HEAD.
func RefBadge(ref gitx.Ref) giode.Element {
	switch ref.Kind {
	case gitx.RefTag:
		return Pill("tag", ref.Name, theme.Accent, theme.AccentSoft)
	case gitx.RefHead:
		return Pill("", "HEAD", theme.OnPrimary, theme.Primary)
	case gitx.RefRemote:
		return Pill("globe", ref.Name, theme.Info, theme.InfoSoft)
	}
	return Pill("branch", ref.Name, theme.Primary, theme.PrimarySoft)
}
