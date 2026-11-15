package screens

import (
	"context"
	"errors"
	"fmt"
	"sync"

	"orchard/internal/batch"
	"orchard/internal/core"
	"orchard/internal/forge"
	"orchard/ui/env"
)

// forgeBulk applies one hosting-service action to many items a few at
// a time, off the UI goroutine. done runs on the UI goroutine with how
// many succeeded and the failures, each named by name.
func forgeBulk[T any](e *env.Env, acc forge.Account, items []T, name func(T) string,
	fn func(ctx context.Context, c forge.Client, item T) error, done func(ok int, err error)) {
	core.Query(e.WS, func(ctx context.Context) (int, error) {
		client, err := e.WS.Client(acc)
		if err != nil {
			return 0, err
		}
		var (
			mu   sync.Mutex
			errs []error
		)
		batch.Each(ctx, items, 4, func(ctx context.Context, it T) error {
			return fn(ctx, client, it)
		}, func(it T, err error) {
			if err != nil {
				mu.Lock()
				errs = append(errs, fmt.Errorf("%s: %w", name(it), err))
				mu.Unlock()
			}
		})
		return len(items) - len(errs), errors.Join(errs...)
	}, done)
}
