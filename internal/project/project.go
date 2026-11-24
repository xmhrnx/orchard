// Package project recognises what a repository is built with and knows
// the commands that run and build it: Giode apps, npm (or pnpm, yarn,
// bun) projects and plain Go modules.
package project

import (
	"os"
	"path/filepath"
)

// Kind is a project's toolchain.
type Kind string

// Supported toolchains, in the order detection tries them.
const (
	Giode Kind = "giode"
	Node  Kind = "node"
	Go    Kind = "go"
)

// Command is one program invocation of a task.
type Command struct {
	Args []string
}

// Project is what Detect found in a folder.
type Project struct {
	Kind Kind
	Dir  string
	// Node projects: the package manager and the scripts that exist.
	Manager string
	Scripts map[string]string
	// NeedsInstall is set for Node projects without node_modules.
	NeedsInstall bool
}

// Detect looks for a giode.toml, then a package.json, then a go.mod.
// ok is false when the folder is none of these.
func Detect(dir string) (p Project, ok bool, err error) {
	switch {
	case exists(dir, "giode.toml"):
		return Project{Kind: Giode, Dir: dir}, true, nil
	case exists(dir, "package.json"):
		return detectNode(dir)
	case exists(dir, "go.mod"):
		return Project{Kind: Go, Dir: dir}, true, nil
	}
	return Project{}, false, nil
}

// Title is the toolchain's display name.
func (k Kind) Title() string {
	switch k {
	case Giode:
		return "Giode"
	case Node:
		return "Node"
	case Go:
		return "Go"
	}
	return string(k)
}

func exists(dir, name string) bool {
	_, err := os.Stat(filepath.Join(dir, name))
	return err == nil
}
