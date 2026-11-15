package gitx

import "context"

// CommitOptions configures Commit.
type CommitOptions struct {
	Message string
	Amend   bool
	All     bool
}

// CreateCommit records the index as a new commit.
func CreateCommit(ctx context.Context, dir string, opts CommitOptions) error {
	args := []string{"commit", "-F", "-"}
	if opts.Amend {
		args = append(args, "--amend")
	}
	if opts.All {
		args = append(args, "--all")
	}
	_, err := runInput(ctx, dir, opts.Message, args...)
	return err
}

// UndoLastCommit moves HEAD back one commit, keeping its changes staged.
func UndoLastCommit(ctx context.Context, dir string) error {
	_, err := run(ctx, dir, "reset", "--soft", "HEAD~1")
	return err
}

// Revert creates a commit undoing hash.
func Revert(ctx context.Context, dir, hash string) error {
	_, err := run(ctx, dir, "revert", "--no-edit", hash)
	return err
}

// CherryPick applies hash on top of the current branch.
func CherryPick(ctx context.Context, dir, hash string) error {
	_, err := run(ctx, dir, "cherry-pick", hash)
	return err
}
