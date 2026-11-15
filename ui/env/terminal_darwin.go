//go:build darwin

package env

import "os/exec"

// openTerminal opens Terminal.app in dir.
func openTerminal(dir string) error {
	return exec.Command("open", "-a", "Terminal", dir).Start()
}
