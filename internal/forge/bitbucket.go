package forge

import (
	"context"
	"fmt"
	"net/http"
	"time"
)

// bitbucketClient speaks Bitbucket Cloud's 2.0 API with an app password.
type bitbucketClient struct {
	http *httpClient
	base string
}

type bitbucketLink struct {
	Href string `json:"href"`
	Name string `json:"name"`
}

type bitbucketRepo struct {
	FullName    string    `json:"full_name"`
	Description string    `json:"description"`
	IsPrivate   bool      `json:"is_private"`
	UpdatedOn   time.Time `json:"updated_on"`
	MainBranch  *struct {
		Name string `json:"name"`
	} `json:"mainbranch"`
	Parent *struct{} `json:"parent"`
	Links  struct {
		HTML  bitbucketLink   `json:"html"`
		Clone []bitbucketLink `json:"clone"`
	} `json:"links"`
}

type bitbucketPull struct {
	ID        int       `json:"id"`
	Title     string    `json:"title"`
	UpdatedOn time.Time `json:"updated_on"`
	Author    struct {
		DisplayName string `json:"display_name"`
	} `json:"author"`
	Source struct {
		Branch struct {
			Name string `json:"name"`
		} `json:"branch"`
	} `json:"source"`
	Destination struct {
		Branch struct {
			Name string `json:"name"`
		} `json:"branch"`
	} `json:"destination"`
	Links struct {
		HTML bitbucketLink `json:"html"`
	} `json:"links"`
}

type bitbucketPage[T any] struct {
	Values []T    `json:"values"`
	Next   string `json:"next"`
}

func (c *bitbucketClient) CurrentUser(ctx context.Context) (string, error) {
	var u struct {
		Username string `json:"username"`
	}
	_, err := c.http.get(ctx, c.base+"/user", &u)
	return u.Username, err
}

func (c *bitbucketClient) Repositories(ctx context.Context) ([]RemoteRepo, error) {
	var all []RemoteRepo
	next := c.base + "/repositories?role=member&pagelen=100&sort=-updated_on"
	for page := 0; next != "" && page < maxPages; page++ {
		var p bitbucketPage[bitbucketRepo]
		if _, err := c.http.get(ctx, next, &p); err != nil {
			return all, err
		}
		for _, r := range p.Values {
			repo := RemoteRepo{
				FullName: r.FullName, Description: r.Description, WebURL: r.Links.HTML.Href,
				Private: r.IsPrivate, Fork: r.Parent != nil, Updated: r.UpdatedOn,
			}
			if r.MainBranch != nil {
				repo.DefaultBranch = r.MainBranch.Name
			}
			for _, l := range r.Links.Clone {
				if l.Name == "ssh" {
					repo.CloneSSH = l.Href
				} else {
					repo.CloneHTTPS = l.Href
				}
			}
			all = append(all, repo)
		}
		next = p.Next
	}
	return all, nil
}

func (c *bitbucketClient) PullRequests(ctx context.Context, repo string) ([]PullRequest, error) {
	var all []PullRequest
	next := fmt.Sprintf("%s/repositories/%s/pullrequests?state=OPEN&pagelen=50", c.base, repo)
	for page := 0; next != "" && page < maxPages; page++ {
		var p bitbucketPage[bitbucketPull]
		if _, err := c.http.get(ctx, next, &p); err != nil {
			return all, err
		}
		for _, pr := range p.Values {
			all = append(all, PullRequest{
				Number: pr.ID, Title: pr.Title, Author: pr.Author.DisplayName, Source: pr.Source.Branch.Name,
				Target: pr.Destination.Branch.Name, Updated: pr.UpdatedOn, WebURL: pr.Links.HTML.Href,
			})
		}
		next = p.Next
	}
	return all, nil
}

func (c *bitbucketClient) ClosePullRequest(ctx context.Context, repo string, number int) error {
	url := fmt.Sprintf("%s/repositories/%s/pullrequests/%d/decline", c.base, repo, number)
	_, err := c.http.do(ctx, http.MethodPost, url, nil, nil)
	return err
}

func (c *bitbucketClient) MergePullRequest(ctx context.Context, repo string, number int) error {
	url := fmt.Sprintf("%s/repositories/%s/pullrequests/%d/merge", c.base, repo, number)
	_, err := c.http.do(ctx, http.MethodPost, url, nil, nil)
	return err
}
