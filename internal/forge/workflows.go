package forge

import (
	"context"
	"errors"
	"time"
)

// RunState is a CI run's state, normalised across services.
type RunState int

// Run states.
const (
	RunQueued RunState = iota
	RunRunning
	RunSucceeded
	RunFailed
	RunCancelled
	RunSkipped
	RunWaiting
)

// WorkflowRun is one execution of a CI workflow or pipeline.
type WorkflowRun struct {
	ID       string
	Number   int
	Workflow string
	Title    string
	Branch   string
	Commit   string
	Event    string
	Actor    string
	State    RunState
	WebURL   string
	Created  time.Time
	Updated  time.Time
}

// Workflow is a CI definition that can be started by hand.
type Workflow struct {
	ID   string
	Name string
	Path string
}

// WorkflowCaps says which run actions a service offers.
type WorkflowCaps struct {
	Rerun    bool
	Cancel   bool
	Delete   bool
	Dispatch bool
}

// Workflows is the CI API of services that have one. Clients without
// CI support don't implement it.
type Workflows interface {
	// WorkflowCaps reports what the service can do with runs.
	WorkflowCaps() WorkflowCaps
	// WorkflowRuns lists recent runs of a repository, newest first.
	WorkflowRuns(ctx context.Context, repo string) ([]WorkflowRun, error)
	// WorkflowDefinitions lists the workflows that can be started.
	WorkflowDefinitions(ctx context.Context, repo string) ([]Workflow, error)
	// RerunWorkflow runs a finished run again.
	RerunWorkflow(ctx context.Context, repo, id string) error
	// CancelWorkflow stops a queued or running run.
	CancelWorkflow(ctx context.Context, repo, id string) error
	// DeleteWorkflowRun removes a finished run and its logs.
	DeleteWorkflowRun(ctx context.Context, repo, id string) error
	// DispatchWorkflow starts a workflow on a branch or tag.
	DispatchWorkflow(ctx context.Context, repo, workflowID, ref string) error
}

// ErrNoWorkflows is returned for services without a CI API Orchard
// supports.
var ErrNoWorkflows = errors.New("this service's CI isn't supported yet")

// Active reports whether a run hasn't finished.
func (s RunState) Active() bool { return s == RunQueued || s == RunRunning || s == RunWaiting }

// Title is the state as shown to people.
func (s RunState) Title() string {
	return [...]string{"Queued", "Running", "Succeeded", "Failed", "Cancelled", "Skipped", "Waiting"}[s]
}

// Duration is how long a run took, or has been going.
func (r WorkflowRun) Duration() time.Duration {
	if r.Created.IsZero() {
		return 0
	}
	end := r.Updated
	if r.State.Active() || end.Before(r.Created) {
		end = time.Now()
	}
	return end.Sub(r.Created)
}

// WorkflowsOf returns the CI API of a client, if its service has one.
func WorkflowsOf(c Client) (Workflows, bool) {
	w, ok := c.(Workflows)
	return w, ok
}
