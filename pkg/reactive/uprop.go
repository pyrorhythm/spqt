package reactive

import (
	"sync"
)

type Comparator[T any, K comparable] interface {
	Compare(a, b T) int
	Key(a T) K
}

type CmpProp[T any, K comparable] struct {
	mu              sync.RWMutex
	value           T
	cmp             Comparator[T, K]
	nextID          uint64
	listeners       map[uint64]func(T)
	exactNextID     uint64
	exactListeners  map[K]map[uint64]func()
}

func NewUProp[T any, K comparable](initial T, cmp Comparator[T, K]) *CmpProp[T, K] {
	return &CmpProp[T, K]{
		value:          initial,
		cmp:            cmp,
		listeners:      make(map[uint64]func(T)),
		exactListeners: make(map[K]map[uint64]func()),
	}
}

func (p *CmpProp[T, _]) Get() T {
	p.mu.RLock()
	defer p.mu.RUnlock()
	return p.value
}

func (p *CmpProp[T, _]) Set(v T) {
	p.mu.Lock()
	if p.cmp.Compare(v, p.value) == 0 { // 0 - eq
		p.mu.Unlock()
		return
	}
	p.value = v

	key := p.cmp.Key(v)
	var exactFns []func()
	if subMap := p.exactListeners[key]; len(subMap) > 0 {
		exactFns = make([]func(), 0, len(subMap))
		for _, fn := range subMap {
			exactFns = append(exactFns, fn)
		}
	}
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

func (p *CmpProp[T, K]) OnChange(fn func(T)) func() {
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

func (p *CmpProp[T, K]) OnExact(wanted T, fn func()) func() {
	id := nextIDUint64()
	p.mu.Lock()
	key := p.cmp.Key(wanted)
	if p.exactListeners[key] == nil {
		p.exactListeners[key] = make(map[uint64]func())
	}
	p.exactListeners[key][id] = fn
	p.mu.Unlock()
	return func() {
		p.mu.Lock()
		delete(p.exactListeners[key], id)
		p.mu.Unlock()
	}
}
