package reactive

import (
	"context"
	"sync"
	"time"
)

type Prop[T comparable] struct {
	mu             sync.RWMutex
	value          T
	nextID         uint64
	listeners      map[uint64]func(T)
	exactNextID    uint64
	exactListeners map[T]map[uint64]func()
}

func NewProp[T comparable](initial T) *Prop[T] {
	return &Prop[T]{
		value:          initial,
		listeners:      make(map[uint64]func(T)),
		exactListeners: make(map[T]map[uint64]func()),
	}
}

func (p *Prop[T]) Poll(ctx context.Context, every time.Duration, pollFn func() T) {
	go func() {
		tim := time.NewTicker(every)
		defer tim.Stop()
		for {
			select {
			case <-ctx.Done():
				return
			case <-tim.C:
				p.Set(pollFn())
			}
		}
	}()
}

func (p *Prop[T]) Get() T {
	p.mu.RLock()
	defer p.mu.RUnlock()
	return p.value
}

func (p *Prop[T]) Set(v T) {
	p.mu.Lock()
	if p.value == v {
		p.mu.Unlock()
		return
	}
	p.value = v

	// Snapshot exact listeners for the new value.
	var exactFns []func()
	if subMap := p.exactListeners[v]; len(subMap) > 0 {
		exactFns = make([]func(), 0, len(subMap))
		for _, fn := range subMap {
			exactFns = append(exactFns, fn)
		}
	}
	// Snapshot all-change listeners.
	var allFns []func(T)
	if len(p.listeners) > 0 {
		allFns = make([]func(T), 0, len(p.listeners))
		for _, fn := range p.listeners {
			allFns = append(allFns, fn)
		}
	}
	p.mu.Unlock()

	for _, efn := range exactFns {
		efn()
	}
	for _, fn := range allFns {
		fn(v)
	}
}

func (p *Prop[T]) OnChange(fn func(T)) func() {
	id := nextIDUint64()
	p.mu.Lock()
	p.listeners[id] = fn
	p.mu.Unlock()
	return func() {
		p.mu.Lock()
		delete(p.listeners, id)
		p.mu.Unlock()
	}
}

func (p *Prop[T]) OnExact(wanted T, fn func()) func() {
	id := nextIDUint64()
	p.mu.Lock()
	if p.exactListeners[wanted] == nil {
		p.exactListeners[wanted] = make(map[uint64]func())
	}
	p.exactListeners[wanted][id] = fn
	p.mu.Unlock()
	return func() {
		p.mu.Lock()
		delete(p.exactListeners[wanted], id)
		p.mu.Unlock()
	}
}
