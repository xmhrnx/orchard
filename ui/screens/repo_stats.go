package screens

import (
	"fmt"

	"github.com/xmhrnx/giode"

	"orchard/internal/core"
	"orchard/ui/components"
	"orchard/ui/theme"
)

// RepoStats is the row of headline tiles over the repository list.
func RepoStats(repos []*core.RepoState) giode.Element {
	var behind, dirty, ahead int
	for _, r := range repos {
		if r.Status.Behind > 0 {
			behind++
		}
		if r.Status.Dirty() {
			dirty++
		}
		if r.Status.Ahead > 0 {
			ahead++
		}
	}
	return giode.Row(giode.Styles{Width: -1, Gap: 14, Wrap: giode.WrapWrap},
		components.StatTile("grid", "Repositories", fmt.Sprint(len(repos)), theme.Primary, theme.PrimarySoft),
		components.StatTile("arrow-down", "Need updating", fmt.Sprint(behind), theme.Info, theme.InfoSoft),
		components.StatTile("file", "Local changes", fmt.Sprint(dirty), theme.Warning, theme.AccentSoft),
		components.StatTile("arrow-up", "Unpushed", fmt.Sprint(ahead), theme.Accent, theme.AccentSoft),
	)
}
