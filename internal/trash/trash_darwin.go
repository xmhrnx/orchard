//go:build darwin

package trash

import (
	"fmt"
	"os/exec"
	"strings"
)

// move asks Finder to trash the folder, so Put Back works.
func move(abs string) error {
	path := strings.ReplaceAll(strings.ReplaceAll(abs, `\`, `\\`), `"`, `\"`)
	script := fmt.Sprintf(`tell application "Finder" to delete POSIX file "%s"`, path)
	if out, err := exec.Command("osascript", "-e", script).CombinedOutput(); err != nil {
		return fmt.Errorf("%v (%s)", err, strings.TrimSpace(string(out)))
	}
	return nil
}
