package gitx

import (
	"context"
	"strconv"
	"strings"
	"time"
)

// Commit is one entry of the history.
type Commit struct {
	Hash    string
	Parents []string
	Author  string
	Email   string
	When    time.Time
	Subject string
	Refs    []Ref
}

// RefKind tells branches, remote branches and tags apart.
type RefKind int

// Kinds of ref a commit can be decorated with.
const (
	RefBranch RefKind = iota
	RefRemote
	RefTag
	RefHead
)

// Ref is a name pointing at a commit.
type Ref struct {
	Name string
	Kind RefKind
}

// Short is the abbreviated hash.
func (c Commit) Short() string {
	if len(c.Hash) > 8 {
		return c.Hash[:8]
	}
	return c.Hash
}

// IsMerge reports whether the commit has more than one parent.
func (c Commit) IsMerge() bool { return len(c.Parents) > 1 }

// LogOptions narrows Log.
type LogOptions struct {
	Limit  int
	All    bool
	Search string
	Path   string
}

const (
	fieldSep  = "\x1f"
	recordSep = "\x1e"
)

// Log returns commits newest first in topological order, which keeps the
// lanes of a graph drawn from them stable.
func Log(ctx context.Context, dir string, opts LogOptions) ([]Commit, error) {
	format := strings.Join([]string{"%H", "%P", "%an", "%ae", "%at", "%D", "%s"}, fieldSep) + recordSep
	args := []string{"log", "--topo-order", "--decorate=full", "--format=" + format}
	if opts.Limit > 0 {
		args = append(args, "-n", strconv.Itoa(opts.Limit))
	}
	if opts.All {
		args = append(args, "--branches", "--remotes", "--tags", "HEAD")
	}
	if opts.Search != "" {
		args = append(args, "-i", "--grep="+opts.Search)
	}
	if opts.Path != "" {
		args = append(args, "--", opts.Path)
	}
	out, err := run(ctx, dir, args...)
	if err != nil {
		if strings.Contains(err.Error(), "does not have any commits") || strings.Contains(err.Error(), "bad revision") {
			return nil, nil
		}
		return nil, err
	}
	return parseLog(out), nil
}

func parseLog(out string) []Commit {
	var commits []Commit
	for _, rec := range strings.Split(out, recordSep) {
		rec = strings.TrimLeft(rec, "\n")
		f := strings.Split(rec, fieldSep)
		if len(f) < 7 {
			continue
		}
		secs, _ := strconv.ParseInt(f[4], 10, 64)
		commits = append(commits, Commit{
			Hash:    f[0],
			Parents: strings.Fields(f[1]),
			Author:  f[2],
			Email:   f[3],
			When:    time.Unix(secs, 0),
			Refs:    splitRefs(f[5]),
			Subject: f[6],
		})
	}
	return commits
}

// splitRefs reads a fully qualified decoration such as
// "HEAD -> refs/heads/main, tag: refs/tags/v1, refs/remotes/origin/main".
func splitRefs(s string) []Ref {
	if strings.TrimSpace(s) == "" {
		return nil
	}
	var refs []Ref
	for _, r := range strings.Split(s, ", ") {
		r = strings.TrimPrefix(r, "HEAD -> ")
		switch {
		case r == "HEAD":
			refs = append(refs, Ref{Name: "HEAD", Kind: RefHead})
		case strings.HasPrefix(r, "tag: "):
			refs = append(refs, Ref{Name: strings.TrimPrefix(strings.TrimPrefix(r, "tag: "), "refs/tags/"), Kind: RefTag})
		case strings.HasPrefix(r, "refs/heads/"):
			refs = append(refs, Ref{Name: strings.TrimPrefix(r, "refs/heads/"), Kind: RefBranch})
		case strings.HasPrefix(r, "refs/remotes/"):
			if name := strings.TrimPrefix(r, "refs/remotes/"); !strings.HasSuffix(name, "/HEAD") {
				refs = append(refs, Ref{Name: name, Kind: RefRemote})
			}
		}
	}
	return refs
}
