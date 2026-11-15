package screens

import (
	"context"
	"strings"

	"github.com/xmhrnx/giode"
	"github.com/xmhrnx/giode/components/dialog"
	"github.com/xmhrnx/giode/components/dropdown"
	"github.com/xmhrnx/giode/components/input"

	"orchard/internal/core"
	"orchard/internal/forge"
	"orchard/ui/components"
	"orchard/ui/env"
	"orchard/ui/theme"
)

// RunWorkflowDialog starts a workflow by hand on a branch or tag.
type RunWorkflowDialog struct {
	env       *env.Env
	open      bool
	account   forge.Account
	slug      string
	workflows []forge.Workflow
	loading   bool
	onStarted func()
	pick      *dropdown.Select
	ref       *input.Input
	start     *giode.ButtonElement
	cancel    *giode.ButtonElement
	dlg       *dialog.Dialog
}

// NewRunWorkflowDialog creates the dialog; started runs after a
// workflow was started. Create it once.
func NewRunWorkflowDialog(e *env.Env, started func()) *RunWorkflowDialog {
	d := &RunWorkflowDialog{env: e, onStarted: started}
	d.pick = giode.Input.Select(components.FieldStyle().With(giode.Styles{Width: -1}), "Loading…")
	d.ref = giode.Input.Text("main", components.FieldStyle().With(giode.Styles{FontFamily: theme.Mono})).Submit(func(string) { d.submit() })
	d.start = components.Button("Run", "play", components.Primary, d.submit)
	d.cancel = components.Button("Cancel", "", components.Secondary, func() { d.open = false })
	d.dlg = giode.Dialog(&d.open, d.content, components.DialogPanel())
	return d
}

// Open shows the dialog for a hosted repository, on branch ref.
func (d *RunWorkflowDialog) Open(acc forge.Account, slug, ref string) {
	d.account, d.slug, d.open, d.loading = acc, slug, true, true
	d.workflows = nil
	d.pick.SetOptions("Loading…")
	d.ref.SetText(ref)
	core.Query(d.env.WS, func(ctx context.Context) ([]forge.Workflow, error) {
		client, err := d.env.WS.Client(acc)
		if err != nil {
			return nil, err
		}
		wf, ok := forge.WorkflowsOf(client)
		if !ok {
			return nil, forge.ErrNoWorkflows
		}
		return wf.WorkflowDefinitions(ctx, slug)
	}, func(list []forge.Workflow, err error) {
		d.loading = false
		if err != nil {
			d.env.Fail("Loading workflows", err)
			return
		}
		d.workflows = list
		names := make([]string, len(list))
		for i, w := range list {
			names[i] = w.Name + "  ·  " + w.Path
		}
		if len(names) == 0 {
			names = []string{"No workflows found"}
		}
		d.pick.SetOptions(names...)
	})
}

func (d *RunWorkflowDialog) submit() {
	i, ref := d.pick.Selected(), strings.TrimSpace(d.ref.Text())
	if i < 0 || i >= len(d.workflows) || ref == "" {
		return
	}
	w, acc, slug := d.workflows[i], d.account, d.slug
	d.start.Loading(true)
	core.Query(d.env.WS, func(ctx context.Context) (struct{}, error) {
		client, err := d.env.WS.Client(acc)
		if err != nil {
			return struct{}{}, err
		}
		wf, ok := forge.WorkflowsOf(client)
		if !ok {
			return struct{}{}, forge.ErrNoWorkflows
		}
		return struct{}{}, wf.DispatchWorkflow(ctx, slug, w.ID, ref)
	}, func(_ struct{}, err error) {
		d.start.Loading(false)
		if err != nil {
			d.env.Fail("Starting "+w.Name, err)
			return
		}
		d.open = false
		d.env.Toasts.Show(giode.ToastSuccess, w.Name+" started on "+ref)
		d.onStarted()
	})
}

func (d *RunWorkflowDialog) content() giode.Element {
	d.start.Disabled(d.loading || len(d.workflows) == 0)
	return giode.Stack(giode.Styles{Gap: 16, Width: 480},
		giode.Text("Run a workflow", theme.Heading),
		components.Field("Workflow", "On GitHub the workflow needs a workflow_dispatch trigger.", d.pick),
		components.Field("Branch or tag", "", d.ref),
		giode.Row(giode.Styles{Width: -1, Gap: 8, Justify: giode.JustifyEnd}, d.cancel, d.start),
	)
}

// Layout draws the dialog.
func (d *RunWorkflowDialog) Layout(gtx giode.Context) giode.Dimensions { return d.dlg.Layout(gtx) }
