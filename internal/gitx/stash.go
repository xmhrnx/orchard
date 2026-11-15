package gitx

import (
	"context"
	"fmt"
	"strings"
	"time"
)

// Stash is a saved set of work-tree changes.
type Stash struct {
	Index   int
	Ref     string
	Message string
	When    time.Time
}

// Stashes lists the stash stack, most recent first.
func Stashes(ctx context.Context, dir string) ([]Stash, error) {
	out, err := run(ctx, dir, "stash", "list", "--format=%gd"+fieldSep+"%gs"+fieldSep+"%ct")
	if err != nil {
		return nil, err
	}
	var list []Stash
	for i, line := range strings.Split(strings.TrimSpace(out), "\n") {
		f := strings.Split(line, fieldSep)
		if len(f) < 3 {
			continue
		}
		list = append(list, Stash{Index: i, Ref: f[0], Message: f[1], When: parseUnix(f[2])})
	}
	return list, nil
}

// SaveStash stashes the work tree, untracked files included.
func SaveStash(ctx context.Context, dir, message string) error {
	args := []string{"stash", "push", "--include-untracked"}
	if message != "" {
		args = append(args, "-m", message)
	}
	_, err := run(ctx, dir, args...)
	return err
}

// ApplyStash applies a stash; pop also drops it on success.
func ApplyStash(ctx context.Context, dir string, index int, pop bool) error {
	verb := "apply"
	if pop {
		verb = "pop"
	}
	_, err := run(ctx, dir, "stash", verb, stashRef(index))
	return err
}

// DropStash discards a stash.
func DropStash(ctx context.Context, dir string, index int) error {
	_, err := run(ctx, dir, "stash", "drop", stashRef(index))
	return err
}

func stashRef(i int) string { return fmt.Sprintf("stash@{%d}", i) }
