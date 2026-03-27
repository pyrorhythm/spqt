package reactive

import (
	"sync"
)

type List[T any] struct {
	mu       sync.RWMutex
	items    []T
	nextID   uint64
	listeners map[uint64]func()
}

func NewList[T any]() *List[T] {
	return &List[T]{
		listeners: make(map[uint64]func()),
	}
}

func (l *List[T]) Items() []T {
	l.mu.RLock()
	defer l.mu.RUnlock()
	return l.items
}

func (l *List[T]) Len() int {
	l.mu.RLock()
	defer l.mu.RUnlock()
	return len(l.items)
}

func (l *List[T]) At(i int) T {
	l.mu.RLock()
	defer l.mu.RUnlock()
	return l.items[i]
}

func (l *List[T]) Append(v T) {
	l.mu.Lock()
	l.items = append(l.items, v)
	l.mu.Unlock()
	l.notify()
}

func (l *List[T]) AppendAll(vs []T) {
	l.mu.Lock()
	l.items = append(l.items, vs...)
	l.mu.Unlock()
	l.notify()
}

func (l *List[T]) Remove(i int) {
	l.mu.Lock()
	l.items = append(l.items[:i], l.items[i+1:]...)
	l.mu.Unlock()
	l.notify()
}

func (l *List[T]) Set(i int, v T) {
	l.mu.Lock()
	l.items[i] = v
	l.mu.Unlock()
	l.notify()
}

func (l *List[T]) Clear() {
	l.mu.Lock()
	l.items = l.items[:0]
	l.mu.Unlock()
	l.notify()
}

func (l *List[T]) SetItems(items []T) {
	l.mu.Lock()
	l.items = items
	l.mu.Unlock()
	l.notify()
}

func (l *List[T]) OnChange(fn func()) func() {
	id := nextIDUint64()
	l.mu.Lock()
	l.listeners[id] = fn
	l.mu.Unlock()
	return func() {
		l.mu.Lock()
		delete(l.listeners, id)
		l.mu.Unlock()
	}
}

func (l *List[T]) notify() {
	l.mu.RLock()
	fns := make([]func(), 0, len(l.listeners))
	for _, fn := range l.listeners {
		fns = append(fns, fn)
	}
	l.mu.RUnlock()
	for _, fn := range fns {
		fn()
	}
}
