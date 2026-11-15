//go:build !windows

package gitx

import "os/exec"

// hideWindow is a no-op outside Windows, where git has no console window.
func hideWindow(*exec.Cmd) {}
