package forge

import (
	"context"
	"fmt"
	"net/http"
	"strconv"
	"time"
)

// githubRelease is the release shape GitHub and Gitea share.
type githubRelease struct {
	ID          int64     `json:"id"`
	TagName     string    `json:"tag_name"`
	Name        string    `json:"name"`
	Body        string    `json:"body"`
	Draft       bool      `json:"draft"`
	Prerelease  bool      `json:"prerelease"`
	Target      string    `json:"target_commitish"`
	HTMLURL     string    `json:"html_url"`
	CreatedAt   time.Time `json:"created_at"`
	PublishedAt time.Time `json:"published_at"`
	Author      struct {
		Login string `json:"login"`
	} `json:"author"`
	Assets []struct {
		Name          string `json:"name"`
		Size          int64  `json:"size"`
		DownloadCount int    `json:"download_count"`
		DownloadURL   string `json:"browser_download_url"`
	} `json:"assets"`
}

func (r githubRelease) release() Release {
	out := Release{
		ID: strconv.FormatInt(r.ID, 10), Tag: r.TagName, Name: r.Name, Notes: r.Body, Author: r.Author.Login,
		Target: r.Target, WebURL: r.HTMLURL, Draft: r.Draft, Prerelease: r.Prerelease,
		Created: r.CreatedAt, Published: r.PublishedAt,
	}
	for _, a := range r.Assets {
		out.Assets = append(out.Assets, ReleaseAsset{Name: a.Name, URL: a.DownloadURL, Size: a.Size, Downloads: a.DownloadCount})
	}
	return out
}

func (c *githubClient) ReleaseCaps() ReleaseCaps {
	return ReleaseCaps{Drafts: true, Prereleases: true, GenerateNotes: true}
}

func (c *githubClient) Releases(ctx context.Context, repo string) ([]Release, error) {
	var list []githubRelease
	if _, err := c.http.get(ctx, fmt.Sprintf("%s/repos/%s/releases?per_page=100", c.base, repo), &list); err != nil {
		return nil, err
	}
	out := make([]Release, len(list))
	for i, r := range list {
		out[i] = r.release()
	}
	return sortReleases(out), nil
}

func (c *githubClient) CreateRelease(ctx context.Context, repo string, r NewRelease) error {
	body := map[string]any{
		"tag_name": r.Tag, "name": r.Name, "body": r.Notes, "draft": r.Draft,
		"prerelease": r.Prerelease, "generate_release_notes": r.GenerateNotes,
	}
	if r.Target != "" {
		body["target_commitish"] = r.Target
	}
	_, err := c.http.do(ctx, http.MethodPost, fmt.Sprintf("%s/repos/%s/releases", c.base, repo), body, nil)
	return err
}

func (c *githubClient) DeleteRelease(ctx context.Context, repo, id string) error {
	_, err := c.http.do(ctx, http.MethodDelete, fmt.Sprintf("%s/repos/%s/releases/%s", c.base, repo, id), nil, nil)
	return err
}

func (c *githubClient) PublishRelease(ctx context.Context, repo, id string) error {
	url := fmt.Sprintf("%s/repos/%s/releases/%s", c.base, repo, id)
	_, err := c.http.do(ctx, http.MethodPatch, url, map[string]bool{"draft": false}, nil)
	return err
}
