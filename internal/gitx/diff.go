package gitx

import (
	"context"
	"strings"
)

// DiffLineKind tells added, removed, context and header lines apart.
type DiffLineKind int

// Kinds of diff line.
const (
	LineContext DiffLineKind = iota
	LineAdded
	LineRemoved
	LineHunk
	LineMeta
)

// DiffLine is one line of a unified diff.
type DiffLine struct {
	Kind DiffLineKind
	Text string
}

// FileDiff returns the diff of one path: staged against HEAD, or the
// work tree against the index. Untracked files show as fully added.
func FileDiff(ctx context.Context, dir string, c FileChange, staged bool) ([]DiffLine, error) {
	var args []string
	switch {
	case c.Kind == Untracked:
		args = []string{"diff", "--no-color", "--no-index", "--", nullDevice, c.Path}
	case staged:
		args = []string{"diff", "--no-color", "--cached", "--", c.Path}
	default:
		args = []string{"diff", "--no-color", "--", c.Path}
	}
	out, err := run(ctx, dir, args...)
	if err != nil && c.Kind != Untracked {
		return nil, err
	}
	return ParseDiff(out), nil
}

// CommitDiff returns the stat summary and patch of a commit.
func CommitDiff(ctx context.Context, dir, hash string) ([]DiffLine, error) {
	out, err := run(ctx, dir, "show", "--no-color", "--format=", "--stat", "--patch", hash)
	if err != nil {
		return nil, err
	}
	return ParseDiff(out), nil
}

// CommitBody returns the full message of a commit.
func CommitBody(ctx context.Context, dir, hash string) (string, error) {
	out, err := run(ctx, dir, "show", "-s", "--format=%B", hash)
	return strings.TrimSpace(out), err
}

// ParseDiff splits unified diff output into classified lines.
func ParseDiff(out string) []DiffLine {
	var lines []DiffLine
	for _, l := range strings.Split(strings.TrimRight(out, "\n"), "\n") {
		kind := LineContext
		switch {
		case strings.HasPrefix(l, "@@"):
			kind = LineHunk
		case strings.HasPrefix(l, "+++"), strings.HasPrefix(l, "---"),
			strings.HasPrefix(l, "diff "), strings.HasPrefix(l, "index "),
			strings.HasPrefix(l, "new file"), strings.HasPrefix(l, "deleted file"):
			kind = LineMeta
		case strings.HasPrefix(l, "+"):
			kind = LineAdded
		case strings.HasPrefix(l, "-"):
			kind = LineRemoved
		}
		lines = append(lines, DiffLine{Kind: kind, Text: l})
	}
	return lines
}
