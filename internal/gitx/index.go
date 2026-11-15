package gitx

import (
	"context"
	"os"
	"path/filepath"
)

// Stage adds paths to the index; no paths stages everything.
func Stage(ctx context.Context, dir string, paths ...string) error {
	args := []string{"add", "--all", "--"}
	if len(paths) == 0 {
		paths = []string{"."}
	}
	_, err := run(ctx, dir, append(args, paths...)...)
	return err
}

// Unstage removes paths from the index, keeping the work-tree changes.
func Unstage(ctx context.Context, dir string, paths ...string) error {
	args := []string{"restore", "--staged", "--"}
	if len(paths) == 0 {
		paths = []string{"."}
	}
	_, err := run(ctx, dir, append(args, paths...)...)
	return err
}

// Discard throws away work-tree changes to tracked paths and deletes the
// untracked ones.
func Discard(ctx context.Context, dir string, tracked, untracked []string) error {
	if len(tracked) > 0 {
		if _, err := run(ctx, dir, append([]string{"restore", "--worktree", "--"}, tracked...)...); err != nil {
			return err
		}
	}
	for _, p := range untracked {
		if err := os.RemoveAll(filepath.Join(dir, p)); err != nil {
			return err
		}
	}
	return nil
}
