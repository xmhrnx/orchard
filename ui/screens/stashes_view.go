package screens

import (
	"context"

	"github.com/xmhrnx/giode"
	"github.com/xmhrnx/giode/components/scroll"

	"orchard/internal/core"
	"orchard/internal/gitx"
	"orchard/ui/components"
	"orchard/ui/env"
	"orchard/ui/theme"
)

// StashesView manages the stash stack: shelve work, bring it back, or
// throw it away.
type StashesView struct {
	env     *env.Env
	repo    *core.RepoState
	stashes []gitx.Stash
	list    *scroll.Scroll
	save    *giode.ButtonElement
	prompt  *components.PromptDialog
	confirm *components.ConfirmDialog
}

// NewStashesView creates the view. Create it once.
func NewStashesView(e *env.Env) *StashesView {
	v := &StashesView{env: e, prompt: components.NewPromptDialog(), confirm: components.NewConfirmDialog()}
	v.list = giode.Scroll(v.rows, giode.Styles{ScrollBar: giode.ScrollBar{Width: 6, AutoHide: true, TrackColor: "transparent"}})
	v.save = components.Button("Stash changes", "stash", components.Primary, v.stash)
	return v
}

// Open loads a repository's stashes.
func (v *StashesView) Open(r *core.RepoState) {
	v.repo = r
	v.load()
}

func (v *StashesView) load() {
	path := v.repo.Path
	core.Query(v.env.WS, func(ctx context.Context) ([]gitx.Stash, error) {
		return gitx.Stashes(ctx, path)
	}, func(list []gitx.Stash, err error) {
		v.env.Fail("Loading stashes", err)
		v.stashes = list
	})
}

func (v *StashesView) act(busy, success string, fn func(ctx context.Context, path string) error) {
	v.env.WS.Do(v.repo.ID, busy, fn, func(err error) {
		v.env.Done(success, err)
		v.load()
	})
}

func (v *StashesView) stash() {
	v.prompt.Ask("Stash changes", "A note to recognise it later (optional words help)", "half-done refactor", "WIP", "Stash", func(msg string) {
		v.act("Stashing", "Changes stashed", func(ctx context.Context, path string) error { return gitx.SaveStash(ctx, path, msg) })
	})
}

func (v *StashesView) row(s gitx.Stash) giode.Element {
	return giode.Row(giode.Styles{
		Width: -1, Gap: 12, Align: giode.AlignCenter, Wrap: giode.WrapNoWrap,
		Padding: giode.SymmetricInset(12, 10), BorderRadius: theme.Radius, Background: theme.Raised,
	},
		giode.Box(giode.Styles{Padding: giode.UniformInset(8), BorderRadius: 9, Background: theme.AccentSoft}, components.Glyph("stash", 16, theme.Accent)),
		giode.Stack(giode.Styles{FlexGrow: 1, Gap: 3},
			giode.Text(s.Message, theme.Ellipsis(theme.Strong)),
			giode.Text(s.Ref+" · "+components.Ago(s.When), theme.CodeDim),
		),
		components.RowMenu("stash", s.Ref+s.Message,
			components.MenuItem{Icon: "undo", Label: "Apply and drop (pop)", Run: func() {
				v.act("Restoring", "Stash restored", func(ctx context.Context, path string) error { return gitx.ApplyStash(ctx, path, s.Index, true) })
			}},
			components.MenuItem{Icon: "copy", Label: "Apply and keep", Run: func() {
				v.act("Applying", "Stash applied", func(ctx context.Context, path string) error { return gitx.ApplyStash(ctx, path, s.Index, false) })
			}},
			components.MenuSeparator,
			components.MenuItem{Icon: "trash", Label: "Drop", Run: func() {
				v.confirm.Ask("Drop this stash?", "\""+s.Message+"\" will be deleted.", "Drop", true, func() {
					v.act("Dropping", "Stash dropped", func(ctx context.Context, path string) error { return gitx.DropStash(ctx, path, s.Index) })
				})
			}},
		),
	)
}

func (v *StashesView) rows() giode.Element {
	rows := make([]giode.Element, len(v.stashes))
	for i, s := range v.stashes {
		rows[i] = v.row(s)
	}
	return giode.Stack(giode.Styles{Width: -1, Gap: 8}, rows...)
}

// Layout draws the view.
func (v *StashesView) Layout(gtx giode.Context) giode.Dimensions {
	body := giode.Element(giode.Box(giode.Styles{Width: -1, Height: -1, Padding: giode.UniformInset(12)}, v.list))
	if len(v.stashes) == 0 {
		body = components.EmptyState("stash", "No stashes", "Stash shelves your uncommitted work so you can switch tasks with a clean tree.", nil)
	}
	return giode.Stack(giode.Styles{Width: -1, Height: -1, Gap: 12},
		giode.Row(giode.Styles{Width: -1, Justify: giode.JustifyEnd}, v.save),
		components.ListFrame(body),
		v.prompt, v.confirm,
	).Layout(gtx)
}
