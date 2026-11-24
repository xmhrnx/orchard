package core

import (
	"os"
	"os/exec"
	"path/filepath"
	"testing"
	"time"

	"orchard/internal/store"
)

// memVault is an in-memory Vault.
type memVault map[string][]byte

func (m memVault) Get(name string) ([]byte, error)     { return m[name], nil }
func (m memVault) Set(name string, value []byte) error { m[name] = value; return nil }
func (m memVault) Delete(name string) error            { delete(m, name); return nil }

// loop runs posted functions in order, standing in for the UI thread.
type loop chan func()

func (l loop) post(fn func()) { l <- fn }

// until runs posted functions until done reports true or time runs out.
func (l loop) until(t *testing.T, done func() bool) {
	t.Helper()
	deadline := time.After(10 * time.Second)
	for !done() {
		select {
		case fn := <-l:
			fn()
		case <-deadline:
			t.Fatal("timed out waiting for the workspace")
		}
	}
}

func TestAddPathsTracksARepositoryOnce(t *testing.T) {
	if _, err := exec.LookPath("git"); err != nil {
		t.Skip("git not installed")
	}
	repo := t.TempDir()
	if out, err := exec.Command("git", "init", "-q", repo).CombinedOutput(); err != nil {
		t.Fatalf("git init: %v %s", err, out)
	}
	sub := filepath.Join(repo, "pkg")
	if err := os.Mkdir(sub, 0o755); err != nil {
		t.Fatal(err)
	}
	l := make(loop, 64)
	st := store.OpenAt(filepath.Join(t.TempDir(), "state.json"))
	w, err := New(l.post, st, memVault{})
	if err != nil {
		t.Fatal(err)
	}
	defer w.Close()

	added := -1
	w.AddPaths([]string{sub}, func(n int, err error) {
		if err != nil {
			t.Errorf("AddPaths: %v", err)
		}
		added = n
	})
	l.until(t, func() bool { return added >= 0 })
	if added != 1 || len(w.Repos()) != 1 {
		t.Fatalf("added %d, tracking %d; want 1 and 1", added, len(w.Repos()))
	}
	if got := w.Repos()[0].Path; filepath.Base(got) != filepath.Base(repo) {
		t.Errorf("tracked %s, want the work tree root %s", got, repo)
	}

	again := -1
	w.AddPaths([]string{repo}, func(n int, _ error) { again = n })
	l.until(t, func() bool { return again >= 0 })
	if again != 0 {
		t.Errorf("re-adding added %d, want 0", again)
	}

	saved, err := st.Load()
	if err != nil || len(saved.Repos) != 1 {
		t.Fatalf("saved state = %+v, %v", saved.Repos, err)
	}
}

func TestTrashUntracksAndMovesTheFolder(t *testing.T) {
	if _, err := exec.LookPath("git"); err != nil {
		t.Skip("git not installed")
	}
	t.Setenv("XDG_DATA_HOME", t.TempDir())
	repo := filepath.Join(t.TempDir(), "doomed")
	if out, err := exec.Command("git", "init", "-q", repo).CombinedOutput(); err != nil {
		t.Fatalf("git init: %v %s", err, out)
	}
	l := make(loop, 64)
	w, err := New(l.post, store.OpenAt(filepath.Join(t.TempDir(), "state.json")), memVault{})
	if err != nil {
		t.Fatal(err)
	}
	defer w.Close()
	added := -1
	w.AddPaths([]string{repo}, func(n int, _ error) { added = n })
	l.until(t, func() bool { return added >= 0 })

	trashed := -1
	w.Trash([]string{w.Repos()[0].ID}, func(n int, err error) {
		if err != nil {
			t.Errorf("Trash: %v", err)
		}
		trashed = n
	})
	l.until(t, func() bool { return trashed >= 0 })
	if trashed != 1 || len(w.Repos()) != 0 {
		t.Fatalf("trashed %d, still tracking %d", trashed, len(w.Repos()))
	}
	if _, err := os.Stat(repo); !os.IsNotExist(err) {
		t.Error("the folder is still there")
	}
}
