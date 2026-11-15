package gitx

import (
	"context"
	"strings"
)

// Remote is a named remote repository.
type Remote struct {
	Name     string
	FetchURL string
	PushURL  string
}

// Remotes lists the configured remotes.
func Remotes(ctx context.Context, dir string) ([]Remote, error) {
	out, err := run(ctx, dir, "remote", "-v")
	if err != nil {
		return nil, err
	}
	byName := map[string]*Remote{}
	var order []string
	for _, line := range strings.Split(strings.TrimSpace(out), "\n") {
		f := strings.Fields(line)
		if len(f) < 3 {
			continue
		}
		r, ok := byName[f[0]]
		if !ok {
			r = &Remote{Name: f[0]}
			byName[f[0]] = r
			order = append(order, f[0])
		}
		if f[2] == "(push)" {
			r.PushURL = f[1]
		} else {
			r.FetchURL = f[1]
		}
	}
	list := make([]Remote, 0, len(order))
	for _, n := range order {
		list = append(list, *byName[n])
	}
	return list, nil
}

// AddRemote adds a remote.
func AddRemote(ctx context.Context, dir, name, url string) error {
	_, err := run(ctx, dir, "remote", "add", name, url)
	return err
}

// RemoveRemote removes a remote and its tracking branches.
func RemoveRemote(ctx context.Context, dir, name string) error {
	_, err := run(ctx, dir, "remote", "remove", name)
	return err
}

// SetRemoteURL changes where a remote points.
func SetRemoteURL(ctx context.Context, dir, name, url string) error {
	_, err := run(ctx, dir, "remote", "set-url", name, url)
	return err
}

// OriginURL returns the fetch URL of origin, or of the first remote.
func OriginURL(ctx context.Context, dir string) string {
	remotes, err := Remotes(ctx, dir)
	if err != nil || len(remotes) == 0 {
		return ""
	}
	for _, r := range remotes {
		if r.Name == "origin" {
			return r.FetchURL
		}
	}
	return remotes[0].FetchURL
}
