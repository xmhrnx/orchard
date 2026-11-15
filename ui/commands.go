package ui

import (
	"orchard/internal/core"
	"orchard/ui/env"
	"orchard/ui/shell"
)

// commands lists what the palette offers: global actions first, then
// every repository to open or update.
func (r *Root) commands() []shell.Command {
	e := r.env
	var ids []string
	for _, repo := range e.WS.Repos() {
		ids = append(ids, repo.ID)
	}
	cmds := []shell.Command{
		{Icon: "pull", Title: "Update all repositories", Hint: "Mod+Shift+U", Run: func() { e.WS.Run(core.OpPull, ids, e.Report) }},
		{Icon: "fetch", Title: "Fetch all repositories", Hint: "Mod+Shift+F", Run: func() { e.WS.Run(core.OpFetch, ids, e.Report) }},
		{Icon: "grid", Title: "Go to repositories", Hint: "Mod+1", Run: func() { e.Nav.Open(env.PageRepos) }},
		{Icon: "clone", Title: "Clone a repository", Hint: "Mod+2", Run: func() { e.Nav.Open(env.PageClone) }},
		{Icon: "user", Title: "Manage accounts", Hint: "Mod+3", Run: func() { e.Nav.Open(env.PageAccounts) }},
		{Icon: "settings", Title: "Open settings", Hint: "Mod+,", Run: func() { e.Nav.Open(env.PageSettings) }},
		{Icon: "sun", Title: "Toggle light / dark theme", Run: r.toggleTheme},
	}
	for _, repo := range e.WS.Repos() {
		id := repo.ID
		cmds = append(cmds,
			shell.Command{Icon: "repo", Title: repo.Name, Hint: "open", Run: func() { e.Nav.OpenRepo(id) }},
			shell.Command{Icon: "pull", Title: "Pull " + repo.Name, Hint: repo.Status.Branch, Run: func() { e.WS.Run(core.OpPull, []string{id}, e.Report) }},
		)
	}
	return cmds
}
