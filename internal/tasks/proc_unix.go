//go:build !windows

package tasks

import (
	"os/exec"
	"syscall"
)

// isolate puts the command in its own process group, so stopping it
// also stops what it started (go run's binary, a dev server's workers).
func isolate(cmd *exec.Cmd) {
	cmd.SysProcAttr = &syscall.SysProcAttr{Setpgid: true}
}

// killTree stops the command's whole process group.
func killTree(cmd *exec.Cmd) {
	if cmd.Process == nil {
		return
	}
	if err := syscall.Kill(-cmd.Process.Pid, syscall.SIGTERM); err != nil {
		cmd.Process.Kill()
	}
}
