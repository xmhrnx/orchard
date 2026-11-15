package forge

import (
	"context"
	"fmt"
	"net/http"
	"net/url"
	"time"
)

type gitlabClient struct {
	http *httpClient
	base string
}

type gitlabProject struct {
	PathWithNamespace string    `json:"path_with_namespace"`
	Description       string    `json:"description"`
	HTTPURL           string    `json:"http_url_to_repo"`
	SSHURL            string    `json:"ssh_url_to_repo"`
	WebURL            string    `json:"web_url"`
	DefaultBranch     string    `json:"default_branch"`
	Visibility        string    `json:"visibility"`
	Stars             int       `json:"star_count"`
	LastActivity      time.Time `json:"last_activity_at"`
	ForkedFrom        *struct{} `json:"forked_from_project"`
}

type gitlabMR struct {
	IID          int       `json:"iid"`
	Title        string    `json:"title"`
	Draft        bool      `json:"draft"`
	WebURL       string    `json:"web_url"`
	UpdatedAt    time.Time `json:"updated_at"`
	SourceBranch string    `json:"source_branch"`
	TargetBranch string    `json:"target_branch"`
	Author       struct {
		Username string `json:"username"`
	} `json:"author"`
}

func (c *gitlabClient) project(repo string) string {
	return c.base + "/projects/" + url.PathEscape(repo)
}

func (c *gitlabClient) CurrentUser(ctx context.Context) (string, error) {
	var u struct {
		Username string `json:"username"`
	}
	_, err := c.http.get(ctx, c.base+"/user", &u)
	return u.Username, err
}

func (c *gitlabClient) Repositories(ctx context.Context) ([]RemoteRepo, error) {
	var all []RemoteRepo
	for page := 1; page <= maxPages; page++ {
		var batch []gitlabProject
		u := fmt.Sprintf("%s/projects?membership=true&per_page=100&order_by=last_activity_at&page=%d", c.base, page)
		if _, err := c.http.get(ctx, u, &batch); err != nil {
			return all, err
		}
		for _, p := range batch {
			all = append(all, RemoteRepo{
				FullName: p.PathWithNamespace, Description: p.Description, CloneHTTPS: p.HTTPURL,
				CloneSSH: p.SSHURL, WebURL: p.WebURL, DefaultBranch: p.DefaultBranch,
				Private: p.Visibility != "public", Fork: p.ForkedFrom != nil, Stars: p.Stars, Updated: p.LastActivity,
			})
		}
		if len(batch) < 100 {
			break
		}
	}
	return all, nil
}

func (c *gitlabClient) PullRequests(ctx context.Context, repo string) ([]PullRequest, error) {
	var all []PullRequest
	for page := 1; page <= maxPages; page++ {
		var batch []gitlabMR
		u := fmt.Sprintf("%s/merge_requests?state=opened&per_page=100&page=%d", c.project(repo), page)
		if _, err := c.http.get(ctx, u, &batch); err != nil {
			return all, err
		}
		for _, m := range batch {
			all = append(all, PullRequest{
				Number: m.IID, Title: m.Title, Author: m.Author.Username, Source: m.SourceBranch,
				Target: m.TargetBranch, Draft: m.Draft, Updated: m.UpdatedAt, WebURL: m.WebURL,
			})
		}
		if len(batch) < 100 {
			break
		}
	}
	return all, nil
}

func (c *gitlabClient) ClosePullRequest(ctx context.Context, repo string, number int) error {
	u := fmt.Sprintf("%s/merge_requests/%d", c.project(repo), number)
	_, err := c.http.do(ctx, http.MethodPut, u, map[string]string{"state_event": "close"}, nil)
	return err
}

func (c *gitlabClient) MergePullRequest(ctx context.Context, repo string, number int) error {
	u := fmt.Sprintf("%s/merge_requests/%d/merge", c.project(repo), number)
	_, err := c.http.do(ctx, http.MethodPut, u, map[string]string{}, nil)
	return err
}
