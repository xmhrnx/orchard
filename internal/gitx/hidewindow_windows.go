//go:build windows

package gitx

import (
	"os/exec"
	"syscall"
)

// hideWindow keeps git from flashing a console window on Windows.
func hideWindow(cmd *exec.Cmd) {
	cmd.SysProcAttr = &syscall.SysProcAttr{HideWindow: true, CreationFlags: 0x08000000}
}
