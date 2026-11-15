package gitx

import (
	"context"
	"errors"
)

// ErrNoUpstream is returned when pulling a branch that tracks nothing.
var ErrNoUpstream = errors.New("the current branch has no upstream to pull from")

// PullStrategy picks how a pull integrates the upstream branch.
type PullStrategy string

// Pull strategies.
const (
	PullFastForward PullStrategy = "ff-only"
	PullRebase      PullStrategy = "rebase"
	PullMerge       PullStrategy = "merge"
)

// Fetch downloads objects and refs from every remote and prunes
// branches deleted upstream.
func Fetch(ctx context.Context, dir string) error {
	_, err := run(ctx, dir, "fetch", "--all", "--prune", "--tags", "--force")
	return err
}

// Pull integrates the current branch's upstream with the given strategy.
// With autostash, local changes are stashed and restored around it, so a
// dirty work tree doesn't block updating.
func Pull(ctx context.Context, dir string, strategy PullStrategy, autostash bool) error {
	if _, err := run(ctx, dir, "rev-parse", "--abbrev-ref", "--symbolic-full-name", "@{u}"); err != nil {
		return ErrNoUpstream
	}
	args := []string{"pull"}
	switch strategy {
	case PullRebase:
		args = append(args, "--rebase")
	case PullMerge:
		args = append(args, "--no-rebase")
	default:
		args = append(args, "--ff-only")
	}
	if autostash {
		args = append(args, "--autostash")
	}
	_, err := run(ctx, dir, args...)
	return err
}

// Push sends the current branch to its upstream, creating the upstream
// on origin when the branch has none.
func Push(ctx context.Context, dir string, force bool) error {
	args := []string{"push"}
	if force {
		args = append(args, "--force-with-lease")
	}
	if _, err := run(ctx, dir, append(args, "--porcelain")...); err != nil {
		branch, berr := CurrentBranch(ctx, dir)
		if berr != nil || branch == "" {
			return err
		}
		_, err = run(ctx, dir, append(args, "--set-upstream", "origin", branch)...)
		return err
	}
	return nil
}

// PushTags sends every local tag to origin.
func PushTags(ctx context.Context, dir string) error {
	_, err := run(ctx, dir, "push", "origin", "--tags")
	return err
}
