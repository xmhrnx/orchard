package core

import (
	"context"
	"errors"
	"fmt"
	"path/filepath"
	"time"

	"orchard/internal/gitx"
	"orchard/internal/store"
)

// AddPaths starts tracking folders. Each is resolved to the root of its
// work tree; folders that are not repositories are reported in err.
// done runs on the UI goroutine with how many were newly added.
func (w *Workspace) AddPaths(paths []string, done func(added int, err error)) {
	go func() {
		var roots []string
		var errs []error
		for _, p := range paths {
			root, err := gitx.TopLevel(w.ctx, p)
			if err != nil {
				errs = append(errs, fmt.Errorf("%s: %w", filepath.Base(p), err))
				continue
			}
			roots = append(roots, root)
		}
		w.post(func() {
			added := w.track(roots)
			err := errors.Join(errs...)
			if added > 0 {
				err = errors.Join(err, w.save())
			}
			if done != nil {
				done(added, err)
			}
		})
	}()
}

// track adds roots not tracked yet and refreshes them.
func (w *Workspace) track(roots []string) int {
	var ids []string
	for _, root := range roots {
		id := repoID(root)
		if _, ok := w.repos[id]; ok {
			continue
		}
		w.repos[id] = &RepoState{Repo: store.Repo{ID: id, Path: root, Name: filepath.Base(root), AddedAt: time.Now()}}
		ids = append(ids, id)
	}
	w.Refresh(ids...)
	return len(ids)
}

// Scan finds repositories under root without adding them.
func (w *Workspace) Scan(root string, done func(found []string, err error)) {
	depth := w.state.Settings.ScanDepth
	go func() {
		found, err := gitx.Discover(w.ctx, root, depth)
		w.post(func() {
			var fresh []string
			for _, p := range found {
				if _, ok := w.repos[repoID(p)]; !ok {
					fresh = append(fresh, p)
				}
			}
			done(fresh, err)
		})
	}()
}

// Remove stops tracking repositories. Files on disk are left alone.
func (w *Workspace) Remove(ids ...string) error {
	for _, id := range ids {
		delete(w.repos, id)
	}
	return w.save()
}

// SetPinned pins or unpins a repository to the top of the list.
func (w *Workspace) SetPinned(id string, pinned bool) error {
	r := w.repos[id]
	if r == nil {
		return nil
	}
	r.Pinned = pinned
	return w.save()
}

// SetGroup files repositories under a group; "" ungroups them.
func (w *Workspace) SetGroup(group string, ids ...string) error {
	for _, id := range ids {
		if r := w.repos[id]; r != nil {
			r.Group = group
		}
	}
	return w.save()
}

// Refresh re-reads the status of repositories; no IDs refreshes all.
func (w *Workspace) Refresh(ids ...string) {
	if len(ids) == 0 {
		for id := range w.repos {
			ids = append(ids, id)
		}
	}
	for _, id := range ids {
		r := w.repos[id]
		if r == nil {
			continue
		}
		path := r.Path
		go func() {
			ctx, cancel := context.WithTimeout(w.ctx, 30*time.Second)
			defer cancel()
			st, err := gitx.ReadStatus(ctx, path)
			origin := gitx.OriginURL(ctx, path)
			w.post(func() {
				if r := w.repos[id]; r != nil {
					r.Status, r.Origin, r.Loaded = st, origin, true
					if r.Busy == "" {
						r.Err = err
					}
				}
			})
		}()
	}
}
