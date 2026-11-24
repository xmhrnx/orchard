//go:build windows

package tasks

import (
	"os/exec"
	"strconv"
	"syscall"
)

// isolate hides the console window a command would otherwise open.
func isolate(cmd *exec.Cmd) {
	cmd.SysProcAttr = &syscall.SysProcAttr{HideWindow: true, CreationFlags: 0x08000000}
}

// killTree stops the command and every process it started.
func killTree(cmd *exec.Cmd) {
	if cmd.Process == nil {
		return
	}
	kill := exec.Command("taskkill", "/T", "/F", "/PID", strconv.Itoa(cmd.Process.Pid))
	isolate(kill)
	if kill.Run() != nil {
		cmd.Process.Kill()
	}
}
