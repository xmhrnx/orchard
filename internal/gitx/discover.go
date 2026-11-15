package gitx

import (
	"context"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
)

// skippedDirs are folders that never contain repositories worth listing
// and are expensive to walk.
var skippedDirs = map[string]bool{
	"node_modules": true, "vendor": true, "Library": true, "AppData": true,
	".cache": true, "target": true, "build": true, "dist": true, ".venv": true,
}

// Discover walks root up to maxDepth folders deep and returns every work
// tree it finds. It does not descend into a repository once found, so
// nested submodules are left to their parent.
func Discover(ctx context.Context, root string, maxDepth int) ([]string, error) {
	root = filepath.Clean(root)
	baseDepth := strings.Count(root, string(os.PathSeparator))
	var found []string
	err := filepath.WalkDir(root, func(path string, d fs.DirEntry, err error) error {
		if ctx.Err() != nil {
			return ctx.Err()
		}
		if err != nil || !d.IsDir() {
			return nil
		}
		name := d.Name()
		if path != root && (skippedDirs[name] || strings.HasPrefix(name, ".")) {
			return filepath.SkipDir
		}
		if IsRepository(path) {
			found = append(found, path)
			return filepath.SkipDir
		}
		if strings.Count(path, string(os.PathSeparator))-baseDepth >= maxDepth {
			return filepath.SkipDir
		}
		return nil
	})
	return found, err
}
