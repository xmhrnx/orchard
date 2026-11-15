//go:build windows

package env

import "os/exec"

// openTerminal starts Windows Terminal in dir, falling back to cmd.
func openTerminal(dir string) error {
	if err := exec.Command("wt.exe", "-d", dir).Start(); err == nil {
		return nil
	}
	cmd := exec.Command("cmd.exe", "/c", "start", "cmd.exe")
	cmd.Dir = dir
	return cmd.Start()
}
