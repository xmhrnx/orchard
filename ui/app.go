// Package ui is Orchard's interface: the root view wiring the window
// frame, sidebar, screens and overlays together.
package ui

import (
	"github.com/xmhrnx/giode"

	"orchard/internal/core"
	"orchard/internal/store"
	"orchard/internal/tasks"
	"orchard/ui/components"
	"orchard/ui/env"
	"orchard/ui/screens"
	"orchard/ui/shell"
	"orchard/ui/theme"
)

// Root owns every screen and the chrome around them.
type Root struct {
	env      *env.Env
	bar      *shell.WindowBar
	sidebar  *shell.Sidebar
	palette  *shell.CommandPalette
	repos    *screens.ReposScreen
	repo     *screens.RepoScreen
	clone    *screens.CloneScreen
	accounts *screens.AccountsScreen
	settings *screens.SettingsScreen
	stage    *giode.TransitionElement
}

// New builds the interface for a workspace.
func New(app *giode.App, ws *core.Workspace, runner *tasks.Runner) *Root {
	e := &env.Env{
		App:    app,
		WS:     ws,
		Tasks:  runner,
		Nav:    &env.Navigator{},
		Toasts: giode.Toaster().Placement(giode.ToastBottomEnd).Max(4).Inset(giode.Inset{Bottom: 8, Right: 8}),
	}
	r := &Root{env: e}
	r.bar = shell.NewWindowBar(app, r.openPalette, r.toggleTheme)
	r.sidebar = shell.NewSidebar(e)
	r.palette = shell.NewCommandPalette(r.commands)
	r.repos = screens.NewReposScreen(e)
	r.repo = screens.NewRepoScreen(e)
	r.clone = screens.NewCloneScreen(e)
	r.accounts = screens.NewAccountsScreen(e)
	r.settings = screens.NewSettingsScreen(e, r.applyTheme)
	r.stage = giode.Transition(giode.Raw(r.layoutPage), giode.Fade|giode.SlideUp, theme.Smooth, giode.EaseOut)
	e.Nav.OnChange(r.navigated)
	return r
}

// Env exposes the shared environment, for the app's startup wiring.
func (r *Root) Env() *env.Env { return r.env }

func (r *Root) navigated(to env.Route) {
	if to.Page == env.PageRepo {
		r.repo.Open(to.RepoID)
	}
	r.stage.Hidden()
	r.stage.Show()
}

func (r *Root) openPalette() { r.palette.Open() }

func (r *Root) toggleTheme() { r.applyTheme(!theme.IsDark()) }

// applyTheme switches the palette and remembers the choice.
func (r *Root) applyTheme(dark bool) {
	theme.SetDark(dark)
	name := "light"
	if dark {
		name = "dark"
	}
	r.env.Fail("Saving settings", r.env.WS.UpdateSettings(func(s *store.Settings) { s.Theme = name }))
}

func (r *Root) layoutPage(gtx giode.Context) giode.Dimensions {
	return r.page().Layout(gtx)
}

func (r *Root) page() giode.Element {
	switch r.env.Nav.Current().Page {
	case env.PageRepo:
		return r.repo
	case env.PageClone:
		return r.clone
	case env.PageAccounts:
		return r.accounts
	case env.PageSettings:
		return r.settings
	}
	return r.repos
}

func (r *Root) title() string {
	switch r.env.Nav.Current().Page {
	case env.PageRepo:
		if repo := r.env.WS.Repo(r.env.Nav.Current().RepoID); repo != nil {
			return repo.Name
		}
	case env.PageClone:
		return "Clone"
	case env.PageAccounts:
		return "Accounts"
	case env.PageSettings:
		return "Settings"
	}
	return "Orchard"
}

// View is the frame's root element.
func (r *Root) View() giode.Element {
	r.shortcuts()
	r.bar.SetTitle(r.title())
	return giode.Stack(giode.Styles{Width: -1, Height: -1, Background: theme.Canvas},
		r.bar,
		giode.Row(giode.Styles{Width: -1, FlexGrow: 1, Wrap: giode.WrapNoWrap},
			r.sidebar,
			giode.Box(giode.Styles{Width: 1, Height: -1, Background: giode.Mix("theme.border", "theme.background", 0.5)}),
			giode.Stack(giode.Styles{FlexGrow: 1, Height: -1},
				components.BusyLine(r.env.WS.Busy() > 0),
				giode.Box(giode.Styles{Width: -1, FlexGrow: 1, Overflow: giode.OverflowHidden}, r.stage),
			),
		),
		r.palette,
		r.env.Toasts,
	)
}
