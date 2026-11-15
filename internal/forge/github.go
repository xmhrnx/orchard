package forge

import (
	"context"
	"fmt"
	"net/http"
	"time"
)

type githubClient struct {
	http *httpClient
	base string
}

type githubRepo struct {
	FullName      string    `json:"full_name"`
	Description   string    `json:"description"`
	CloneURL      string    `json:"clone_url"`
	SSHURL        string    `json:"ssh_url"`
	HTMLURL       string    `json:"html_url"`
	DefaultBranch string    `json:"default_branch"`
	Private       bool      `json:"private"`
	Fork          bool      `json:"fork"`
	Stars         int       `json:"stargazers_count"`
	PushedAt      time.Time `json:"pushed_at"`
}

type githubPull struct {
	Number    int       `json:"number"`
	Title     string    `json:"title"`
	Draft     bool      `json:"draft"`
	HTMLURL   string    `json:"html_url"`
	UpdatedAt time.Time `json:"updated_at"`
	User      struct {
		Login string `json:"login"`
	} `json:"user"`
	Head struct {
		Ref string `json:"ref"`
	} `json:"head"`
	Base struct {
		Ref string `json:"ref"`
	} `json:"base"`
}

func (c *githubClient) CurrentUser(ctx context.Context) (string, error) {
	var u struct {
		Login string `json:"login"`
	}
	_, err := c.http.get(ctx, c.base+"/user", &u)
	return u.Login, err
}

func (c *githubClient) Repositories(ctx context.Context) ([]RemoteRepo, error) {
	var all []RemoteRepo
	for page := 1; page <= maxPages; page++ {
		var batch []githubRepo
		url := fmt.Sprintf("%s/user/repos?per_page=100&sort=pushed&page=%d", c.base, page)
		if _, err := c.http.get(ctx, url, &batch); err != nil {
			return all, err
		}
		for _, r := range batch {
			all = append(all, RemoteRepo{
				FullName: r.FullName, Description: r.Description, CloneHTTPS: r.CloneURL, CloneSSH: r.SSHURL,
				WebURL: r.HTMLURL, DefaultBranch: r.DefaultBranch, Private: r.Private, Fork: r.Fork,
				Stars: r.Stars, Updated: r.PushedAt,
			})
		}
		if len(batch) < 100 {
			break
		}
	}
	return all, nil
}

func (c *githubClient) PullRequests(ctx context.Context, repo string) ([]PullRequest, error) {
	var all []PullRequest
	for page := 1; page <= maxPages; page++ {
		var batch []githubPull
		url := fmt.Sprintf("%s/repos/%s/pulls?state=open&per_page=100&page=%d", c.base, repo, page)
		if _, err := c.http.get(ctx, url, &batch); err != nil {
			return all, err
		}
		for _, p := range batch {
			all = append(all, PullRequest{
				Number: p.Number, Title: p.Title, Author: p.User.Login, Source: p.Head.Ref,
				Target: p.Base.Ref, Draft: p.Draft, Updated: p.UpdatedAt, WebURL: p.HTMLURL,
			})
		}
		if len(batch) < 100 {
			break
		}
	}
	return all, nil
}

func (c *githubClient) ClosePullRequest(ctx context.Context, repo string, number int) error {
	url := fmt.Sprintf("%s/repos/%s/pulls/%d", c.base, repo, number)
	_, err := c.http.do(ctx, http.MethodPatch, url, map[string]string{"state": "closed"}, nil)
	return err
}

func (c *githubClient) MergePullRequest(ctx context.Context, repo string, number int) error {
	url := fmt.Sprintf("%s/repos/%s/pulls/%d/merge", c.base, repo, number)
	_, err := c.http.do(ctx, http.MethodPut, url, map[string]string{}, nil)
	return err
}
