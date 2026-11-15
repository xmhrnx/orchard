package gitx

import "context"

// ResetMode picks what a reset keeps.
type ResetMode string

// Reset modes.
const (
	ResetSoft  ResetMode = "--soft"
	ResetMixed ResetMode = "--mixed"
	ResetHard  ResetMode = "--hard"
)

// Reset moves the current branch to hash.
func Reset(ctx context.Context, dir, hash string, mode ResetMode) error {
	_, err := run(ctx, dir, "reset", string(mode), hash)
	return err
}
