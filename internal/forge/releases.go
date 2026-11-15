package forge

import (
	"context"
	"errors"
	"sort"
	"time"
)

// Release is a published (or draft) release of a repository.
type Release struct {
	ID         string
	Tag        string
	Name       string
	Notes      string
	Author     string
	Target     string
	WebURL     string
	Draft      bool
	Prerelease bool
	Latest     bool
	Created    time.Time
	Published  time.Time
	Assets     []ReleaseAsset
}

// ReleaseAsset is a file attached to a release.
type ReleaseAsset struct {
	Name      string
	URL       string
	Size      int64
	Downloads int
}

// NewRelease describes a release to create. When Tag doesn't exist yet,
// the service creates it at Target (a branch or commit).
type NewRelease struct {
	Tag           string
	Target        string
	Name          string
	Notes         string
	Draft         bool
	Prerelease    bool
	GenerateNotes bool
}

// ReleaseCaps says which release features a service has.
type ReleaseCaps struct {
	Drafts        bool
	Prereleases   bool
	GenerateNotes bool
}

// Releases is the release API of services that have one.
type Releases interface {
	// ReleaseCaps reports the release features of the service.
	ReleaseCaps() ReleaseCaps
	// Releases lists a repository's releases, newest first.
	Releases(ctx context.Context, repo string) ([]Release, error)
	// CreateRelease publishes (or drafts) a release.
	CreateRelease(ctx context.Context, repo string, r NewRelease) error
	// DeleteRelease removes a release; its tag stays.
	DeleteRelease(ctx context.Context, repo, id string) error
	// PublishRelease turns a draft into a published release.
	PublishRelease(ctx context.Context, repo, id string) error
}

// ErrNoReleases is returned for services without releases.
var ErrNoReleases = errors.New("this service has no releases")

// errNoDrafts is returned when publishing on a service without drafts.
var errNoDrafts = errors.New("this service has no draft releases")

// ReleasesOf returns the release API of a client, if its service has one.
func ReleasesOf(c Client) (Releases, bool) {
	r, ok := c.(Releases)
	return r, ok
}

// sortReleases orders releases newest first and marks the latest: the
// newest one that is neither a draft nor a pre-release.
func sortReleases(list []Release) []Release {
	when := func(r Release) time.Time {
		if !r.Published.IsZero() {
			return r.Published
		}
		return r.Created
	}
	sort.SliceStable(list, func(i, j int) bool {
		if list[i].Draft != list[j].Draft {
			return list[i].Draft
		}
		return when(list[i]).After(when(list[j]))
	})
	for i := range list {
		if !list[i].Draft && !list[i].Prerelease {
			list[i].Latest = true
			break
		}
	}
	return list
}
