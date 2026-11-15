package forge

import (
	"context"
	"fmt"
	"net/http"
	"strconv"
	"time"
)

type githubRun struct {
	ID           int64     `json:"id"`
	RunNumber    int       `json:"run_number"`
	Name         string    `json:"name"`
	DisplayTitle string    `json:"display_title"`
	HeadBranch   string    `json:"head_branch"`
	HeadSHA      string    `json:"head_sha"`
	Event        string    `json:"event"`
	Status       string    `json:"status"`
	Conclusion   string    `json:"conclusion"`
	HTMLURL      string    `json:"html_url"`
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`
	Actor        struct {
		Login string `json:"login"`
	} `json:"actor"`
}

func (c *githubClient) WorkflowCaps() WorkflowCaps {
	return WorkflowCaps{Rerun: true, Cancel: true, Delete: true, Dispatch: true}
}

func (c *githubClient) WorkflowRuns(ctx context.Context, repo string) ([]WorkflowRun, error) {
	var page struct {
		Runs []githubRun `json:"workflow_runs"`
	}
	if _, err := c.http.get(ctx, fmt.Sprintf("%s/repos/%s/actions/runs?per_page=100", c.base, repo), &page); err != nil {
		return nil, err
	}
	runs := make([]WorkflowRun, len(page.Runs))
	for i, r := range page.Runs {
		runs[i] = WorkflowRun{
			ID: strconv.FormatInt(r.ID, 10), Number: r.RunNumber, Workflow: r.Name, Title: r.DisplayTitle,
			Branch: r.HeadBranch, Commit: r.HeadSHA, Event: r.Event, Actor: r.Actor.Login,
			State: githubState(r.Status, r.Conclusion), WebURL: r.HTMLURL, Created: r.CreatedAt, Updated: r.UpdatedAt,
		}
	}
	return runs, nil
}

func githubState(status, conclusion string) RunState {
	switch status {
	case "queued", "requested", "pending":
		return RunQueued
	case "in_progress":
		return RunRunning
	case "waiting", "action_required":
		return RunWaiting
	}
	switch conclusion {
	case "success", "neutral":
		return RunSucceeded
	case "cancelled", "stale":
		return RunCancelled
	case "skipped":
		return RunSkipped
	case "action_required":
		return RunWaiting
	}
	return RunFailed
}

func (c *githubClient) WorkflowDefinitions(ctx context.Context, repo string) ([]Workflow, error) {
	var page struct {
		Workflows []struct {
			ID    int64  `json:"id"`
			Name  string `json:"name"`
			Path  string `json:"path"`
			State string `json:"state"`
		} `json:"workflows"`
	}
	if _, err := c.http.get(ctx, fmt.Sprintf("%s/repos/%s/actions/workflows?per_page=100", c.base, repo), &page); err != nil {
		return nil, err
	}
	var list []Workflow
	for _, w := range page.Workflows {
		if w.State == "active" {
			list = append(list, Workflow{ID: strconv.FormatInt(w.ID, 10), Name: w.Name, Path: w.Path})
		}
	}
	return list, nil
}

func (c *githubClient) runAction(ctx context.Context, method, repo, id, action string) error {
	url := fmt.Sprintf("%s/repos/%s/actions/runs/%s%s", c.base, repo, id, action)
	_, err := c.http.do(ctx, method, url, nil, nil)
	return err
}

func (c *githubClient) RerunWorkflow(ctx context.Context, repo, id string) error {
	return c.runAction(ctx, http.MethodPost, repo, id, "/rerun")
}

func (c *githubClient) CancelWorkflow(ctx context.Context, repo, id string) error {
	return c.runAction(ctx, http.MethodPost, repo, id, "/cancel")
}

func (c *githubClient) DeleteWorkflowRun(ctx context.Context, repo, id string) error {
	return c.runAction(ctx, http.MethodDelete, repo, id, "")
}

func (c *githubClient) DispatchWorkflow(ctx context.Context, repo, workflowID, ref string) error {
	url := fmt.Sprintf("%s/repos/%s/actions/workflows/%s/dispatches", c.base, repo, workflowID)
	_, err := c.http.do(ctx, http.MethodPost, url, map[string]string{"ref": ref}, nil)
	return err
}
