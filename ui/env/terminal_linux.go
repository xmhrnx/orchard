//go:build linux

package env

import (
	"errors"
	"os"
	"os/exec"
)

// openTerminal starts the user's terminal emulator in dir.
func openTerminal(dir string) error {
	candidates := []string{os.Getenv("TERMINAL"), "x-terminal-emulator", "gnome-terminal", "konsole", "kitty", "alacritty", "xterm"}
	for _, t := range candidates {
		if t == "" {
			continue
		}
		if path, err := exec.LookPath(t); err == nil {
			cmd := exec.Command(path)
			cmd.Dir = dir
			return cmd.Start()
		}
	}
	return errors.New("no terminal emulator found; set $TERMINAL")
}
