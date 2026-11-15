package screens

import (
	"context"
	"strings"

	"github.com/xmhrnx/giode"
	"github.com/xmhrnx/giode/components/scroll"

	"orchard/internal/core"
	"orchard/internal/gitx"
	"orchard/ui/components"
	"orchard/ui/env"
	"orchard/ui/theme"
)

// commitDetail is what the panel loads for a commit.
type commitDetail struct {
	body string
	diff []gitx.DiffLine
}

// CommitPanel shows one commit beside the history: message, author,
// hash, the patch, and what can be done with it.
type CommitPanel struct {
	env      *env.Env
	repo     *core.RepoState
	commit   gitx.Commit
	body     string
	diff     *components.DiffView
	changed  func()
	menu     *giode.MenuElement
	actions  giode.Element
	close    *components.IconAction
	copyHash *components.IconAction
	prompt   *components.PromptDialog
	confirm  *components.ConfirmDialog
	message  *scroll.Scroll
}

// NewCommitPanel creates the panel; changed runs after an action alters
// the history and closed when the user closes the panel. Create it once.
func NewCommitPanel(e *env.Env, changed, closed func()) *CommitPanel {
	p := &CommitPanel{env: e, changed: changed, diff: components.NewDiffView()}
	p.prompt = components.NewPromptDialog()
	p.confirm = components.NewConfirmDialog()
	p.close = components.NewIconAction("x", "Close", closed)
	p.copyHash = components.NewIconAction("copy", "Copy hash", func() { e.Copy("Hash", p.commit.Hash) })
	p.message = giode.Scroll(func() giode.Element {
		return giode.Text(p.body, giode.Styles{FontSize: 13, Color: theme.Text, LineHeight: 19}).Selectable("commit-body")
	}).FitContent(true)
	p.menu = giode.Menu().
		IconItem(components.GlyphSource("branch"), "New branch here…", p.branchHere).
		IconItem(components.GlyphSource("tag"), "Tag this commit…", p.tagHere).
		IconItem(components.GlyphSource("commit"), "Check out (detached)", p.checkout).
		Separator().
		IconItem(components.GlyphSource("merge"), "Cherry-pick onto current branch", p.cherryPick).
		IconItem(components.GlyphSource("undo"), "Revert this commit", p.revert).
		Separator().
		IconItem(components.GlyphSource("history"), "Reset branch here (keep changes)", func() { p.reset(gitx.ResetMixed) }).
		IconItem(components.GlyphSource("alert"), "Reset branch here (discard changes)", func() { p.reset(gitx.ResetHard) })
	p.actions = p.menu.Trigger(components.Button("Actions", "sparkle", components.Secondary, func() {}))
	return p
}

// Show loads a commit into the panel.
func (p *CommitPanel) Show(r *core.RepoState, c gitx.Commit) {
	p.repo, p.commit, p.body = r, c, c.Subject
	p.diff.SetLines(nil)
	path, hash := r.Path, c.Hash
	core.Query(p.env.WS, func(ctx context.Context) (commitDetail, error) {
		body, err := gitx.CommitBody(ctx, path, hash)
		if err != nil {
			return commitDetail{}, err
		}
		diff, err := gitx.CommitDiff(ctx, path, hash)
		return commitDetail{body: body, diff: diff}, err
	}, func(d commitDetail, err error) {
		if p.commit.Hash != hash {
			return
		}
		if err != nil {
			p.env.Fail("Loading commit", err)
			return
		}
		p.body = d.body
		p.diff.SetLines(d.diff)
	})
}

// Clear forgets the commit shown.
func (p *CommitPanel) Clear() {
	p.commit = gitx.Commit{}
	p.diff.SetLines(nil)
}

// act runs a git action on the commit and reports the outcome.
func (p *CommitPanel) act(busy, success string, fn func(ctx context.Context, path string) error) {
	p.env.WS.Do(p.repo.ID, busy, fn, func(err error) {
		p.env.Done(success, err)
		if err == nil {
			p.changed()
		}
	})
}

func (p *CommitPanel) branchHere() {
	hash := p.commit.Hash
	p.prompt.Ask("New branch at "+p.commit.Short(), "Branch name", "feature/…", "", "Create & switch", func(name string) {
		p.act("Branching", "Switched to "+name, func(ctx context.Context, path string) error {
			return gitx.CreateBranch(ctx, path, name, hash, true)
		})
	})
}

func (p *CommitPanel) tagHere() {
	hash := p.commit.Hash
	p.prompt.Ask("Tag "+p.commit.Short(), "Tag name", "v1.2.3", "", "Create tag", func(name string) {
		p.act("Tagging", "Tagged "+name, func(ctx context.Context, path string) error {
			return gitx.CreateTag(ctx, path, name, hash, "")
		})
	})
}

func (p *CommitPanel) checkout() {
	hash := p.commit.Hash
	p.act("Checking out", "Checked out "+p.commit.Short(), func(ctx context.Context, path string) error {
		return gitx.Checkout(ctx, path, gitx.Branch{Name: hash})
	})
}

func (p *CommitPanel) cherryPick() {
	hash := p.commit.Hash
	p.act("Cherry-picking", "Cherry-picked "+p.commit.Short(), func(ctx context.Context, path string) error {
		return gitx.CherryPick(ctx, path, hash)
	})
}

func (p *CommitPanel) revert() {
	hash := p.commit.Hash
	p.act("Reverting", "Reverted "+p.commit.Short(), func(ctx context.Context, path string) error {
		return gitx.Revert(ctx, path, hash)
	})
}

func (p *CommitPanel) reset(mode gitx.ResetMode) {
	hash, short := p.commit.Hash, p.commit.Short()
	body := "The current branch will point at " + short + ". Later commits stay in the reflog; your work-tree changes are kept."
	danger := mode == gitx.ResetHard
	if danger {
		body = "The current branch will point at " + short + " and every uncommitted change will be thrown away. This can't be undone."
	}
	p.confirm.Ask("Reset to "+short+"?", body, "Reset", danger, func() {
		p.act("Resetting", "Reset to "+short, func(ctx context.Context, path string) error {
			return gitx.Reset(ctx, path, hash, mode)
		})
	})
}

func (p *CommitPanel) meta() giode.Element {
	c := p.commit
	parents := make([]string, len(c.Parents))
	for i, h := range c.Parents {
		parents[i] = h[:min(8, len(h))]
	}
	line := func(icon, text string) giode.Element {
		return giode.Row(giode.Styles{Gap: 8, Align: giode.AlignCenter, Wrap: giode.WrapNoWrap},
			components.Glyph(icon, 13, theme.Muted), giode.Text(text, theme.Ellipsis(theme.Caption)))
	}
	return giode.Stack(giode.Styles{Gap: 6},
		line("user", c.Author+" <"+c.Email+">"),
		line("history", c.When.Format("Mon 2 Jan 2006, 15:04")+" · "+components.Ago(c.When)),
		giode.Row(giode.Styles{Gap: 8, Align: giode.AlignCenter, Wrap: giode.WrapNoWrap},
			components.Glyph("commit", 13, theme.Muted),
			giode.Text(c.Hash, theme.Ellipsis(theme.CodeDim)), p.copyHash),
		line("merge", "parents "+strings.Join(parents, ", ")),
	)
}

// Layout draws the panel.
func (p *CommitPanel) Layout(gtx giode.Context) giode.Dimensions {
	diff := giode.Element(giode.Box(giode.Styles{Width: -1, FlexGrow: 1, Padding: giode.UniformInset(24), Justify: giode.JustifyCenter},
		giode.Spinner(giode.Styles{Width: 22, Height: 22, Color: theme.Primary})))
	if !p.diff.Empty() {
		diff = giode.Box(giode.Styles{Width: -1, FlexGrow: 1, BorderRadius: theme.Radius, Background: theme.Sunken, Overflow: giode.OverflowHidden}, p.diff)
	}
	return components.Frame(giode.Styles{Width: 460, Height: -1, Gap: 14, Padding: giode.UniformInset(18)},
		giode.Row(giode.Styles{Width: -1, Gap: 8, Align: giode.AlignCenter, Wrap: giode.WrapNoWrap},
			giode.Box(giode.Styles{FlexGrow: 1}, giode.Text("Commit "+p.commit.Short(), theme.Overline)),
			p.actions, p.close),
		giode.Box(giode.Styles{Width: -1, MaxHeight: 140}, p.message),
		p.meta(),
		diff,
		p.menu, p.prompt, p.confirm,
	).Layout(gtx)
}
