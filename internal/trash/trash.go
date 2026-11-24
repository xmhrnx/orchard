// Package trash moves folders to the system's trash (Trash on macOS
// and Linux desktops, the Recycle Bin on Windows), so a deleted
// repository can still be restored.
package trash

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// Move puts path in the trash. It refuses filesystem roots and the
// home folder itself, which no repository should be.
func Move(path string) error {
	abs, err := filepath.Abs(path)
	if err != nil {
		return err
	}
	if err := guard(abs); err != nil {
		return err
	}
	if _, err := os.Stat(abs); err != nil {
		return err
	}
	if err := move(abs); err != nil {
		return fmt.Errorf("moving %s to the trash: %w", filepath.Base(abs), err)
	}
	return nil
}

func guard(abs string) error {
	clean := filepath.Clean(abs)
	if clean == filepath.VolumeName(clean)+string(filepath.Separator) {
		return errors.New("refusing to delete a filesystem root")
	}
	if home, err := os.UserHomeDir(); err == nil && strings.EqualFold(clean, filepath.Clean(home)) {
		return errors.New("refusing to delete the home folder")
	}
	return nil
}
