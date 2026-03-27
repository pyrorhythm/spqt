package reactive

import (
	"sync"
	"sync/atomic"
)

var idCounter uint64

func nextIDUint64() uint64 {
	return atomic.AddUint64(&idCounter, 1)
}

type Signal[T any] struct {
	mu        sync.RWMutex
	nextID    uint64
	listeners map[uint64]func(T)
}

func NewSignal[T any]() *Signal[T] {
	return &Signal[T]{
		listeners: make(map[uint64]func(T)),
	}
}

func (s *Signal[T]) Emit(v T) {
	s.mu.RLock()
	fns := make([]func(T), 0, len(s.listeners))
	for _, fn := range s.listeners {
		fns = append(fns, fn)
	}
	s.mu.RUnlock()
	for _, fn := range fns {
		fn(v)
	}
}

func (s *Signal[T]) Subscribe(fn func(T)) func() {
	id := nextIDUint64()
	s.mu.Lock()
	s.listeners[id] = fn
	s.mu.Unlock()
	return func() {
		s.mu.Lock()
		delete(s.listeners, id)
		s.mu.Unlock()
	}
}

type VoidSignal struct {
	mu        sync.RWMutex
	nextID    uint64
	listeners map[uint64]func()
}

func NewVoidSignal() *VoidSignal {
	return &VoidSignal{
		listeners: make(map[uint64]func()),
	}
}

func (s *VoidSignal) Emit() {
	s.mu.RLock()
	ids := make([]uint64, 0, len(s.listeners))
	fns := make([]func(), 0, len(s.listeners))
	for id, fn := range s.listeners {
		ids = append(ids, id)
		fns = append(fns, fn)
	}
	s.mu.RUnlock()
	for _, fn := range fns {
		fn()
	}
}

func (s *VoidSignal) Subscribe(fn func()) func() {
	id := nextIDUint64()
	s.mu.Lock()
	s.listeners[id] = fn
	s.mu.Unlock()
	return func() {
		s.mu.Lock()
		delete(s.listeners, id)
		s.mu.Unlock()
	}
}
