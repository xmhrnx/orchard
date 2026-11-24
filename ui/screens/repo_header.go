package screens

import (
	"github.com/xmhrnx/giode"

	"orchard/internal/core"
	"orchard/ui/components"
	"orchard/ui/env"
	"orchard/ui/theme"
)

// RepoHeader is the top of a repository page: identity, branch, status
// and the everyday sync actions.
type RepoHeader struct {
	env      *env.Env
	repoID   func() string
	back     *components.IconAction
	path     *giode.PressableElement
	fetch    *giode.ButtonElement
	pull     *giode.ButtonElement
	push     *giode.ButtonElement
	project  *ProjectBar
	reveal   *components.IconAction
	terminal *components.IconAction
	more     *giode.MenuElement
	moreBtn  giode.Element
}

// NewRepoHeader creates the header for whichever repository repoID
// names; showConsole reopens the task console. Create it once.
func NewRepoHeader(e *env.Env, repoID func() string, project *ProjectBar, showConsole func()) *RepoHeader {
	h := &RepoHeader{env: e, repoID: repoID, project: project}
	run := func(op core.Operation) func() {
		return func() { e.WS.Run(op, []string{repoID()}, e.Report) }
	}
	h.back = components.NewIconAction("chevron-left", "Back to repositories", func() { e.Nav.Open(env.PageRepos) })
	h.fetch = components.Button("Fetch", "fetch", components.Secondary, run(core.OpFetch))
	h.pull = components.Button("Pull", "pull", components.Primary, run(core.OpPull))
	h.push = components.Button("Push", "push", components.Secondary, run(core.OpPush))
	h.reveal = components.NewIconAction("folder", "Show in file manager", func() { e.Reveal(h.repo().Path) })
	h.terminal = components.NewIconAction("terminal", "Open terminal here", func() { e.Terminal(h.repo().Path) })
	h.more = giode.Menu().
		IconItem(components.GlyphSource("copy"), "Copy path", func() { e.Copy("Path", h.repo().Path) }).
		Separator().
		IconItem(components.GlyphSource("command"), "Show task console", showConsole)
	h.moreBtn = h.more.Trigger(components.NewIconAction("dots", "More", func() {}))
	h.path = giode.Pressable(func(s giode.PressState) giode.Element {
		st := theme.Ellipsis(theme.CodeDim)
		if s.Hovered {
			st.Color = theme.Primary
		}
		return giode.Row(giode.Styles{Gap: 6, Align: giode.AlignCenter, Wrap: giode.WrapNoWrap},
			giode.Text(components.ShortPath(h.repo().Path), st), components.Glyph("copy", 12, st.Color))
	}).Label("Copy path").OnClick(func(giode.ClickEvent) { e.Copy("Path", h.repo().Path) })
	return h
}

func (h *RepoHeader) repo() *core.RepoState {
	if r := h.env.WS.Repo(h.repoID()); r != nil {
		return r
	}
	return &core.RepoState{}
}

// Layout draws the header.
func (h *RepoHeader) Layout(gtx giode.Context) giode.Dimensions {
	r := h.repo()
	busy := r.Busy != ""
	for _, b := range []*giode.ButtonElement{h.fetch, h.pull, h.push} {
		b.Disabled(busy)
	}
	h.pull.Loading(r.Busy == core.OpPull.Label())
	h.fetch.Loading(r.Busy == core.OpFetch.Label())
	h.push.Loading(r.Busy == core.OpPush.Label())
	return giode.Row(giode.Styles{Width: -1, Gap: 14, Align: giode.AlignCenter, Wrap: giode.WrapNoWrap},
		h.back,
		giode.Box(giode.Styles{Padding: giode.UniformInset(11), BorderRadius: 14, BackgroundGradient: giode.Linear(135, theme.PrimarySoft, theme.AccentSoft)},
			components.Glyph("repo", 22, theme.Primary)),
		giode.Stack(giode.Styles{FlexGrow: 1, Gap: 6, MinWidth: 160},
			giode.Row(giode.Styles{Gap: 10, Align: giode.AlignCenter, Wrap: giode.WrapNoWrap},
				giode.Text(r.Name, theme.Ellipsis(theme.Display)),
				components.BranchLabel(r.Status),
				components.RepoStatus(r),
			),
			giode.Row(giode.Styles{Gap: 14, Align: giode.AlignCenter, Wrap: giode.WrapNoWrap}, h.path, h.project),
		),
		h.reveal, h.terminal, h.fetch, h.pull, h.push, h.moreBtn, h.more,
	).Layout(gtx)
}
