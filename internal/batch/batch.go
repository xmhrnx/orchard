// Package batch runs one operation over many items with bounded
// parallelism, for updating dozens of repositories at once.
package batch

import (
	"context"
	"sync"
)

// Each calls fn for every item using at most workers goroutines and
// reports each outcome to done, from the worker's goroutine. It returns
// when every item has finished or ctx is cancelled.
func Each[T any](ctx context.Context, items []T, workers int, fn func(context.Context, T) error, done func(T, error)) {
	if workers < 1 {
		workers = 1
	}
	queue := make(chan T)
	var wg sync.WaitGroup
	for range min(workers, len(items)) {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for item := range queue {
				err := ctx.Err()
				if err == nil {
					err = fn(ctx, item)
				}
				done(item, err)
			}
		}()
	}
	for _, item := range items {
		queue <- item
	}
	close(queue)
	wg.Wait()
}
