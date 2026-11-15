package screens

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/xmhrnx/giode"
	"github.com/xmhrnx/giode/components/checkbox"
	"github.com/xmhrnx/giode/components/segmented"

	"orchard/internal/core"
	"orchard/internal/forge"
	"orchard/ui/components"
	"orchard/ui/env"
	"orchard/ui/theme"
)

// workflowPoll is how often runs refresh while some are still going.
const workflowPoll = 8 * time.Second

// runAction is something that can be done to many runs at once.
type runAction int

const (
	actRerun runAction = iota
	actCancel
	actDelete
)

// workflowsData is a loaded run list and what the service allows.
type workflowsData struct {
	runs []forge.WorkflowRun
	caps forge.WorkflowCaps
}

// WorkflowsView shows a repository's CI runs (GitHub Actions, GitLab
// pipelines, Bitbucket Pipelines) with live status, and re-runs,
// cancels or deletes many at once.
type WorkflowsView struct {
	env        *env.Env
	repo       *core.RepoState
	account    forge.Account
	slug       string
	linked     bool
	supported  bool
	loading    bool
	polling    bool
	laidOut    time.Time
	data       workflowsData
	shown      []forge.WorkflowRun
	sel        *components.Selection
	search     *components.SearchBox
	filter     *segmented.Control
	selectAll  *checkbox.Checkbox
	list       *giode.LazyListElement
	bulk       *components.BulkBar
	bulkRerun  *giode.ButtonElement
	bulkCancel *giode.ButtonElement
	bulkDelete *giode.ButtonElement
	runBtn     *giode.ButtonElement
	reload     *components.IconAction
	connect    *giode.ButtonElement
	confirm    *components.ConfirmDialog
	dispatch   *RunWorkflowDialog
}

// NewWorkflowsView creates the view. Create it once.
func NewWorkflowsView(e *env.Env) *WorkflowsView {
	v := &WorkflowsView{env: e, sel: components.NewSelection(), confirm: components.NewConfirmDialog()}
	v.search = components.NewSearchBox("Filter by title, workflow, branch or author")
	v.filter = giode.Input.Segmented("All", "Active", "Failed", "Succeeded")
	v.selectAll = giode.Input.Checkbox("", giode.Styles{Background: theme.Primary, BorderColor: theme.Muted}).
		OnChange(func(on bool) { v.sel.SetAll(v.shownIDs(), on) })
	v.reload = components.NewIconAction("fetch", "Reload", v.load)
	v.connect = components.Button("Connect an account", "user", components.Primary, func() { e.Nav.Open(env.PageAccounts) })
	v.dispatch = NewRunWorkflowDialog(e, func() { time.AfterFunc(3*time.Second, func() { e.App.Post(v.load) }) })
	v.runBtn = components.Button("Run workflow", "play", components.Primary, func() {
		v.dispatch.Open(v.account, v.slug, v.repo.Status.Branch)
	})
	v.list = giode.LazyList(
		func() int { return len(v.shown) },
		func(i int) any { return v.shown[i].ID },
		func(i int) giode.Element { return v.row(v.shown[i]) },
		giode.Styles{Gap: 2, ScrollBar: giode.ScrollBar{Width: 6, AutoHide: true, TrackColor: "transparent"}},
	)
	v.bulkRerun = components.Button("Re-run", "fetch", components.Secondary, func() { v.confirmBulk(actRerun, v.sel.Keys()) })
	v.bulkCancel = components.Button("Cancel", "stop", components.Secondary, func() { v.confirmBulk(actCancel, v.sel.Keys()) })
	v.bulkDelete = components.Button("Delete", "trash", components.Destructive, func() { v.confirmBulk(actDelete, v.sel.Keys()) })
	v.bulk = components.NewBulkBar(v.sel, "runs", v.bulkRerun, v.bulkCancel, v.bulkDelete)
	return v
}

// Open loads the runs of a repository.
func (v *WorkflowsView) Open(r *core.RepoState) {
	if v.repo == nil || v.repo.ID != r.ID {
		v.sel.Clear()
		v.data = workflowsData{}
	}
	v.repo = r
	v.account, v.slug, v.linked = v.env.WS.HostedRepo(r)
	v.supported = true
	v.load()
}

func (v *WorkflowsView) load() {
	if !v.linked {
		return
	}
	acc, slug, repoID := v.account, v.slug, v.repo.ID
	v.loading = true
	core.Query(v.env.WS, func(ctx context.Context) (workflowsData, error) {
		client, err := v.env.WS.Client(acc)
		if err != nil {
			return workflowsData{}, err
		}
		wf, ok := forge.WorkflowsOf(client)
		if !ok {
			return workflowsData{}, forge.ErrNoWorkflows
		}
		runs, err := wf.WorkflowRuns(ctx, slug)
		return workflowsData{runs: runs, caps: wf.WorkflowCaps()}, err
	}, func(d workflowsData, err error) {
		if v.repo == nil || v.repo.ID != repoID {
			return
		}
		v.loading = false
		if errors.Is(err, forge.ErrNoWorkflows) {
			v.supported = false
			return
		}
		if err != nil {
			v.env.Fail("Loading workflow runs", err)
			return
		}
		v.data = d
		v.sel.Retain(v.allIDs())
		v.schedulePoll()
	})
}

// schedulePoll refreshes the runs shortly while any is still going and
// the tab is on screen, so statuses update without reloading by hand.
func (v *WorkflowsView) schedulePoll() {
	active := false
	for _, r := range v.data.runs {
		active = active || r.State.Active()
	}
	if !active || v.polling {
		return
	}
	v.polling = true
	repoID := v.repo.ID
	time.AfterFunc(workflowPoll, func() {
		v.env.App.Post(func() {
			v.polling = false
			if v.repo != nil && v.repo.ID == repoID && time.Since(v.laidOut) < 2*time.Second {
				v.load()
			}
		})
	})
}

// applicable keeps the runs an action makes sense for: re-run and delete
// finished runs, cancel running ones.
func (v *WorkflowsView) applicable(act runAction, ids []string) []forge.WorkflowRun {
	want := map[string]bool{}
	for _, id := range ids {
		want[id] = true
	}
	var out []forge.WorkflowRun
	for _, r := range v.data.runs {
		if want[r.ID] && r.State.Active() == (act == actCancel) {
			out = append(out, r)
		}
	}
	return out
}

func (v *WorkflowsView) supports(act runAction) bool {
	c := v.data.caps
	return [...]bool{c.Rerun, c.Cancel, c.Delete}[act]
}

func (v *WorkflowsView) confirmBulk(act runAction, ids []string) {
	if !v.supports(act) {
		v.env.Toasts.Show(giode.ToastInfo, v.account.Kind.Title()+" doesn't offer this through its API")
		return
	}
	runs := v.applicable(act, ids)
	skipped := len(ids) - len(runs)
	verb := [...]string{"Re-run", "Cancel", "Delete"}[act]
	if len(runs) == 0 {
		why := [...]string{"Only finished runs can be re-run.", "Only queued or running runs can be cancelled.", "Only finished runs can be deleted."}[act]
		v.env.Toasts.Show(giode.ToastInfo, why)
		return
	}
	body := fmt.Sprintf("%d runs will be %s.", len(runs), [...]string{"started again", "stopped", "deleted with their logs"}[act])
	if skipped > 0 {
		body += fmt.Sprintf(" %d selected runs don't apply and are left alone.", skipped)
	}
	v.confirm.Ask(fmt.Sprintf("%s %d runs?", verb, len(runs)), body, verb, act == actDelete, func() {
		v.sel.Clear()
		v.apply(act, runs)
	})
}

func (v *WorkflowsView) apply(act runAction, runs []forge.WorkflowRun) {
	slug := v.slug
	forgeBulk(v.env, v.account, runs, func(r forge.WorkflowRun) string { return fmt.Sprintf("run #%d", r.Number) },
		func(ctx context.Context, c forge.Client, r forge.WorkflowRun) error {
			wf, ok := forge.WorkflowsOf(c)
			if !ok {
				return forge.ErrNoWorkflows
			}
			switch act {
			case actRerun:
				return wf.RerunWorkflow(ctx, slug, r.ID)
			case actCancel:
				return wf.CancelWorkflow(ctx, slug, r.ID)
			}
			return wf.DeleteWorkflowRun(ctx, slug, r.ID)
		}, func(ok int, err error) {
			if ok > 0 {
				v.env.Toasts.Show(giode.ToastSuccess, fmt.Sprintf("%s %d runs", [...]string{"Restarted", "Cancelled", "Deleted"}[act], ok))
			}
			v.env.Fail("Some runs failed", err)
			v.load()
		})
}

func (v *WorkflowsView) row(r forge.WorkflowRun) giode.Element {
	return components.SelectRow("run-row", r.ID, v.sel, func() giode.Element {
		meta := []string{r.Workflow, fmt.Sprintf("#%d", r.Number)}
		for _, part := range []string{r.Event, r.Branch, shortHash(r.Commit), r.Actor} {
			if part != "" {
				meta = append(meta, part)
			}
		}
		title := r.Title
		if title == "" {
			title = r.Workflow
		}
		return giode.Row(giode.Styles{Width: -1, Gap: 12, Align: giode.AlignCenter, Wrap: giode.WrapNoWrap},
			components.RunBadge(r.State),
			giode.Stack(giode.Styles{FlexGrow: 1, Gap: 3},
				giode.Text(title, theme.Ellipsis(theme.Strong)),
				giode.Text(strings.Join(meta, " · "), theme.Ellipsis(theme.Caption))),
			giode.Stack(giode.Styles{Width: 110, Gap: 3, Align: giode.AlignEnd},
				giode.Text(r.State.Title(), giode.Styles{FontSize: 12, FontWeight: giode.FontWeightMedium, Color: runColor(r.State)}),
				giode.Text(strings.TrimSpace(components.Duration(r.Duration())+"  "+components.Ago(r.Created)), theme.Caption)),
			components.RowMenu("run", r.ID,
				components.MenuItem{Icon: "external", Label: "Open in browser", Run: func() { v.env.OpenURL(r.WebURL) }},
				components.MenuItem{Icon: "copy", Label: "Copy link", Run: func() { v.env.Copy("Link", r.WebURL) }},
				components.MenuSeparator,
				components.MenuItem{Icon: "fetch", Label: "Re-run", Run: func() { v.confirmBulk(actRerun, []string{r.ID}) }},
				components.MenuItem{Icon: "stop", Label: "Cancel", Run: func() { v.confirmBulk(actCancel, []string{r.ID}) }},
				components.MenuItem{Icon: "trash", Label: "Delete", Run: func() { v.confirmBulk(actDelete, []string{r.ID}) }},
			),
		)
	})
}

func runColor(s forge.RunState) giode.Color {
	switch s {
	case forge.RunSucceeded:
		return "theme.success"
	case forge.RunFailed:
		return theme.Danger
	case forge.RunRunning:
		return theme.Info
	case forge.RunWaiting:
		return theme.Warning
	}
	return theme.Muted
}

func shortHash(h string) string {
	if len(h) > 7 {
		return h[:7]
	}
	return h
}

func (v *WorkflowsView) allIDs() []string {
	ids := make([]string, len(v.data.runs))
	for i, r := range v.data.runs {
		ids[i] = r.ID
	}
	return ids
}

func (v *WorkflowsView) shownIDs() []string {
	ids := make([]string, len(v.shown))
	for i, r := range v.shown {
		ids[i] = r.ID
	}
	return ids
}

func (v *WorkflowsView) filterRuns() {
	q := strings.ToLower(strings.TrimSpace(v.search.Query()))
	mode := v.filter.Selected()
	v.shown = v.shown[:0]
	for _, r := range v.data.runs {
		switch {
		case mode == 1 && !r.State.Active(),
			mode == 2 && r.State != forge.RunFailed,
			mode == 3 && r.State != forge.RunSucceeded:
			continue
		}
		if q == "" || strings.Contains(strings.ToLower(r.Title+" "+r.Workflow+" "+r.Branch+" "+r.Actor), q) {
			v.shown = append(v.shown, r)
		}
	}
	v.selectAll.SetChecked(v.sel.AllOf(v.shownIDs()))
}

func (v *WorkflowsView) body() giode.Element {
	switch {
	case !v.linked:
		host := "this repository's host"
		if loc, ok := forge.ParseRemote(v.repo.Origin); ok {
			host = loc.Host
		}
		return components.EmptyState("workflow", "Connect "+host,
			"Workflow runs come from the hosting service. Connect an account for "+host+" to watch, re-run, cancel and clean up CI runs here.", v.connect)
	case !v.supported:
		return components.EmptyState("workflow", "CI not available here",
			v.account.Kind.Title()+"'s CI isn't supported yet. GitHub Actions, GitLab pipelines and Bitbucket Pipelines are.", nil)
	case v.loading && len(v.data.runs) == 0:
		return giode.Box(giode.Styles{Width: -1, Padding: giode.UniformInset(40), Justify: giode.JustifyCenter},
			giode.Spinner(giode.Styles{Width: 28, Height: 28, Color: theme.Primary}))
	case len(v.data.runs) == 0:
		return components.EmptyState("workflow", "No runs yet", "Push a commit or start a workflow by hand to see it here.", nil)
	case len(v.shown) == 0:
		return components.EmptyState("filter", "Nothing matches", "Try another filter.", nil)
	}
	return giode.Box(giode.Styles{Width: -1, Height: -1, Padding: giode.UniformInset(8)}, v.list)
}

// Layout draws the view.
func (v *WorkflowsView) Layout(gtx giode.Context) giode.Dimensions {
	v.laidOut = time.Now()
	v.filterRuns()
	caps := v.data.caps
	v.bulkRerun.Disabled(!caps.Rerun)
	v.bulkCancel.Disabled(!caps.Cancel)
	v.bulkDelete.Disabled(!caps.Delete)
	v.runBtn.Disabled(!caps.Dispatch)
	var head []giode.Element
	if v.linked && v.supported {
		head = []giode.Element{v.selectAll, giode.Box(giode.Styles{Width: 320}, v.search), v.filter,
			giode.Box(giode.Styles{FlexGrow: 1}), v.reload, v.runBtn}
	} else {
		head = []giode.Element{giode.Box(giode.Styles{FlexGrow: 1}, giode.Text(v.slug, theme.CodeDim))}
	}
	return giode.Box(giode.Styles{Width: -1, Height: -1},
		giode.Stack(giode.Styles{Width: -1, Height: -1, Gap: 12},
			giode.Row(giode.Styles{Width: -1, Gap: 10, Align: giode.AlignCenter, Wrap: giode.WrapNoWrap, Padding: giode.Inset{Left: 20}}, head...),
			components.ListFrame(v.body()),
		),
		giode.Box(giode.Styles{Position: giode.PositionAbsolute, Bottom: 20, Left: giode.Flush, Right: giode.Flush, Justify: giode.JustifyCenter, ZIndex: 5}, v.bulk),
		v.confirm, v.dispatch,
	).Layout(gtx)
}
