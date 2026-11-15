package screens

import (
	"github.com/xmhrnx/giode"
	"github.com/xmhrnx/giode/components/scroll"

	"orchard/internal/core"
	"orchard/internal/gitx"
	"orchard/ui/components"
	"orchard/ui/env"
	"orchard/ui/theme"
)

// RewriteView gathers what plain git makes hard: erasing files and
// secrets from history, fixing identities, tidying commits and
// publishing the result.
type RewriteView struct {
	ctx     *rewriteContext
	purge   *PurgePathsCard
	scrub   *ScrubSecretsCard
	author  *AuthorCard
	strip   *StripBlobsCard
	squash  *SquashCard
	publish *PublishCard
	scroll  *scroll.Scroll
}

// NewRewriteView creates the view. Create it once.
func NewRewriteView(e *env.Env) *RewriteView {
	c := &rewriteContext{env: e, confirm: components.NewConfirmDialog()}
	v := &RewriteView{ctx: c}
	v.purge = NewPurgePathsCard(c)
	v.scrub = NewScrubSecretsCard(c)
	v.author = NewAuthorCard(c)
	v.strip = NewStripBlobsCard(c)
	v.squash = NewSquashCard(c)
	v.publish = NewPublishCard(c)
	v.scroll = giode.Scroll(v.cards, giode.Styles{ScrollBar: giode.ScrollBar{Width: 6, AutoHide: true, TrackColor: "transparent"}})
	return v
}

// Open points the tools at a repository.
func (v *RewriteView) Open(r *core.RepoState) {
	if v.ctx.repo == nil || v.ctx.repo.ID != r.ID {
		v.purge.Reset()
		v.scrub.Reset()
		v.author.Reset()
	}
	v.ctx.repo = r
}

func (v *RewriteView) banner() giode.Element {
	tool := components.Pill("check", "git-filter-repo found: fast, thorough rewrites", theme.Primary, theme.PrimarySoft)
	if !gitx.HasFilterRepo() {
		tool = components.Pill("alert", "git-filter-repo not found: using filter-branch; some tools are off", theme.Warning, theme.AccentSoft)
	}
	return giode.Row(giode.Styles{
		Width: -1, Gap: 12, Align: giode.AlignCenter, Wrap: giode.WrapWrap,
		Padding: giode.UniformInset(14), BorderRadius: 14,
		Background: theme.DangerSoft, BorderWidth: 1, BorderColor: giode.Mix("theme.danger", "theme.surface", 0.35),
	},
		components.Glyph("alert", 18, theme.Danger),
		giode.Box(giode.Styles{FlexGrow: 1, MinWidth: 280}, giode.Text("These tools rewrite history. They are permanent and change commit hashes; back up first and coordinate with collaborators.",
			giode.Styles{FontSize: 13, Color: theme.Text, LineHeight: 19})),
		tool,
	)
}

func (v *RewriteView) cards() giode.Element {
	return giode.Stack(giode.Styles{Width: -1, Gap: 14, Padding: giode.Inset{Right: 10, Bottom: 24}},
		v.banner(), v.purge, v.scrub, v.author, v.strip, v.squash, v.publish,
	)
}

// Layout draws the view.
func (v *RewriteView) Layout(gtx giode.Context) giode.Dimensions {
	return giode.Box(giode.Styles{Width: -1, Height: -1}, v.scroll, v.ctx.confirm).Layout(gtx)
}
