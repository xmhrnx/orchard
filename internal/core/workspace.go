// Package core is Orchard's application layer: the tracked repositories
// with their live status, the operations run on them, and the accounts
// used to reach hosting services. It has no UI dependency; results come
// back through the post function, which the UI points at its own thread.
package core

import (
	"context"
	"crypto/sha1"
	"encoding/hex"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"orchard/internal/gitx"
	"orchard/internal/store"
)

// Vault keeps secrets such as access tokens.
type Vault interface {
	Get(name string) ([]byte, error)
	Set(name string, value []byte) error
	Delete(name string) error
}

// RepoState is a tracked repository and what Orchard last learned of it.
type RepoState struct {
	store.Repo
	Status gitx.Status
	Origin string
	Loaded bool
	Busy   string
	Err    error
	Synced time.Time
}

// Workspace owns the tracked repositories. Its methods must be called on
// the UI goroutine; slow work runs in goroutines that post results back.
type Workspace struct {
	ctx    context.Context
	cancel context.CancelFunc
	post   func(func())
	store  *store.Store
	vault  Vault
	state  store.State
	repos  map[string]*RepoState
}

// New loads the saved state. A state file that fails to load still
// yields a usable, empty workspace along with the error.
func New(post func(func()), st *store.Store, vault Vault) (*Workspace, error) {
	ctx, cancel := context.WithCancel(context.Background())
	w := &Workspace{ctx: ctx, cancel: cancel, post: post, store: st, vault: vault, repos: map[string]*RepoState{}}
	state, err := st.Load()
	w.state = state
	for _, r := range state.Repos {
		w.repos[r.ID] = &RepoState{Repo: r}
	}
	return w, err
}

// Close cancels running operations.
func (w *Workspace) Close() { w.cancel() }

// Repos returns the tracked repositories: pinned first, then by name.
func (w *Workspace) Repos() []*RepoState {
	list := make([]*RepoState, 0, len(w.repos))
	for _, r := range w.repos {
		list = append(list, r)
	}
	sort.Slice(list, func(i, j int) bool {
		if list[i].Pinned != list[j].Pinned {
			return list[i].Pinned
		}
		return strings.ToLower(list[i].Name) < strings.ToLower(list[j].Name)
	})
	return list
}

// Repo returns a tracked repository by ID, or nil.
func (w *Workspace) Repo(id string) *RepoState { return w.repos[id] }

// Groups returns the distinct group names in use, sorted.
func (w *Workspace) Groups() []string {
	seen := map[string]bool{}
	var groups []string
	for _, r := range w.repos {
		if r.Group != "" && !seen[r.Group] {
			seen[r.Group] = true
			groups = append(groups, r.Group)
		}
	}
	sort.Strings(groups)
	return groups
}

// Settings returns the current preferences.
func (w *Workspace) Settings() store.Settings { return w.state.Settings }

// UpdateSettings changes the preferences and saves them.
func (w *Workspace) UpdateSettings(change func(*store.Settings)) error {
	change(&w.state.Settings)
	return w.save()
}

// StatePath is where the state file lives, shown in settings.
func (w *Workspace) StatePath() string { return w.store.Path() }

// save writes the repositories and accounts to disk.
func (w *Workspace) save() error {
	w.state.Repos = w.state.Repos[:0]
	for _, r := range w.Repos() {
		w.state.Repos = append(w.state.Repos, r.Repo)
	}
	return w.store.Save(w.state)
}

// repoID derives a stable ID from a path, so adding a folder twice
// doesn't track it twice.
func repoID(path string) string {
	sum := sha1.Sum([]byte(strings.ToLower(filepath.Clean(path))))
	return hex.EncodeToString(sum[:6])
}
