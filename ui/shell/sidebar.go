package shell

import (
	"fmt"

	"github.com/xmhrnx/giode"
	"github.com/xmhrnx/giode/components/scroll"

	"orchard/internal/core"
	"orchard/ui/components"
	"orchard/ui/env"
	"orchard/ui/theme"
)

// Sidebar is the left rail: the brand, the main destinations, pinned
// repositories and a live activity footer.
type Sidebar struct {
	env    *env.Env
	nav    []*components.NavItem
	scroll *scroll.Scroll
}

// NewSidebar creates the sidebar. Create it once.
func NewSidebar(e *env.Env) *Sidebar {
	s := &Sidebar{env: e}
	page := func(p env.Page) func() bool {
		return func() bool { return e.Nav.Current().Page == p }
	}
	s.nav = []*components.NavItem{
		components.NewNavItem("grid", "Repositories", page(env.PageRepos), func() int { return len(e.WS.Repos()) }, func() { e.Nav.Open(env.PageRepos) }),
		components.NewNavItem("clone", "Clone", page(env.PageClone), nil, func() { e.Nav.Open(env.PageClone) }),
		components.NewNavItem("user", "Accounts", page(env.PageAccounts), func() int { return len(e.WS.Accounts()) }, func() { e.Nav.Open(env.PageAccounts) }),
		components.NewNavItem("settings", "Settings", page(env.PageSettings), nil, func() { e.Nav.Open(env.PageSettings) }),
	}
	s.scroll = giode.Scroll(s.pinned, giode.Styles{ScrollBar: giode.ScrollBar{Hidden: true}})
	return s
}

func (s *Sidebar) pinned() giode.Element {
	var rows []giode.Element
	for _, r := range s.env.WS.Repos() {
		if !r.Pinned && !s.recent(r) {
			continue
		}
		rows = append(rows, s.repoLink(r))
		if len(rows) == 12 {
			break
		}
	}
	if len(rows) == 0 {
		rows = append(rows, giode.Box(giode.Styles{Padding: giode.SymmetricInset(14, 6)},
			giode.Text("Pin repositories from their menu to keep them here.", giode.Styles{FontSize: 12, Color: theme.Faint})))
	}
	return giode.Stack(giode.Styles{Width: -1, Gap: 2}, rows...)
}

// recent reports whether a repository is the one open now, so it shows
// in the rail even when not pinned.
func (s *Sidebar) recent(r *core.RepoState) bool {
	cur := s.env.Nav.Current()
	return cur.Page == env.PageRepo && cur.RepoID == r.ID
}

func (s *Sidebar) repoLink(r *core.RepoState) giode.Element {
	id := r.ID
	return giode.Keyed("sidebar-repo", id, func() *giode.PressableElement {
		return giode.Pressable(func(ps giode.PressState) giode.Element {
			repo := s.env.WS.Repo(id)
			if repo == nil {
				return giode.Box(giode.Styles{})
			}
			on := s.recent(repo)
			fg, bg := theme.Muted, giode.Color("transparent")
			if on {
				fg, bg = theme.Text, theme.Selected
			} else if ps.Hovered {
				fg, bg = theme.Text, theme.Hover
			}
			dot := theme.Primary
			switch {
			case repo.Err != nil:
				dot = theme.Danger
			case repo.Status.Behind > 0:
				dot = theme.Info
			case repo.Status.Dirty():
				dot = theme.Warning
			}
			return giode.Row(giode.Styles{
				Width: -1, Gap: 10, Align: giode.AlignCenter, Wrap: giode.WrapNoWrap,
				Padding: giode.SymmetricInset(14, 6), BorderRadius: 8, Background: bg, Transition: theme.HoverFade,
			},
				giode.Box(giode.Styles{Width: 7, Height: 7, BorderRadius: 4, Background: dot}),
				giode.Box(giode.Styles{FlexGrow: 1}, giode.Text(repo.Name, theme.Ellipsis(giode.Styles{FontSize: 13, Color: fg}))),
			)
		}).Label(r.Name).OnClick(func(giode.ClickEvent) { s.env.Nav.OpenRepo(id) })
	})
}

func (s *Sidebar) footer() giode.Element {
	busy := s.env.WS.Busy()
	status := giode.Row(giode.Styles{Gap: 8, Align: giode.AlignCenter},
		giode.Box(giode.Styles{Width: 8, Height: 8, BorderRadius: 4, Background: theme.Primary}),
		giode.Text("All quiet", giode.Styles{FontSize: 12, Color: theme.Muted}),
	)
	if busy > 0 {
		status = giode.Row(giode.Styles{Gap: 8, Align: giode.AlignCenter},
			giode.Spinner(giode.Styles{Width: 12, Height: 12, Color: theme.Primary}),
			giode.Text(fmt.Sprintf("Working on %d…", busy), giode.Styles{FontSize: 12, Color: theme.Primary}),
		)
	}
	return giode.Box(giode.Styles{
		Width: -1, Padding: giode.SymmetricInset(12, 10), BorderRadius: 12,
		Background: theme.Sunken, BorderWidth: 1, BorderColor: theme.Hairline,
	}, status)
}

// Layout draws the sidebar.
func (s *Sidebar) Layout(gtx giode.Context) giode.Dimensions {
	nav := make([]giode.Element, 0, len(s.nav))
	for _, n := range s.nav {
		nav = append(nav, n)
	}
	return giode.Stack(giode.Styles{
		Width: 240, Height: -1, Gap: 6, Padding: giode.Inset{Top: 18, Bottom: 14, Left: 12, Right: 12},
		Background:         giode.Mix("theme.surface", "theme.background", 0.45),
		BackgroundGradient: giode.Linear(180, giode.Mix("theme.primary", "theme.background", 0.06), "theme.background"),
	},
		giode.Box(giode.Styles{Padding: giode.Inset{Left: 6, Bottom: 18}}, components.BrandMark()),
		giode.Box(giode.Styles{Padding: giode.Inset{Left: 8, Bottom: 4}}, giode.Text("Workspace", theme.Overline)),
		giode.Stack(giode.Styles{Width: -1, Gap: 2}, nav...),
		giode.Box(giode.Styles{Padding: giode.Inset{Left: 8, Top: 18, Bottom: 4}}, giode.Text("Pinned", theme.Overline)),
		giode.Box(giode.Styles{Width: -1, FlexGrow: 1}, s.scroll),
		s.footer(),
	).Layout(gtx)
}
