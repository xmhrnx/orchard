package forge

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestGithubStateMapping(t *testing.T) {
	cases := []struct {
		status, conclusion string
		want               RunState
	}{
		{"queued", "", RunQueued},
		{"in_progress", "", RunRunning},
		{"completed", "success", RunSucceeded},
		{"completed", "failure", RunFailed},
		{"completed", "timed_out", RunFailed},
		{"completed", "cancelled", RunCancelled},
		{"completed", "skipped", RunSkipped},
		{"waiting", "", RunWaiting},
	}
	for _, c := range cases {
		if got := githubState(c.status, c.conclusion); got != c.want {
			t.Errorf("githubState(%q, %q) = %v, want %v", c.status, c.conclusion, got, c.want)
		}
	}
}

func TestGitlabStateMapping(t *testing.T) {
	for status, want := range map[string]RunState{
		"pending": RunQueued, "running": RunRunning, "success": RunSucceeded,
		"failed": RunFailed, "canceled": RunCancelled, "manual": RunWaiting, "skipped": RunSkipped,
	} {
		if got := gitlabState(status); got != want {
			t.Errorf("gitlabState(%q) = %v, want %v", status, got, want)
		}
	}
}

func TestGithubWorkflowRunsAndActions(t *testing.T) {
	var hits []string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		hits = append(hits, r.Method+" "+r.URL.Path)
		if r.Header.Get("Authorization") != "Bearer tok" {
			w.WriteHeader(http.StatusUnauthorized)
			return
		}
		if r.Method == http.MethodGet {
			json.NewEncoder(w).Encode(map[string]any{"workflow_runs": []map[string]any{{
				"id": 42, "run_number": 7, "name": "CI", "display_title": "Fix build", "head_branch": "main",
				"status": "completed", "conclusion": "failure", "actor": map[string]string{"login": "dev"},
			}}})
			return
		}
		w.WriteHeader(http.StatusNoContent)
	}))
	defer srv.Close()
	c := &githubClient{http: newHTTP("tok"), base: srv.URL}
	ctx := context.Background()
	runs, err := c.WorkflowRuns(ctx, "o/r")
	if err != nil {
		t.Fatal(err)
	}
	if len(runs) != 1 || runs[0].ID != "42" || runs[0].State != RunFailed || runs[0].Actor != "dev" {
		t.Fatalf("runs = %+v", runs)
	}
	if err := c.RerunWorkflow(ctx, "o/r", "42"); err != nil {
		t.Fatal(err)
	}
	if err := c.DeleteWorkflowRun(ctx, "o/r", "42"); err != nil {
		t.Fatal(err)
	}
	want := []string{"GET /repos/o/r/actions/runs", "POST /repos/o/r/actions/runs/42/rerun", "DELETE /repos/o/r/actions/runs/42"}
	for i, h := range want {
		if hits[i] != h {
			t.Errorf("request %d = %q, want %q", i, hits[i], h)
		}
	}
}

func TestEveryClientExceptGiteaHasWorkflows(t *testing.T) {
	for _, k := range Kinds {
		c, err := New(Account{Kind: k, Host: DefaultHost(k)}, "t")
		if err != nil {
			t.Fatal(err)
		}
		_, ok := WorkflowsOf(c)
		if ok == (k == Gitea) {
			t.Errorf("%s: workflows supported = %v", k, ok)
		}
	}
}
