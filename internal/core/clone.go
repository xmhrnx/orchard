package core

import (
	"errors"
	"os"
	"path/filepath"

	"orchard/internal/forge"
	"orchard/internal/gitx"
)

// CloneRequest describes a clone started from the UI.
type CloneRequest struct {
	URL       string
	ParentDir string
	Folder    string
	Branch    string
	Shallow   bool
	Recursive bool
	Group     string
	Account   *forge.Account
}

// Dest is the folder the repository will be cloned into.
func (c CloneRequest) Dest() string {
	name := c.Folder
	if name == "" {
		name = gitx.RepoNameFromURL(c.URL)
	}
	return filepath.Join(c.ParentDir, name)
}

// Clone clones a repository and starts tracking it. progress and done
// run on the UI goroutine.
func (w *Workspace) Clone(req CloneRequest, progress func(gitx.Progress), done func(error)) {
	dest := req.Dest()
	if _, err := os.Stat(dest); err == nil {
		done(errors.New(dest + " already exists"))
		return
	}
	go func() {
		opts := gitx.CloneOptions{URL: req.URL, Dest: dest, Branch: req.Branch, Recursive: req.Recursive}
		if req.Shallow {
			opts.Depth = 1
		}
		if req.Account != nil {
			if token, err := w.vault.Get(req.Account.TokenKey()); err == nil {
				opts.AuthUser, opts.AuthToken = req.Account.GitUser(), string(token)
			}
		}
		err := gitx.Clone(w.ctx, opts, func(p gitx.Progress) {
			w.post(func() { progress(p) })
		})
		w.post(func() {
			if err == nil {
				id := repoID(dest)
				w.track([]string{dest})
				if r := w.repos[id]; r != nil && req.Group != "" {
					r.Group = req.Group
				}
				err = w.save()
			}
			done(err)
		})
	}()
}
