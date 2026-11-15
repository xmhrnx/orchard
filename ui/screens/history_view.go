package screens

import (
	"context"
	"fmt"
	"strings"

	"github.com/xmhrnx/giode"
	"github.com/xmhrnx/giode/components/toggle"

	"orchard/internal/core"
	"orchard/internal/gitx"
	"orchard/internal/graph"
	"orchard/ui/components"
	"orchard/ui/env"
	"orchard/ui/theme"
)

// historyLimit bounds how many commits the graph loads.
const historyLimit = 3000

// historyData is a loaded history with its graph layout.
type historyData struct {
	commits []gitx.Commit
	layout  graph.Layout
	head    string
}

// HistoryView is the commit tree: every branch's history drawn as
// lanes, searchable, with the selected commit's details beside it.
type HistoryView struct {
	env      *env.Env
	repo     *core.RepoState
	data     historyData
	loading  bool
	visible  []int
	selected string
	search   *components.SearchBox
	all      *toggle.Switch
	reload   *components.IconAction
	list     *giode.LazyListElement
	panel    *CommitPanel
}

// NewHistoryView creates the view. Create it once.
func NewHistoryView(e *env.Env) *HistoryView {
	v := &HistoryView{env: e}
	v.search = components.NewSearchBox("Search subject, author or hash")
	v.all = giode.Input.Switch("All branches", giode.Styles{FontSize: 12.5, Color: theme.Muted}).OnChange(func(bool) { v.load() })
	v.all.SetChecked(true)
	v.reload = components.NewIconAction("fetch", "Reload", func() { v.load() })
	v.panel = NewCommitPanel(e, func() { v.load() }, func() { v.selected = "" })
	v.list = giode.LazyList(
		func() int { return len(v.visible) },
		func(i int) any { return v.data.commits[v.visible[i]].Hash },
		func(i int) giode.Element { return v.row(v.visible[i]) },
		giode.Styles{ScrollBar: giode.ScrollBar{Width: 6, AutoHide: true, TrackColor: "transparent"}},
	)
	return v
}

// Open loads the history of a repository.
func (v *HistoryView) Open(r *core.RepoState) {
	if v.repo == nil || v.repo.ID != r.ID {
		v.selected = ""
		v.panel.Clear()
	}
	v.repo = r
	v.load()
}

func (v *HistoryView) load() {
	if v.repo == nil {
		return
	}
	path, all := v.repo.Path, v.all.Checked()
	v.loading = true
	core.Query(v.env.WS, func(ctx context.Context) (historyData, error) {
		commits, err := gitx.Log(ctx, path, gitx.LogOptions{Limit: historyLimit, All: all})
		if err != nil {
			return historyData{}, err
		}
		nodes := make([]graph.Node, len(commits))
		for i, c := range commits {
			nodes[i] = graph.Node{Hash: c.Hash, Parents: c.Parents}
		}
		st, _ := gitx.ReadStatus(ctx, path)
		return historyData{commits: commits, layout: graph.Compute(nodes), head: st.Head}, nil
	}, func(d historyData, err error) {
		v.loading = false
		if err != nil {
			v.env.Fail("Loading history", err)
			return
		}
		v.data = d
	})
}

func (v *HistoryView) filter() {
	q := strings.ToLower(strings.TrimSpace(v.search.Query()))
	v.visible = v.visible[:0]
	for i, c := range v.data.commits {
		if q == "" || strings.Contains(strings.ToLower(c.Subject+" "+c.Author+" "+c.Hash), q) {
			v.visible = append(v.visible, i)
		}
	}
}

func (v *HistoryView) row(i int) giode.Element {
	hash := v.data.commits[i].Hash
	return giode.Keyed("history-row", hash, func() *giode.PressableElement {
		return giode.Pressable(func(s giode.PressState) giode.Element {
			idx := v.indexOf(hash)
			if idx < 0 {
				return giode.Box(giode.Styles{})
			}
			c := v.data.commits[idx]
			row, lanes := v.data.layout.Rows[idx], v.data.layout.Lanes
			if v.search.Query() != "" {
				row, lanes = graph.Row{Color: row.Color}, 1
			}
			return components.CommitRow(c, row, lanes, components.CommitRowState{
				Head: c.Hash == v.data.head, Selected: v.selected == hash, Hovered: s.Hovered, Compact: v.selected != "",
			})
		}).Label(v.data.commits[i].Subject).OnClick(func(giode.ClickEvent) {
			idx := v.indexOf(hash)
			if idx < 0 {
				return
			}
			v.selected = hash
			v.panel.Show(v.repo, v.data.commits[idx])
		})
	})
}

func (v *HistoryView) indexOf(hash string) int {
	for _, i := range v.visible {
		if v.data.commits[i].Hash == hash {
			return i
		}
	}
	return -1
}

func (v *HistoryView) listPane() giode.Element {
	if len(v.data.commits) == 0 {
		if v.loading {
			return giode.Box(giode.Styles{Width: -1, Padding: giode.UniformInset(40), Justify: giode.JustifyCenter},
				giode.Spinner(giode.Styles{Width: 28, Height: 28, Color: theme.Primary}))
		}
		return components.EmptyState("history", "No commits yet", "Make the first commit from the Changes tab.", nil)
	}
	return giode.Box(giode.Styles{Width: -1, Height: -1, Padding: giode.UniformInset(6)}, v.list)
}

// Layout draws the view.
func (v *HistoryView) Layout(gtx giode.Context) giode.Dimensions {
	v.filter()
	count := fmt.Sprintf("%d commits", len(v.visible))
	if len(v.data.commits) == historyLimit {
		count = fmt.Sprintf("latest %d commits", historyLimit)
	}
	list := components.FilledFrame(giode.Styles{FlexGrow: 1, Height: -1, Padding: giode.UniformInset(0)}, v.listPane())
	body := []giode.Element{list}
	if v.selected != "" {
		body = append(body, v.panel)
	}
	return giode.Stack(giode.Styles{Width: -1, Height: -1, Gap: 12},
		giode.Row(giode.Styles{Width: -1, Gap: 12, Align: giode.AlignCenter, Wrap: giode.WrapNoWrap},
			giode.Box(giode.Styles{Width: 340}, v.search),
			giode.Box(giode.Styles{Width: 150}, v.all),
			giode.Box(giode.Styles{FlexGrow: 1}, giode.Text(count, theme.Caption)),
			v.reload,
		),
		giode.Row(giode.Styles{Width: -1, FlexGrow: 1, Gap: 14, Wrap: giode.WrapNoWrap}, body...),
	).Layout(gtx)
}
