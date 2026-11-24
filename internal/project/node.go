package project

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
)

// NodeScripts are the package.json scripts Orchard puts on buttons.
var NodeScripts = []string{"start", "dev", "build"}

// lockfiles name the package manager a project uses.
var lockfiles = []struct{ file, manager string }{
	{"pnpm-lock.yaml", "pnpm"},
	{"yarn.lock", "yarn"},
	{"bun.lockb", "bun"},
	{"bun.lock", "bun"},
	{"package-lock.json", "npm"},
}

func detectNode(dir string) (Project, bool, error) {
	data, err := os.ReadFile(filepath.Join(dir, "package.json"))
	if err != nil {
		return Project{}, false, err
	}
	var pkg struct {
		Scripts map[string]string `json:"scripts"`
	}
	if err := json.Unmarshal(data, &pkg); err != nil {
		return Project{}, false, fmt.Errorf("reading package.json: %w", err)
	}
	p := Project{Kind: Node, Dir: dir, Manager: "npm", Scripts: map[string]string{}}
	for _, l := range lockfiles {
		if exists(dir, l.file) {
			p.Manager = l.manager
			break
		}
	}
	for _, s := range NodeScripts {
		if cmd, ok := pkg.Scripts[s]; ok {
			p.Scripts[s] = cmd
		}
	}
	p.NeedsInstall = !exists(dir, "node_modules")
	return p, true, nil
}

// HasScript reports whether package.json defines a script.
func (p Project) HasScript(name string) bool {
	_, ok := p.Scripts[name]
	return ok
}

// NodeScript is the command running a package.json script, installing
// dependencies first when node_modules is missing.
func (p Project) NodeScript(name string) []Command {
	var steps []Command
	if p.NeedsInstall {
		steps = append(steps, Command{Args: []string{p.Manager, "install"}})
	}
	args := []string{p.Manager, "run", name}
	if p.Manager == "yarn" {
		args = []string{"yarn", name}
	}
	return append(steps, Command{Args: args})
}

// NodeInstall installs dependencies.
func (p Project) NodeInstall() []Command {
	return []Command{{Args: []string{p.Manager, "install"}}}
}
