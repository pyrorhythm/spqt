package async

import (
	"context"
	"sync"
)

// Pool is a fixed-size worker pool. N goroutines are started at creation
// and pull work from an internal channel. Submit never blocks unless the
// queue is full — and the queue is buffered to queueSize.
type Pool struct {
	work chan func()
	wg   sync.WaitGroup
	stop context.CancelFunc
}

// NewPool creates a pool with n persistent workers and a buffered work queue.
// If queueSize <= 0 it defaults to n*4.
func NewPool(n int, queueSize int) *Pool {
	if queueSize <= 0 {
		queueSize = n * 4
	}

	ctx, cancel := context.WithCancel(context.Background())
	p := &Pool{
		work: make(chan func(), queueSize),
		stop: cancel,
	}

	p.wg.Add(n)
	for range n {
		go p.worker(ctx)
	}

	return p
}

func (p *Pool) worker(ctx context.Context) {
	defer p.wg.Done()
	for {
		select {
		case fn, ok := <-p.work:
			if !ok {
				return
			}
			fn()
		case <-ctx.Done():
			return
		}
	}
}

// Submit enqueues work. Non-blocking if queue has capacity.
// Returns ctx.Err() if context expires while waiting for a slot.
func (p *Pool) Submit(ctx context.Context, fn func()) error {
	select {
	case p.work <- fn:
		return nil
	case <-ctx.Done():
		return ctx.Err()
	}
}

// Close stops all workers after draining the queue.
// Blocks until all workers exit.
func (p *Pool) Close() {
	close(p.work)
	p.wg.Wait()
}

// Kill stops all workers immediately without draining.
func (p *Pool) Kill() {
	p.stop()
	p.wg.Wait()
}

// Map fans out work across the pool and returns a channel of results.
// Channel is closed when all items are processed.
func Map[In, Out any](ctx context.Context, p *Pool, items []In, fn func(context.Context, In) (Out, bool)) <-chan Out {
	out := make(chan Out, len(items))
	var wg sync.WaitGroup

	for _, item := range items {
		wg.Add(1)
		item := item

		if err := p.Submit(ctx, func() {
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
