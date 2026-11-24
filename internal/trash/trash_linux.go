//go:build linux

package trash

import (
	"fmt"
	"net/url"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"time"
)

// move follows the freedesktop.org trash specification, so the file
// manager lists the folder in its Trash and can restore it. When the
// folder is on another filesystem than the home trash, gio (part of
// every GNOME and many other desktops) handles it.
func move(abs string) error {
	if err := moveToHomeTrash(abs); err == nil {
		return nil
	} else if gio, lookErr := exec.LookPath("gio"); lookErr == nil {
		if out, gioErr := exec.Command(gio, "trash", abs).CombinedOutput(); gioErr != nil {
			return fmt.Errorf("%v (%s)", gioErr, out)
		}
		return nil
	} else {
		return err
	}
}

func moveToHomeTrash(abs string) error {
	data := os.Getenv("XDG_DATA_HOME")
	if data == "" {
		home, err := os.UserHomeDir()
		if err != nil {
			return err
		}
		data = filepath.Join(home, ".local", "share")
	}
	files, info := filepath.Join(data, "Trash", "files"), filepath.Join(data, "Trash", "info")
	for _, d := range []string{files, info} {
		if err := os.MkdirAll(d, 0o700); err != nil {
			return err
		}
	}
	name := filepath.Base(abs)
	for i := 2; exists(filepath.Join(files, name)) || exists(filepath.Join(info, name+".trashinfo")); i++ {
		name = filepath.Base(abs) + "." + strconv.Itoa(i)
	}
	record := fmt.Sprintf("[Trash Info]\nPath=%s\nDeletionDate=%s\n",
		(&url.URL{Path: abs}).EscapedPath(), time.Now().Format("2006-01-02T15:04:05"))
	infoPath := filepath.Join(info, name+".trashinfo")
	if err := os.WriteFile(infoPath, []byte(record), 0o600); err != nil {
		return err
	}
	if err := os.Rename(abs, filepath.Join(files, name)); err != nil {
		os.Remove(infoPath)
		return err
	}
	return nil
}

func exists(p string) bool {
	_, err := os.Lstat(p)
	return err == nil
}
