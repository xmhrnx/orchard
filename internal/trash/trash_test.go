//go:build linux

package trash

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestMoveFollowsTheFreedesktopTrash(t *testing.T) {
	data := t.TempDir()
	t.Setenv("XDG_DATA_HOME", data)
	repo := filepath.Join(data, "work", "my repo")
	if err := os.MkdirAll(filepath.Join(repo, ".git"), 0o755); err != nil {
		t.Fatal(err)
	}
	for i := 0; i < 2; i++ {
		if err := os.MkdirAll(repo, 0o755); err != nil {
			t.Fatal(err)
		}
		if err := Move(repo); err != nil {
			t.Fatal(err)
		}
		if _, err := os.Stat(repo); !os.IsNotExist(err) {
			t.Fatal("folder still in place")
		}
	}
	for _, name := range []string{"my repo", "my repo.2"} {
		if _, err := os.Stat(filepath.Join(data, "Trash", "files", name)); err != nil {
			t.Errorf("trashed %q missing: %v", name, err)
		}
		info, err := os.ReadFile(filepath.Join(data, "Trash", "info", name+".trashinfo"))
		if err != nil || !strings.Contains(string(info), "Path=") || !strings.Contains(string(info), "my%20repo") {
			t.Errorf("trashinfo for %q = %q, %v", name, info, err)
		}
	}
}

func TestMoveRefusesTheHomeFolderAndRoots(t *testing.T) {
	home, _ := os.UserHomeDir()
	for _, p := range []string{"/", home} {
		if err := Move(p); err == nil {
			t.Errorf("Move(%q) was allowed", p)
		}
	}
}
