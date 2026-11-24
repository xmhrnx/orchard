//go:build windows

package tasks

import (
	"os/exec"
	"strconv"

	"orchard/internal/tools"
)

// isolate hides the console window a command would otherwise open.
func isolate(cmd *exec.Cmd) { tools.HideWindow(cmd) }

// killTree stops the command and every process it started.
func killTree(cmd *exec.Cmd) {
	if cmd.Process == nil {
		return
	}
	kill := exec.Command("taskkill", "/T", "/F", "/PID", strconv.Itoa(cmd.Process.Pid))
	tools.HideWindow(kill)
	if kill.Run() != nil {
		cmd.Process.Kill()
	}
}
