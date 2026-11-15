package store

import (
	"path/filepath"
	"testing"
)

func TestSaveLoadRoundTrip(t *testing.T) {
	s := OpenAt(filepath.Join(t.TempDir(), "state.json"))
	st, err := s.Load()
	if err != nil {
		t.Fatal(err)
	}
	if st.Settings.Concurrency == 0 {
		t.Fatal("defaults not applied to a missing file")
	}
	st.Repos = append(st.Repos, Repo{ID: "1", Path: "/tmp/x", Name: "x"})
	if err := s.Save(st); err != nil {
		t.Fatal(err)
	}
	back, err := s.Load()
	if err != nil {
		t.Fatal(err)
	}
	if len(back.Repos) != 1 || back.Repos[0].Name != "x" {
		t.Fatalf("round trip lost repos: %+v", back.Repos)
	}
}
