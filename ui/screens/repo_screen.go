package screens

import (
	"github.com/xmhrnx/giode"

	"orchard/ui/components"
	"orchard/ui/env"
	"orchard/ui/theme"
)

// RepoScreen is one repository in depth: its header and a tab for each
// area of work.
type RepoScreen struct {
	env    *env.Env
	id     string
	header *RepoHeader
	tabs   *components.TabStrip
	views  []repoView
	stage  *giode.TransitionElement
}

// NewRepoScreen creates the page. Create it once; Open shows a
// repository.
func NewRepoScreen(e *env.Env) *RepoScreen {
	s := &RepoScreen{env: e}
	s.header = NewRepoHeader(e, func() string { return s.id })
	changes := NewChangesView(e)
	s.views = []repoView{
		NewHistoryView(e),
		changes,
		NewBranchesView(e),
		NewTagsView(e),
		NewReleasesView(e),
		NewStashesView(e),
		NewRemotesView(e),
		NewPullsView(e),
		NewWorkflowsView(e),
		NewRewriteView(e),
	}
	s.tabs = components.NewTabStrip(s.switchTab,
		components.Tab{Label: "History", Icon: "history"},
		components.Tab{Label: "Changes", Icon: "file", Count: func() int {
			if r := e.WS.Repo(s.id); r != nil {
				return r.Status.ChangeCount()
			}
			return 0
		}},
		components.Tab{Label: "Branches", Icon: "branch"},
		components.Tab{Label: "Tags", Icon: "tag"},
		components.Tab{Label: "Releases", Icon: "release"},
		components.Tab{Label: "Stashes", Icon: "stash"},
		components.Tab{Label: "Remotes", Icon: "globe"},
		components.Tab{Label: "Pull requests", Icon: "pull-request"},
		components.Tab{Label: "Workflows", Icon: "workflow"},
		components.Tab{Label: "Rewrite", Icon: "scissors"},
	)
	s.stage = giode.Transition(giode.Raw(s.layoutView), giode.Fade|giode.SlideUp, theme.Quick*2, giode.EaseOut)
	return s
}

// Open shows a repository, reloading the active tab.
func (s *RepoScreen) Open(id string) {
	s.id = id
	if r := s.env.WS.Repo(id); r != nil {
		s.env.WS.Refresh(id)
		s.views[s.tabs.Active()].Open(r)
	}
}

func (s *RepoScreen) switchTab(i int) {
	if r := s.env.WS.Repo(s.id); r != nil {
		s.views[i].Open(r)
	}
	s.stage.Hidden()
	s.stage.Show()
}

func (s *RepoScreen) layoutView(gtx giode.Context) giode.Dimensions {
	return s.views[s.tabs.Active()].Layout(gtx)
}

// Layout draws the page.
func (s *RepoScreen) Layout(gtx giode.Context) giode.Dimensions {
	if s.env.WS.Repo(s.id) == nil {
		return components.EmptyState("repo", "Repository not found", "It may have been removed from your orchard.", nil).Layout(gtx)
	}
	return giode.Stack(giode.Styles{Width: -1, Height: -1, Gap: 18, Padding: giode.Inset{Top: 22, Left: 28, Right: 28, Bottom: 20}},
		s.header,
		s.tabs,
		giode.Box(giode.Styles{Width: -1, FlexGrow: 1}, s.stage),
	).Layout(gtx)
}
