package forge

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	neturl "net/url"
	"time"
)

type bitbucketPipeline struct {
	UUID        string    `json:"uuid"`
	BuildNumber int       `json:"build_number"`
	CreatedOn   time.Time `json:"created_on"`
	CompletedOn time.Time `json:"completed_on"`
	State       struct {
		Name   string `json:"name"`
		Result *struct {
			Name string `json:"name"`
		} `json:"result"`
		Stage *struct {
			Name string `json:"name"`
		} `json:"stage"`
	} `json:"state"`
	Target struct {
		RefName string `json:"ref_name"`
		Commit  struct {
			Hash string `json:"hash"`
		} `json:"commit"`
	} `json:"target"`
	Trigger struct {
		Name string `json:"name"`
	} `json:"trigger"`
	Creator struct {
		DisplayName string `json:"display_name"`
	} `json:"creator"`
}

// errBitbucketUnsupported marks run actions Bitbucket's API lacks.
var errBitbucketUnsupported = errors.New("Bitbucket Pipelines can't do this through its API")

// bitbucketPipelineID stands for the repository's pipeline definition.
const bitbucketPipelineID = "pipeline"

func (c *bitbucketClient) WorkflowCaps() WorkflowCaps {
	return WorkflowCaps{Cancel: true, Dispatch: true}
}

func (c *bitbucketClient) WorkflowRuns(ctx context.Context, repo string) ([]WorkflowRun, error) {
	var page bitbucketPage[bitbucketPipeline]
	url := fmt.Sprintf("%s/repositories/%s/pipelines/?sort=-created_on&pagelen=50", c.base, repo)
	if _, err := c.http.get(ctx, url, &page); err != nil {
		return nil, err
	}
	runs := make([]WorkflowRun, len(page.Values))
	for i, p := range page.Values {
		runs[i] = WorkflowRun{
			ID: p.UUID, Number: p.BuildNumber, Workflow: "Pipeline",
			Title:  fmt.Sprintf("Pipeline #%d on %s", p.BuildNumber, p.Target.RefName),
			Branch: p.Target.RefName, Commit: p.Target.Commit.Hash, Event: p.Trigger.Name, Actor: p.Creator.DisplayName,
			State: bitbucketState(p), WebURL: fmt.Sprintf("https://bitbucket.org/%s/pipelines/results/%d", repo, p.BuildNumber),
			Created: p.CreatedOn, Updated: p.CompletedOn,
		}
	}
	return runs, nil
}

func bitbucketState(p bitbucketPipeline) RunState {
	switch p.State.Name {
	case "PENDING":
		return RunQueued
	case "IN_PROGRESS":
		if p.State.Stage != nil && p.State.Stage.Name == "PAUSED" {
			return RunWaiting
		}
		return RunRunning
	}
	if p.State.Result == nil {
		return RunFailed
	}
	switch p.State.Result.Name {
	case "SUCCESSFUL":
		return RunSucceeded
	case "STOPPED":
		return RunCancelled
	}
	return RunFailed
}

func (c *bitbucketClient) WorkflowDefinitions(context.Context, string) ([]Workflow, error) {
	return []Workflow{{ID: bitbucketPipelineID, Name: "Pipeline", Path: "bitbucket-pipelines.yml"}}, nil
}

func (c *bitbucketClient) RerunWorkflow(context.Context, string, string) error {
	return errBitbucketUnsupported
}

func (c *bitbucketClient) CancelWorkflow(ctx context.Context, repo, id string) error {
	url := fmt.Sprintf("%s/repositories/%s/pipelines/%s/stopPipeline", c.base, repo, neturl.PathEscape(id))
	_, err := c.http.do(ctx, http.MethodPost, url, nil, nil)
	return err
}

func (c *bitbucketClient) DeleteWorkflowRun(context.Context, string, string) error {
	return errBitbucketUnsupported
}

func (c *bitbucketClient) DispatchWorkflow(ctx context.Context, repo, _, ref string) error {
	body := map[string]any{"target": map[string]string{"type": "pipeline_ref_target", "ref_type": "branch", "ref_name": ref}}
	_, err := c.http.do(ctx, http.MethodPost, fmt.Sprintf("%s/repositories/%s/pipelines/", c.base, repo), body, nil)
	return err
}
