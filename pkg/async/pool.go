package async

import (
	"context"
	"sync"
)

// Pool is a bounded worker pool backed by a semaphore.
// Safe for concurrent use. Share one instance to cap total concurrency.
type Pool struct {
	sem chan struct{}
}

// NewPool creates a pool with the given max concurrent workers.
func NewPool(size int) *Pool {
	return &Pool{sem: make(chan struct{}, size)}
}

// Go runs fn in a goroutine, blocking until a worker slot is available.
// Returns ctx.Err() if context is cancelled while waiting.
func (p *Pool) Go(ctx context.Context, fn func()) error {
	select {
	case p.sem <- struct{}{}:
	case <-ctx.Done():
		return ctx.Err()
	}

	go func() {
		defer func() { <-p.sem }()
		fn()
	}()

	return nil
}

// Map fans out work across the pool and returns a channel of results.
// Items that return ok=false are skipped. Channel is closed when all done.
func Map[In, Out any](ctx context.Context, p *Pool, items []In, fn func(context.Context, In) (Out, bool)) <-chan Out {
	out := make(chan Out, len(items))
	var wg sync.WaitGroup

	for _, item := range items {
		wg.Add(1)
		item := item

		if err := p.Go(ctx, func() {
			defer wg.Done()
			if result, ok := fn(ctx, item); ok {
				select {
				case out <- result:
				case <-ctx.Done():
				}
			}
		}); err != nil {
			wg.Done()
			break
		}
	}

	go func() {
		wg.Wait()
		close(out)
	}()

	return out
}

// Collect drains a channel into a slice. Blocks until the channel is closed.
func Collect[T any](ch <-chan T) []T {
	var out []T
	for v := range ch {
		out = append(out, v)
	}
	return out
}
