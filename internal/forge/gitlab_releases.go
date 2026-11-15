package forge

import (
	"context"
	"fmt"
	"net/http"
	"net/url"
	"time"
)

type gitlabRelease struct {
	TagName     string    `json:"tag_name"`
	Name        string    `json:"name"`
	Description string    `json:"description"`
	CreatedAt   time.Time `json:"created_at"`
	ReleasedAt  time.Time `json:"released_at"`
	Upcoming    bool      `json:"upcoming_release"`
	Author      struct {
		Username string `json:"username"`
	} `json:"author"`
	Commit struct {
		ID string `json:"id"`
	} `json:"commit"`
	Links struct {
		Self string `json:"self"`
	} `json:"_links"`
	Assets struct {
		Links []struct {
			Name string `json:"name"`
			URL  string `json:"url"`
		} `json:"links"`
	} `json:"assets"`
}

// GitLab releases are keyed by tag and have no drafts; an upcoming
// release (dated in the future) is shown as a pre-release.
func (c *gitlabClient) ReleaseCaps() ReleaseCaps { return ReleaseCaps{} }

func (c *gitlabClient) Releases(ctx context.Context, repo string) ([]Release, error) {
	var list []gitlabRelease
	if _, err := c.http.get(ctx, c.project(repo)+"/releases?per_page=100", &list); err != nil {
		return nil, err
	}
	out := make([]Release, len(list))
	for i, r := range list {
		rel := Release{
			ID: r.TagName, Tag: r.TagName, Name: r.Name, Notes: r.Description, Author: r.Author.Username,
			Target: r.Commit.ID, WebURL: r.Links.Self, Prerelease: r.Upcoming, Created: r.CreatedAt, Published: r.ReleasedAt,
		}
		for _, l := range r.Assets.Links {
			rel.Assets = append(rel.Assets, ReleaseAsset{Name: l.Name, URL: l.URL})
		}
		out[i] = rel
	}
	return sortReleases(out), nil
}

func (c *gitlabClient) CreateRelease(ctx context.Context, repo string, r NewRelease) error {
	body := map[string]string{"tag_name": r.Tag, "name": r.Name, "description": r.Notes}
	if r.Target != "" {
		body["ref"] = r.Target
	}
	_, err := c.http.do(ctx, http.MethodPost, c.project(repo)+"/releases", body, nil)
	return err
}

func (c *gitlabClient) DeleteRelease(ctx context.Context, repo, id string) error {
	_, err := c.http.do(ctx, http.MethodDelete, fmt.Sprintf("%s/releases/%s", c.project(repo), url.PathEscape(id)), nil, nil)
	return err
}

func (c *gitlabClient) PublishRelease(context.Context, string, string) error { return errNoDrafts }
