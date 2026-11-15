// Package store persists Orchard's repository list, accounts and
// settings as a JSON file in the user's configuration folder.
package store

import (
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
	"sync"
	"time"

	"orchard/internal/forge"
	"orchard/internal/gitx"
)

// Repo is a repository the user keeps track of.
type Repo struct {
	ID      string    `json:"id"`
	Path    string    `json:"path"`
	Name    string    `json:"name"`
	Group   string    `json:"group,omitempty"`
	Pinned  bool      `json:"pinned,omitempty"`
	AddedAt time.Time `json:"added_at"`
}

// Settings are the user's preferences.
type Settings struct {
	Theme            string            `json:"theme"`
	Concurrency      int               `json:"concurrency"`
	CloneDir         string            `json:"clone_dir"`
	PullStrategy     gitx.PullStrategy `json:"pull_strategy"`
	AutoStash        bool              `json:"auto_stash"`
	AutoFetchMinutes int               `json:"auto_fetch_minutes"`
	ScanDepth        int               `json:"scan_depth"`
	PreferSSH        bool              `json:"prefer_ssh"`
	// GroupBy sections the repository list: "host", "group" or "none".
	GroupBy string `json:"group_by"`
}

// State is everything Orchard saves.
type State struct {
	Repos    []Repo          `json:"repos"`
	Accounts []forge.Account `json:"accounts"`
	Settings Settings        `json:"settings"`
}

// Store reads and writes the state file. Its methods are safe from any
// goroutine.
type Store struct {
	path string
	mu   sync.Mutex
}

// Open returns the store at the default location for this user.
func Open() (*Store, error) {
	dir, err := os.UserConfigDir()
	if err != nil {
		return nil, err
	}
	return &Store{path: filepath.Join(dir, "Orchard", "state.json")}, nil
}

// OpenAt returns a store backed by a specific file.
func OpenAt(path string) *Store { return &Store{path: path} }

// Path is the file the state lives in.
func (s *Store) Path() string { return s.path }

// Load reads the state, returning defaults when nothing is saved yet.
func (s *Store) Load() (State, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	st := State{Settings: DefaultSettings()}
	data, err := os.ReadFile(s.path)
	if errors.Is(err, os.ErrNotExist) {
		return st, nil
	}
	if err != nil {
		return st, err
	}
	if err := json.Unmarshal(data, &st); err != nil {
		return State{Settings: DefaultSettings()}, err
	}
	st.Settings = st.Settings.withDefaults()
	return st, nil
}

// Save writes the state atomically, so a crash never leaves half a file.
func (s *Store) Save(st State) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	data, err := json.MarshalIndent(st, "", "  ")
	if err != nil {
		return err
	}
	if err := os.MkdirAll(filepath.Dir(s.path), 0o700); err != nil {
		return err
	}
	tmp := s.path + ".tmp"
	if err := os.WriteFile(tmp, data, 0o600); err != nil {
		return err
	}
	return os.Rename(tmp, s.path)
}
