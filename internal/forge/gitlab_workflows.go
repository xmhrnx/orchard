package forge

import (
	"context"
	"fmt"
	"net/http"
	"strconv"
	"time"
)

type gitlabPipeline struct {
	ID        int64     `json:"id"`
	IID       int       `json:"iid"`
	Name      string    `json:"name"`
	Ref       string    `json:"ref"`
	SHA       string    `json:"sha"`
	Source    string    `json:"source"`
	Status    string    `json:"status"`
	WebURL    string    `json:"web_url"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

// gitlabPipelineID stands for the project's one pipeline definition.
const gitlabPipelineID = "pipeline"

func (c *gitlabClient) WorkflowCaps() WorkflowCaps {
	return WorkflowCaps{Rerun: true, Cancel: true, Delete: true, Dispatch: true}
}

func (c *gitlabClient) WorkflowRuns(ctx context.Context, repo string) ([]WorkflowRun, error) {
	var list []gitlabPipeline
	if _, err := c.http.get(ctx, c.project(repo)+"/pipelines?per_page=100", &list); err != nil {
		return nil, err
	}
	runs := make([]WorkflowRun, len(list))
	for i, p := range list {
		name := p.Name
		if name == "" {
			name = "Pipeline"
		}
		runs[i] = WorkflowRun{
			ID: strconv.FormatInt(p.ID, 10), Number: p.IID, Workflow: name, Title: name + " on " + p.Ref,
			Branch: p.Ref, Commit: p.SHA, Event: p.Source, State: gitlabState(p.Status),
			WebURL: p.WebURL, Created: p.CreatedAt, Updated: p.UpdatedAt,
		}
	}
	return runs, nil
}

func gitlabState(status string) RunState {
	switch status {
	case "created", "pending", "preparing", "waiting_for_resource", "scheduled":
		return RunQueued
	case "running":
		return RunRunning
	case "manual":
		return RunWaiting
	case "success":
		return RunSucceeded
	case "canceled", "canceling":
		return RunCancelled
	case "skipped":
		return RunSkipped
	}
	return RunFailed
}

func (c *gitlabClient) WorkflowDefinitions(context.Context, string) ([]Workflow, error) {
	return []Workflow{{ID: gitlabPipelineID, Name: "Pipeline", Path: ".gitlab-ci.yml"}}, nil
}

func (c *gitlabClient) pipelineAction(ctx context.Context, method, repo, id, action string) error {
	_, err := c.http.do(ctx, method, fmt.Sprintf("%s/pipelines/%s%s", c.project(repo), id, action), nil, nil)
	return err
}

func (c *gitlabClient) RerunWorkflow(ctx context.Context, repo, id string) error {
	return c.pipelineAction(ctx, http.MethodPost, repo, id, "/retry")
}

func (c *gitlabClient) CancelWorkflow(ctx context.Context, repo, id string) error {
	return c.pipelineAction(ctx, http.MethodPost, repo, id, "/cancel")
}

func (c *gitlabClient) DeleteWorkflowRun(ctx context.Context, repo, id string) error {
	return c.pipelineAction(ctx, http.MethodDelete, repo, id, "")
}

func (c *gitlabClient) DispatchWorkflow(ctx context.Context, repo, _, ref string) error {
	_, err := c.http.do(ctx, http.MethodPost, c.project(repo)+"/pipeline", map[string]string{"ref": ref}, nil)
	return err
}
