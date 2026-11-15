package components

import (
	"fmt"

	"github.com/xmhrnx/giode"

	"orchard/internal/core"
	"orchard/ui/theme"
)

// RepoStatus is the row of pills summarising a repository: incoming and
// outgoing commits, local changes, stashes, or that all is well.
func RepoStatus(r *core.RepoState) giode.Element {
	st := r.Status
	var pills []giode.Element
	switch {
	case r.Busy != "":
		pills = append(pills, giode.Row(giode.Styles{Gap: 6, Align: giode.AlignCenter},
			giode.Spinner(giode.Styles{Width: 14, Height: 14, Color: theme.Primary}),
			giode.Text(r.Busy+"…", giode.Styles{FontSize: 12, Color: theme.Primary})))
	case !r.Loaded:
		pills = append(pills, giode.Text("Reading…", theme.Caption))
	case r.Err != nil:
		pills = append(pills, Pill("alert", "Needs attention", theme.Danger, theme.DangerSoft))
	}
	if r.Loaded && r.Busy == "" {
		if st.Behind > 0 {
			pills = append(pills, Pill("arrow-down", fmt.Sprint(st.Behind), theme.Info, theme.InfoSoft))
		}
		if st.Ahead > 0 {
			pills = append(pills, Pill("arrow-up", fmt.Sprint(st.Ahead), theme.Accent, theme.AccentSoft))
		}
		if n := st.ChangeCount(); n > 0 {
			pills = append(pills, Pill("file", fmt.Sprintf("%d changed", n), theme.Warning, theme.AccentSoft))
		}
		if st.Stashes > 0 {
			pills = append(pills, Pill("stash", fmt.Sprint(st.Stashes), theme.Muted, theme.Hover))
		}
		if st.Upstream == "" && !st.Detached && r.Err == nil {
			pills = append(pills, Pill("link", "no upstream", theme.Muted, theme.Hover))
		}
		if len(pills) == 0 {
			pills = append(pills, Pill("check", "Up to date", theme.Primary, theme.PrimarySoft))
		}
	}
	return giode.Row(giode.Styles{Gap: 6, Align: giode.AlignCenter, Wrap: giode.WrapNoWrap}, pills...)
}
