package screens

import (
	"context"

	"orchard/internal/core"
	"orchard/ui/components"
	"orchard/ui/env"
)

// rewriteContext is what the history-rewriting cards share: the open
// repository, one confirmation dialog and a way to run a rewrite.
type rewriteContext struct {
	env     *env.Env
	repo    *core.RepoState
	confirm *components.ConfirmDialog
}

// guarded asks before running a rewrite, then runs it and reports.
func (c *rewriteContext) guarded(title, body, action, busy, success string, fn func(ctx context.Context, path string) error) {
	c.confirm.Ask(title, body, action, true, func() {
		c.env.WS.Do(c.repo.ID, busy, fn, func(err error) { c.env.Done(success, err) })
	})
}

// rewriteWarning is appended to every rewrite confirmation.
const rewriteWarning = " Commit hashes change from the first affected commit onwards; anyone else with a copy will need to re-clone or reset. Make sure you have a backup."
