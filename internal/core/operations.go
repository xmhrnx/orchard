package core

import (
	"context"
	"fmt"
	"time"

	"orchard/internal/batch"
	"orchard/internal/gitx"
)

// Operation is something Orchard can do to many repositories at once.
type Operation int

// Bulk operations.
const (
	OpFetch Operation = iota
	OpPull
	OpPush
)

// Result summarises a bulk operation.
type Result struct {
	Label  string
	Done   int
	Failed []RepoError
}

// RepoError is one repository's failure in a bulk operation.
type RepoError struct {
	Name string
	Err  error
}

// Summary is a one-line description of the result.
func (r Result) Summary() string {
	if len(r.Failed) == 0 {
		return fmt.Sprintf("%s: %d %s", r.Label, r.Done, plural(r.Done, "repository", "repositories"))
	}
	if len(r.Failed) == 1 {
		return fmt.Sprintf("%s: %d ok, %s failed: %v", r.Label, r.Done, r.Failed[0].Name, r.Failed[0].Err)
	}
	return fmt.Sprintf("%s: %d ok, %d failed", r.Label, r.Done, len(r.Failed))
}

// Label is the verb shown while an operation runs.
func (op Operation) Label() string {
	return [...]string{"Fetching", "Pulling", "Pushing"}[op]
}

func (op Operation) title() string {
	return [...]string{"Fetched", "Updated", "Pushed"}[op]
}

// Run performs op on the repositories with the given IDs, a few at a
// time, refreshing each one's status as it finishes. done runs on the UI
// goroutine once all are finished.
func (w *Workspace) Run(op Operation, ids []string, done func(Result)) {
	settings := w.state.Settings
	fn := func(ctx context.Context, path string) error {
		switch op {
		case OpFetch:
			return gitx.Fetch(ctx, path)
		case OpPull:
			if err := gitx.Fetch(ctx, path); err != nil {
				return err
			}
			return gitx.Pull(ctx, path, settings.PullStrategy, settings.AutoStash)
		default:
			return gitx.Push(ctx, path, false)
		}
	}
	w.RunEach(op.title(), op.Label(), ids, fn, done)
}

// RunEach runs any per-repository function over ids with the bulk
// machinery: busy markers, bounded parallelism, status refresh and a
// summary.
func (w *Workspace) RunEach(title, busy string, ids []string, fn func(ctx context.Context, path string) error, done func(Result)) {
	type job struct{ id, name, path string }
	var jobs []job
	for _, id := range ids {
		r := w.repos[id]
		if r == nil || r.Busy != "" {
			continue
		}
		r.Busy, r.Err = busy, nil
		jobs = append(jobs, job{id, r.Name, r.Path})
	}
	if len(jobs) == 0 {
		return
	}
	workers := w.state.Settings.Concurrency
	go func() {
		res := Result{Label: title}
		batch.Each(w.ctx, jobs, workers, func(ctx context.Context, j job) error {
			ctx, cancel := context.WithTimeout(ctx, 10*time.Minute)
			defer cancel()
			return fn(ctx, j.path)
		}, func(j job, err error) {
			w.post(func() {
				if err != nil {
					res.Failed = append(res.Failed, RepoError{Name: j.name, Err: err})
				} else {
					res.Done++
				}
				if r := w.repos[j.id]; r != nil {
					r.Busy, r.Err = "", err
					if err == nil {
						r.Synced = time.Now()
					}
				}
				w.Refresh(j.id)
			})
		})
		w.post(func() {
			if done != nil {
				done(res)
			}
		})
	}()
}

// Busy reports how many repositories have an operation running.
func (w *Workspace) Busy() int {
	n := 0
	for _, r := range w.repos {
		if r.Busy != "" {
			n++
		}
	}
	return n
}

func plural(n int, one, many string) string {
	if n == 1 {
		return one
	}
	return many
}
