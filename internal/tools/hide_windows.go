//go:build windows

package tools

import (
	"os/exec"
	"syscall"
)

// createNoWindow is CREATE_NO_WINDOW: the process and the programs it
// starts get a console that is never shown.
const createNoWindow = 0x08000000

// HideWindow keeps a background command from flashing a console window
// on Windows. Every command Orchard runs out of sight must use it.
func HideWindow(cmd *exec.Cmd) {
	if cmd.SysProcAttr == nil {
		cmd.SysProcAttr = &syscall.SysProcAttr{}
	}
	cmd.SysProcAttr.HideWindow = true
	cmd.SysProcAttr.CreationFlags |= createNoWindow
}
