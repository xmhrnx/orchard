package screens

import (
	"strings"

	"github.com/xmhrnx/giode"
	"github.com/xmhrnx/giode/components/checkbox"
	"github.com/xmhrnx/giode/components/dropdown"
	"github.com/xmhrnx/giode/components/segmented"

	"orchard/internal/core"
	"orchard/ui/components"
	"orchard/ui/env"
	"orchard/ui/theme"
)

// ReposScreen is the overview of every tracked repository: headline
// numbers, filters, and the list with bulk update controls.
type ReposScreen struct {
	env        *env.Env
	sel        *components.Selection
	search     *components.SearchBox
	filter     *segmented.Control
	group      *dropdown.Select
	groupsSeen string
	selectAll  *checkbox.Checkbox
	list       *giode.LazyListElement
	rows       []*core.RepoState
	actions    *components.RepoActions
	bulk       *components.BulkBar
	scan       *ScanDialog
	prompt     *components.PromptDialog
	confirm    *components.ConfirmDialog
	add        *giode.ButtonElement
	scanBtn    *giode.ButtonElement
	fetchAll   *giode.ButtonElement
	updateAll  *giode.ButtonElement
	addFirst   *giode.ButtonElement
	cloneFirst *giode.ButtonElement
}

// NewReposScreen creates the overview. Create it once.
func NewReposScreen(e *env.Env) *ReposScreen {
	s := &ReposScreen{env: e, sel: components.NewSelection()}
	s.search = components.NewSearchBox("Filter by name, path, branch or group")
	s.filter = giode.Input.Segmented(repoFilterNames...)
	s.group = giode.Input.Select(components.FieldStyle().With(giode.Styles{Width: 170}), "All groups")
	s.selectAll = giode.Input.Checkbox("", giode.Styles{Background: theme.Primary, BorderColor: theme.Muted}).
		OnChange(func(on bool) { s.sel.SetAll(s.visibleIDs(), on) })
	s.scan = NewScanDialog(e)
	s.prompt = components.NewPromptDialog()
	s.confirm = components.NewConfirmDialog()
	s.actions = s.repoActions()
	s.list = giode.LazyList(
		func() int { return len(s.rows) },
		func(i int) any { return s.rows[i].ID },
		func(i int) giode.Element {
			return giode.Box(giode.Styles{Width: -1, Padding: giode.Inset{Bottom: 8}}, components.RepoCardFor(s.rows[i], s.sel, s.actions))
		},
		giode.Styles{ScrollBar: giode.ScrollBar{Width: 6, AutoHide: true, TrackColor: "transparent"}},
	)
	s.add = components.Button("Add folder", "folder-plus", components.Secondary, s.addFolder)
	s.scanBtn = components.Button("Scan", "search", components.Secondary, s.scanFolder)
	s.fetchAll = components.Button("Fetch all", "fetch", components.Secondary, func() { s.runOn(core.OpFetch, s.allIDs()) })
	s.updateAll = components.Button("Update all", "pull", components.Primary, func() { s.runOn(core.OpPull, s.allIDs()) })
	s.addFirst = components.Button("Add a folder", "folder-plus", components.Primary, s.addFolder)
	s.cloneFirst = components.Button("Clone a repository", "clone", components.Secondary, func() { e.Nav.Open(env.PageClone) })
	s.bulk = components.NewBulkBar(s.sel, "repositories",
		components.Button("Update", "pull", components.Primary, func() { s.runOn(core.OpPull, s.sel.Keys()) }),
		components.Button("Fetch", "fetch", components.Secondary, func() { s.runOn(core.OpFetch, s.sel.Keys()) }),
		components.Button("Push", "push", components.Secondary, func() { s.runOn(core.OpPush, s.sel.Keys()) }),
		components.Button("Group", "layers", components.Secondary, s.groupSelected),
		components.Button("Untrack", "trash", components.Destructive, func() { s.confirmRemove(s.sel.Keys()) }),
	)
	return s
}

func (s *ReposScreen) repoActions() *components.RepoActions {
	e := s.env
	path := func(id string) string {
		if r := e.WS.Repo(id); r != nil {
			return r.Path
		}
		return ""
	}
	return &components.RepoActions{
		Open:     e.Nav.OpenRepo,
		Pull:     func(id string) { s.runOn(core.OpPull, []string{id}) },
		Fetch:    func(id string) { s.runOn(core.OpFetch, []string{id}) },
		Push:     func(id string) { s.runOn(core.OpPush, []string{id}) },
		Reveal:   func(id string) { e.Reveal(path(id)) },
		Terminal: func(id string) { e.Terminal(path(id)) },
		CopyPath: func(id string) { e.Copy("Path", path(id)) },
		TogglePin: func(id string) {
			if r := e.WS.Repo(id); r != nil {
				e.Fail("Pinning", e.WS.SetPinned(id, !r.Pinned))
			}
		},
		Remove: func(id string) { s.confirmRemove([]string{id}) },
	}
}

func (s *ReposScreen) runOn(op core.Operation, ids []string) {
	if len(ids) == 0 {
		return
	}
	s.env.WS.Run(op, ids, s.env.Report)
}

func (s *ReposScreen) allIDs() []string {
	var ids []string
	for _, r := range s.env.WS.Repos() {
		ids = append(ids, r.ID)
	}
	return ids
}

func (s *ReposScreen) visibleIDs() []string {
	ids := make([]string, len(s.rows))
	for i, r := range s.rows {
		ids[i] = r.ID
	}
	return ids
}

func (s *ReposScreen) addFolder() {
	s.env.PickFolder("Add a repository", s.env.WS.Settings().CloneDir, func(dir string) {
		s.env.WS.AddPaths([]string{dir}, func(added int, err error) {
			if err != nil {
				s.env.Fail("Adding", err)
			} else if added == 0 {
				s.env.Toasts.Show(giode.ToastInfo, "That repository is already in your orchard")
			} else {
				s.env.Toasts.Show(giode.ToastSuccess, "Repository added")
			}
		})
	})
}

func (s *ReposScreen) scanFolder() {
	s.env.PickFolder("Find repositories in…", "", func(dir string) { s.scan.Start(dir) })
}

func (s *ReposScreen) groupSelected() {
	ids := s.sel.Keys()
	s.prompt.Ask("Group repositories", "Group name (leave the same name to add more)", "e.g. Work, Clients, Dotfiles", "", "Group", func(name string) {
		s.env.Fail("Grouping", s.env.WS.SetGroup(name, ids...))
	})
}

func (s *ReposScreen) confirmRemove(ids []string) {
	body := "Orchard will stop tracking these repositories. Nothing is deleted from your disk."
	s.confirm.Ask("Stop tracking?", body, "Untrack", true, func() {
		giode.AfterLayout(func() {
			s.env.Fail("Untracking", s.env.WS.Remove(ids...))
			s.sel.SetAll(ids, false)
		})
	})
}

// syncGroups refreshes the group dropdown when groups come and go.
func (s *ReposScreen) syncGroups() {
	groups := s.env.WS.Groups()
	key := strings.Join(groups, "\x00")
	if key == s.groupsSeen {
		return
	}
	s.groupsSeen = key
	s.group.SetOptions(append([]string{"All groups"}, groups...)...)
}

func (s *ReposScreen) collectRows() {
	group := ""
	if s.group.Selected() > 0 {
		group = s.group.Value()
	}
	f := repoFilter(s.filter.Selected())
	s.rows = s.rows[:0]
	for _, r := range s.env.WS.Repos() {
		if f.matches(r, s.search.Query(), group) {
			s.rows = append(s.rows, r)
		}
	}
	s.sel.Retain(s.allIDs())
	s.selectAll.SetChecked(s.sel.AllOf(s.visibleIDs()))
}

func (s *ReposScreen) toolbar() giode.Element {
	kids := []giode.Element{s.selectAll, giode.Box(giode.Styles{FlexGrow: 1, MinWidth: 220}, s.search), s.filter}
	if len(s.env.WS.Groups()) > 0 {
		kids = append(kids, s.group)
	}
	return giode.Row(giode.Styles{Width: -1, Gap: 12, Align: giode.AlignCenter, Wrap: giode.WrapNoWrap, Padding: giode.Inset{Left: 14}}, kids...)
}

func (s *ReposScreen) body() giode.Element {
	all := s.env.WS.Repos()
	if len(all) == 0 {
		return components.EmptyState("branch", "Plant your first tree",
			"Add repositories already on this computer, scan a folder for all of them at once, or clone new ones from GitHub, GitLab and friends.",
			giode.Row(giode.Styles{Gap: 10}, s.addFirst, s.cloneFirst))
	}
	list := giode.Element(giode.Box(giode.Styles{Width: -1, FlexGrow: 1}, s.list))
	if len(s.rows) == 0 {
		list = components.EmptyState("filter", "Nothing matches", "Try another filter or search.", nil)
	}
	return giode.Stack(giode.Styles{Width: -1, Height: -1, Gap: 18},
		RepoStats(all),
		s.toolbar(),
		list,
	)
}

// Layout draws the screen.
func (s *ReposScreen) Layout(gtx giode.Context) giode.Dimensions {
	s.syncGroups()
	s.collectRows()
	return giode.Box(giode.Styles{Width: -1, Height: -1},
		giode.Stack(giode.Styles{Width: -1, Height: -1, Gap: 22, Padding: giode.Inset{Top: 26, Left: 32, Right: 32}},
			components.PageHeader("Repositories", "Everything you grow, in one place. Update one, a few, or all of them.",
				s.add, s.scanBtn, s.fetchAll, s.updateAll),
			s.body(),
		),
		giode.Box(giode.Styles{Position: giode.PositionAbsolute, Bottom: 24, Left: giode.Flush, Right: giode.Flush, Justify: giode.JustifyCenter, ZIndex: 5}, s.bulk),
		s.scan, s.prompt, s.confirm,
	).Layout(gtx)
}
