// Package forge talks to git hosting services (GitHub, GitLab, Gitea and
// Forgejo, Bitbucket) for what plain git can't do: listing an account's
// repositories and working with pull and merge requests.
package forge

import (
	"context"
	"fmt"
	"time"
)

// Kind is a hosting service's API family.
type Kind string

// Supported services. Gitea also covers Forgejo and Codeberg.
const (
	GitHub    Kind = "github"
	GitLab    Kind = "gitlab"
	Gitea     Kind = "gitea"
	Bitbucket Kind = "bitbucket"
)

// Kinds lists the services in the order the UI offers them.
var Kinds = []Kind{GitHub, GitLab, Gitea, Bitbucket}

// Account is a signed-in identity on a service. The token lives in the
// system's secure store, never in the settings file.
type Account struct {
	ID       string `json:"id"`
	Kind     Kind   `json:"kind"`
	Host     string `json:"host"`
	Username string `json:"username"`
}

// RemoteRepo is a repository the account can see.
type RemoteRepo struct {
	FullName      string
	Description   string
	CloneHTTPS    string
	CloneSSH      string
	WebURL        string
	DefaultBranch string
	Private       bool
	Fork          bool
	Stars         int
	Updated       time.Time
}

// PullRequest is an open pull request (merge request on GitLab).
type PullRequest struct {
	Number  int
	Title   string
	Author  string
	Source  string
	Target  string
	Draft   bool
	Updated time.Time
	WebURL  string
}

// Client is the API of one service for one account.
type Client interface {
	// CurrentUser returns the login the token belongs to.
	CurrentUser(ctx context.Context) (string, error)
	// Repositories lists repositories the user owns or belongs to.
	Repositories(ctx context.Context) ([]RemoteRepo, error)
	// PullRequests lists open pull requests of a repository ("owner/name").
	PullRequests(ctx context.Context, repo string) ([]PullRequest, error)
	// ClosePullRequest closes (declines) a pull request without merging.
	ClosePullRequest(ctx context.Context, repo string, number int) error
	// MergePullRequest merges a pull request.
	MergePullRequest(ctx context.Context, repo string, number int) error
}

// New returns the client for an account.
func New(a Account, token string) (Client, error) {
	h := newHTTP(token)
	switch a.Kind {
	case GitHub:
		return &githubClient{http: h, base: githubAPI(a.Host)}, nil
	case GitLab:
		return &gitlabClient{http: h, base: "https://" + a.Host + "/api/v4"}, nil
	case Gitea:
		return &giteaClient{http: h, base: "https://" + a.Host + "/api/v1"}, nil
	case Bitbucket:
		return &bitbucketClient{http: h.withBasic(a.Username), base: "https://api.bitbucket.org/2.0"}, nil
	}
	return nil, fmt.Errorf("unsupported service %q", a.Kind)
}

// DefaultHost is the public host of a service, prefilled when adding one.
func DefaultHost(k Kind) string {
	switch k {
	case GitHub:
		return "github.com"
	case GitLab:
		return "gitlab.com"
	case Gitea:
		return "codeberg.org"
	case Bitbucket:
		return "bitbucket.org"
	}
	return ""
}

// Title is the service's display name.
func (k Kind) Title() string {
	switch k {
	case GitHub:
		return "GitHub"
	case GitLab:
		return "GitLab"
	case Gitea:
		return "Gitea / Forgejo"
	case Bitbucket:
		return "Bitbucket"
	}
	return string(k)
}

// GitUser is the user name to pair with the token for HTTPS git access.
func (a Account) GitUser() string {
	switch a.Kind {
	case GitHub:
		return "x-access-token"
	case GitLab:
		return "oauth2"
	}
	return a.Username
}

// TokenKey is the secure-store entry holding the account's token.
func (a Account) TokenKey() string { return "token:" + a.ID }

func githubAPI(host string) string {
	if host == "" || host == "github.com" {
		return "https://api.github.com"
	}
	return "https://" + host + "/api/v3"
}
