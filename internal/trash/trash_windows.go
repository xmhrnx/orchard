//go:build windows

package trash

import (
	"fmt"
	"os/exec"
	"strings"
	"syscall"
)

// move sends the folder to the Recycle Bin through the shell's own
// file operation, as Explorer does.
func move(abs string) error {
	path := strings.ReplaceAll(abs, "'", "''")
	script := "Add-Type -AssemblyName Microsoft.VisualBasic; " +
		"[Microsoft.VisualBasic.FileIO.FileSystem]::DeleteDirectory('" + path + "', 'OnlyErrorDialogs', 'SendToRecycleBin')"
	cmd := exec.Command("powershell", "-NoProfile", "-NonInteractive", "-Command", script)
	cmd.SysProcAttr = &syscall.SysProcAttr{HideWindow: true, CreationFlags: 0x08000000}
	if out, err := cmd.CombinedOutput(); err != nil {
		return fmt.Errorf("%v (%s)", err, strings.TrimSpace(string(out)))
	}
	return nil
}
