package ui

import (
	"github.com/xmhrnx/giode"

	"orchard/internal/core"
	"orchard/ui/env"
)

// shortcuts registers the app-wide keys for this frame.
func (r *Root) shortcuts() {
	e := r.env
	all := func() []string {
		var ids []string
		for _, repo := range e.WS.Repos() {
			ids = append(ids, repo.ID)
		}
		return ids
	}
	giode.Shortcuts(
		giode.Shortcut("Mod+K", r.openPalette),
		giode.Shortcut("Mod+P", r.openPalette),
		giode.Shortcut("Mod+1", func() { e.Nav.Open(env.PageRepos) }),
		giode.Shortcut("Mod+2", func() { e.Nav.Open(env.PageClone) }),
		giode.Shortcut("Mod+3", func() { e.Nav.Open(env.PageAccounts) }),
		giode.Shortcut("Mod+,", func() { e.Nav.Open(env.PageSettings) }),
		giode.Shortcut("Alt+Left", e.Nav.Back),
		giode.Shortcut("F5", func() { e.WS.Refresh() }),
		giode.Shortcut("Mod+Shift+F", func() { e.WS.Run(core.OpFetch, all(), e.Report) }),
		giode.Shortcut("Mod+Shift+U", func() { e.WS.Run(core.OpPull, all(), e.Report) }),
	)
}
