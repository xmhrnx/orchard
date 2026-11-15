package batch

import (
	"context"
	"errors"
	"sync"
	"sync/atomic"
	"testing"
)

func TestEachRunsEveryItemWithinTheWorkerLimit(t *testing.T) {
	var running, peak atomic.Int32
	var mu sync.Mutex
	failed := 0
	items := []int{1, 2, 3, 4, 5, 6, 7, 8}
	Each(context.Background(), items, 3, func(_ context.Context, n int) error {
		now := running.Add(1)
		for {
			p := peak.Load()
			if now <= p || peak.CompareAndSwap(p, now) {
				break
			}
		}
		defer running.Add(-1)
		if n%2 == 0 {
			return errors.New("even")
		}
		return nil
	}, func(_ int, err error) {
		mu.Lock()
		defer mu.Unlock()
		if err != nil {
			failed++
		}
	})
	if failed != 4 {
		t.Errorf("failed = %d, want 4", failed)
	}
	if peak.Load() > 3 {
		t.Errorf("peak concurrency %d exceeds 3", peak.Load())
	}
}
