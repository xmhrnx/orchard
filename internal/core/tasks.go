package core

import (
	"context"
	"time"
)

// Do runs one mutation on one repository: it marks the repository busy,
// runs fn off the UI goroutine, refreshes the status and calls done on
// the UI goroutine.
func (w *Workspace) Do(id, busy string, fn func(ctx context.Context, path string) error, done func(error)) {
	r := w.repos[id]
	if r == nil {
		return
	}
	r.Busy = busy
	path := r.Path
	go func() {
		ctx, cancel := context.WithTimeout(w.ctx, 30*time.Minute)
		defer cancel()
		err := fn(ctx, path)
		w.post(func() {
			if r := w.repos[id]; r != nil {
				r.Busy = ""
			}
			w.Refresh(id)
			if done != nil {
				done(err)
			}
		})
	}()
}

// Query runs a read off the UI goroutine and hands the result to done on
// it.
func Query[T any](w *Workspace, fn func(ctx context.Context) (T, error), done func(T, error)) {
	go func() {
		ctx, cancel := context.WithTimeout(w.ctx, 2*time.Minute)
		defer cancel()
		v, err := fn(ctx)
		w.post(func() { done(v, err) })
	}()
}
