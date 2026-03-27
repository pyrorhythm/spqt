package vm

import (
	"fmt"
	"sync"
	"sync/atomic"

	metadatapb "github.com/devgianlu/go-librespot/proto/spotify/metadata"

	"github.com/pyrorhythm/spqt/pkg/cache"
)

type TrackListVM struct {
	mu        sync.Mutex
	length    int
	scrollY   int
	lru       *cache.LRU[*metadatapb.Track]
	nextSubID uint64
	lenSubs   map[uint64]func(int)
	searchID  uint64 // incremented on each new search; stale goroutines check this before writing
}

func NewTrackListVM(lru *cache.LRU[*metadatapb.Track]) *TrackListVM {
	return &TrackListVM{
		lru:     lru,
		lenSubs: make(map[uint64]func(int)),
	}
}

func (vm *TrackListVM) Len() int {
	vm.mu.Lock()
	defer vm.mu.Unlock()
	return vm.length
}

func (vm *TrackListVM) Get(index int) (**metadatapb.Track, bool) {
	return vm.lru.Get(fmt.Sprintf("%d", index))
}

func (vm *TrackListVM) LoadAsync(index int, cb func(*metadatapb.Track)) {
	vm.lru.GetAsync(fmt.Sprintf("%d", index), cb)
}

func (vm *TrackListVM) OnLengthChanged(cb func(int)) func() {
	id := atomic.AddUint64(&vm.nextSubID, 1)
	vm.mu.Lock()
	vm.lenSubs[id] = cb
	vm.mu.Unlock()

	return func() {
		vm.mu.Lock()
		delete(vm.lenSubs, id)
		vm.mu.Unlock()
	}
}

func (vm *TrackListVM) AddBatch(tracks []*metadatapb.Track) {
	vm.mu.Lock()
	start := vm.length
	for i, t := range tracks {
		if t == nil {
			continue
		}
		vm.lru.Put(fmt.Sprintf("%d", start+i), t)
	}
	vm.length += len(tracks)
	newLen := vm.length
	subs := make([]func(int), 0, len(vm.lenSubs))
	for _, cb := range vm.lenSubs {
		subs = append(subs, cb)
	}
	vm.mu.Unlock()

	for _, cb := range subs {
		cb(newLen)
	}
}

// AddBatchGuarded is like AddBatch but only writes if the provided searchID
// matches the current searchID, discarding results from superseded searches.
func (vm *TrackListVM) AddBatchGuarded(searchID uint64, tracks []*metadatapb.Track) {
	vm.mu.Lock()
	if searchID != vm.searchID {
		vm.mu.Unlock()
		return // stale, discard
	}
	start := vm.length
	for i, t := range tracks {
		if t == nil {
			continue
		}
		vm.lru.Put(fmt.Sprintf("%d", start+i), t)
	}
	vm.length += len(tracks)
	newLen := vm.length
	subs := make([]func(int), 0, len(vm.lenSubs))
	for _, cb := range vm.lenSubs {
		subs = append(subs, cb)
	}
	vm.mu.Unlock()

	for _, cb := range subs {
		cb(newLen)
	}
}

func (vm *TrackListVM) SaveScrollY(y int) {
	vm.mu.Lock()
	vm.scrollY = y
	vm.mu.Unlock()
}

func (vm *TrackListVM) LastScrollY() int {
	vm.mu.Lock()
	defer vm.mu.Unlock()
	return vm.scrollY
}

func (vm *TrackListVM) Clear() {
	vm.mu.Lock()
	vm.length = 0
	vm.searchID++ // invalidate any in-flight goroutines
	vm.lru.Clear()
	subs := make([]func(int), 0, len(vm.lenSubs))
	for _, cb := range vm.lenSubs {
		subs = append(subs, cb)
	}
	vm.mu.Unlock()

	for _, cb := range subs {
		cb(0)
	}
}

// NewSearchID increments and returns a new search ID.
// Goroutines should capture this ID and only write results if it matches CurrentSearchID.
func (vm *TrackListVM) NewSearchID() uint64 {
	vm.mu.Lock()
	defer vm.mu.Unlock()
	vm.searchID++
	return vm.searchID
}

// CurrentSearchID returns the current search ID without incrementing.
func (vm *TrackListVM) CurrentSearchID() uint64 {
	vm.mu.Lock()
	defer vm.mu.Unlock()
	return vm.searchID
}
