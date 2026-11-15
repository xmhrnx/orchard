package forge

import (
	"context"
	"fmt"
	"net/http"
	"time"
)

// giteaClient speaks the Gitea API, which Forgejo and Codeberg share.
type giteaClient struct {
	http *httpClient
	base string
}

type giteaRepo struct {
	FullName      string    `json:"full_name"`
	Description   string    `json:"description"`
	CloneURL      string    `json:"clone_url"`
	SSHURL        string    `json:"ssh_url"`
	HTMLURL       string    `json:"html_url"`
	DefaultBranch string    `json:"default_branch"`
	Private       bool      `json:"private"`
	Fork          bool      `json:"fork"`
	Stars         int       `json:"stars_count"`
	Updated       time.Time `json:"updated_at"`
}

type giteaPull struct {
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

func (c *giteaClient) CurrentUser(ctx context.Context) (string, error) {
	var u struct {
		Login string `json:"login"`
	}
	_, err := c.http.get(ctx, c.base+"/user", &u)
	return u.Login, err
}

func (c *giteaClient) Repositories(ctx context.Context) ([]RemoteRepo, error) {
	var all []RemoteRepo
	for page := 1; page <= maxPages; page++ {
		var batch []giteaRepo
		url := fmt.Sprintf("%s/user/repos?limit=50&page=%d", c.base, page)
		if _, err := c.http.get(ctx, url, &batch); err != nil {
			return all, err
		}
		for _, r := range batch {
			all = append(all, RemoteRepo{
				FullName: r.FullName, Description: r.Description, CloneHTTPS: r.CloneURL, CloneSSH: r.SSHURL,
				WebURL: r.HTMLURL, DefaultBranch: r.DefaultBranch, Private: r.Private, Fork: r.Fork,
				Stars: r.Stars, Updated: r.Updated,
			})
		}
		if len(batch) < 50 {
			break
		}
	}
	return all, nil
}

func (c *giteaClient) PullRequests(ctx context.Context, repo string) ([]PullRequest, error) {
	var all []PullRequest
	for page := 1; page <= maxPages; page++ {
		var batch []giteaPull
		url := fmt.Sprintf("%s/repos/%s/pulls?state=open&limit=50&page=%d", c.base, repo, page)
		if _, err := c.http.get(ctx, url, &batch); err != nil {
			return all, err
		}
		for _, p := range batch {
			all = append(all, PullRequest{
				Number: p.Number, Title: p.Title, Author: p.User.Login, Source: p.Head.Ref,
				Target: p.Base.Ref, Draft: p.Draft, Updated: p.UpdatedAt, WebURL: p.HTMLURL,
			})
		}
		if len(batch) < 50 {
			break
		}
	}
	return all, nil
}

func (c *giteaClient) ClosePullRequest(ctx context.Context, repo string, number int) error {
	url := fmt.Sprintf("%s/repos/%s/pulls/%d", c.base, repo, number)
	_, err := c.http.do(ctx, http.MethodPatch, url, map[string]string{"state": "closed"}, nil)
	return err
}

func (c *giteaClient) MergePullRequest(ctx context.Context, repo string, number int) error {
	url := fmt.Sprintf("%s/repos/%s/pulls/%d/merge", c.base, repo, number)
	_, err := c.http.do(ctx, http.MethodPost, url, map[string]string{"Do": "merge"}, nil)
	return err
}
