// Package tools finds the developer tools Orchard runs (git, giode, go,
// npm, ...), also where a desktop app's PATH doesn't reach.
package tools

import (
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
)

// extraDirs are where developer tools often live but a desktop app's
// PATH may not reach: apps launched from the Dock or Start menu don't
// read the shell profile that adds them.
func extraDirs() []string {
	home, _ := os.UserHomeDir()
	dirs := []string{filepath.Join(home, "go", "bin")}
	if gopath := os.Getenv("GOPATH"); gopath != "" {
		dirs = append(dirs, filepath.Join(gopath, "bin"))
	}
	switch runtime.GOOS {
	case "darwin":
		dirs = append(dirs, "/opt/homebrew/bin", "/usr/local/bin", "/usr/local/go/bin", filepath.Join(home, ".bun", "bin"))
	case "linux":
		dirs = append(dirs, "/usr/local/go/bin", "/usr/local/bin", filepath.Join(home, ".local", "bin"), filepath.Join(home, ".bun", "bin"))
	case "windows":
		dirs = append(dirs, `C:\Program Files\Go\bin`, `C:\Program Files\nodejs`, filepath.Join(os.Getenv("APPDATA"), "npm"))
	}
	return dirs
}

// SearchPath is PATH with the usual tool folders added, for children.
func SearchPath() string {
	parts := filepath.SplitList(os.Getenv("PATH"))
	seen := map[string]bool{}
	for _, p := range parts {
		seen[p] = true
	}
	for _, d := range extraDirs() {
		if !seen[d] {
			parts = append(parts, d)
		}
	}
	return strings.Join(parts, string(os.PathListSeparator))
}

// Find locates a program on SearchPath.
func Find(name string) (string, error) {
	if p, err := exec.LookPath(name); err == nil {
		return p, nil
	}
	exts := []string{""}
	if runtime.GOOS == "windows" {
		exts = []string{".exe", ".cmd", ".bat"}
	}
	for _, dir := range filepath.SplitList(SearchPath()) {
		for _, ext := range exts {
			p := filepath.Join(dir, name+ext)
			if info, err := os.Stat(p); err == nil && !info.IsDir() {
				return p, nil
			}
		}
	}
	return "", &exec.Error{Name: name, Err: exec.ErrNotFound}
}
