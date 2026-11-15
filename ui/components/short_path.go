package components

import (
	"os"
	"path/filepath"
	"strings"
)

// ShortPath shows a path with the home folder as "~", the way people
// read their own folders.
func ShortPath(p string) string {
	home, err := os.UserHomeDir()
	if err != nil || home == "" {
		return p
	}
	if rel, ok := strings.CutPrefix(p, home); ok && (rel == "" || rel[0] == filepath.Separator) {
		return "~" + rel
	}
	return p
}
