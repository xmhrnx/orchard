//go:build !windows

package tools

import "os/exec"

// HideWindow keeps a background command from flashing a console window
// on Windows; elsewhere there is none to hide.
func HideWindow(*exec.Cmd) {}
