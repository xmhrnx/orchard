package gitx

import (
	"context"
	"os"
	"path/filepath"
	"strings"
)

// TopLevel returns the root folder of the work tree containing dir.
func TopLevel(ctx context.Context, dir string) (string, error) {
	out, err := run(ctx, dir, "rev-parse", "--show-toplevel")
	if err != nil {
		return "", ErrNotARepository
	}
	return filepath.Clean(strings.TrimSpace(out)), nil
}

// IsRepository reports whether dir is the root of a git work tree.
func IsRepository(dir string) bool {
	info, err := os.Stat(filepath.Join(dir, ".git"))
	return err == nil && (info.IsDir() || info.Mode().IsRegular())
}

// Init creates an empty repository in dir.
func Init(ctx context.Context, dir string) error {
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return err
	}
	_, err := run(ctx, dir, "init")
	return err
}
