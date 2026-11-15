package forge

import (
	"context"
	"fmt"
	"net/http"
)

func (c *giteaClient) ReleaseCaps() ReleaseCaps {
	return ReleaseCaps{Drafts: true, Prereleases: true}
}

func (c *giteaClient) Releases(ctx context.Context, repo string) ([]Release, error) {
	var out []Release
	for page := 1; page <= maxPages; page++ {
		var list []githubRelease
		if _, err := c.http.get(ctx, fmt.Sprintf("%s/repos/%s/releases?limit=50&page=%d", c.base, repo, page), &list); err != nil {
			return out, err
		}
		for _, r := range list {
			out = append(out, r.release())
		}
		if len(list) < 50 {
			break
		}
	}
	return sortReleases(out), nil
}

func (c *giteaClient) CreateRelease(ctx context.Context, repo string, r NewRelease) error {
	body := map[string]any{"tag_name": r.Tag, "name": r.Name, "body": r.Notes, "draft": r.Draft, "prerelease": r.Prerelease}
	if r.Target != "" {
		body["target_commitish"] = r.Target
	}
	_, err := c.http.do(ctx, http.MethodPost, fmt.Sprintf("%s/repos/%s/releases", c.base, repo), body, nil)
	return err
}

func (c *giteaClient) DeleteRelease(ctx context.Context, repo, id string) error {
	_, err := c.http.do(ctx, http.MethodDelete, fmt.Sprintf("%s/repos/%s/releases/%s", c.base, repo, id), nil, nil)
	return err
}

func (c *giteaClient) PublishRelease(ctx context.Context, repo, id string) error {
	url := fmt.Sprintf("%s/repos/%s/releases/%s", c.base, repo, id)
	_, err := c.http.do(ctx, http.MethodPatch, url, map[string]bool{"draft": false}, nil)
	return err
}
