package cache

import (
	"container/list"
	"sync"

	"github.com/dgraph-io/badger/v4"
)

// LRU is a generic in-memory LRU cache with Badger as L2 persistent store.
// Thread-safe via sync.Mutex.
type LRU[T any] struct {
	mu        sync.Mutex
	hot       map[string]*T
	order     *list.List // doubly linked list for eviction order (front = most recent)
	index     map[string]*list.Element
	maxHot    int
	store     *badger.DB
	keyPrefix string
	marshal   func(T) []byte
	unmarshal func([]byte) T
}

// NewLRU creates a new LRU cache backed by the given Badger DB.
//
//   - store:     opened Badger DB instance
//   - keyPrefix: namespaces keys written to Badger (e.g. "img:")
//   - maxHot:    maximum number of items kept in the in-memory hot cache
//   - marshal:   serialises T to bytes for Badger storage
//   - unmarshal: deserialises bytes from Badger back to T
func NewLRU[T any](
	store *badger.DB,
	keyPrefix string,
	maxHot int,
	marshal func(T) []byte,
	unmarshal func([]byte) T,
) *LRU[T] {
	return &LRU[T]{
		hot:       make(map[string]*T),
		order:     list.New(),
		index:     make(map[string]*list.Element),
		maxHot:    maxHot,
		store:     store,
		keyPrefix: keyPrefix,
		marshal:   marshal,
		unmarshal: unmarshal,
	}
}

// Get looks up key in the hot cache. On hit it promotes the entry to
// most-recently-used and returns a pointer to the value plus true.
// Returns nil, false on a miss (the caller should use GetAsync for L2).
func (c *LRU[T]) Get(key string) (*T, bool) {
	c.mu.Lock()
	defer c.mu.Unlock()

	if val, ok := c.hot[key]; ok {
		// Promote to front (most recent).
		if elem, ok := c.index[key]; ok {
			c.order.MoveToFront(elem)
		}
		return val, true
	}
	return nil, false
}

// GetAsync attempts to find key in the hot cache first. If missing, it spawns a
// goroutine to read from Badger, promotes the result into hot, and then calls cb
// with the value. If the key is not in Badger either, cb is not called.
//
// The caller is responsible for any UI thread marshaling needed before touching
// UI elements inside cb.
func (c *LRU[T]) GetAsync(key string, cb func(T)) {
	// Fast path: check hot cache under lock.
	c.mu.Lock()
	if val, ok := c.hot[key]; ok {
		if elem, ok := c.index[key]; ok {
			c.order.MoveToFront(elem)
		}
		v := *val
		c.mu.Unlock()
		cb(v)
		return
	}
	c.mu.Unlock()

	// Slow path: read from Badger in a goroutine.
	go func() {
		var result T
		var found bool

		dbKey := []byte(c.keyPrefix + key)
		err := c.store.View(func(txn *badger.Txn) error {
			item, err := txn.Get(dbKey)
			if err != nil {
				return err
			}
			return item.Value(func(val []byte) error {
				result = c.unmarshal(val)
				found = true
				return nil
			})
		})

		if err != nil || !found {
			return
		}

		// Promote into hot cache.
		c.mu.Lock()
		c.putHot(key, &result)
		c.mu.Unlock()

		cb(result)
	}()
}

// Put stores item in both the hot cache and Badger.
// If hot is full the least-recently-used entry is evicted from hot
// (it remains in Badger and can be recovered via GetAsync).
func (c *LRU[T]) Put(key string, item T) {
	// Persist to Badger (outside lock; Badger is internally concurrent).
	dbKey := []byte(c.keyPrefix + key)
	bytes := c.marshal(item)
	_ = c.store.Update(func(txn *badger.Txn) error {
		return txn.Set(dbKey, bytes)
	})

	c.mu.Lock()
	defer c.mu.Unlock()
	c.putHot(key, &item)
}

// Len returns the number of items currently in the hot cache.
func (c *LRU[T]) Len() int {
	c.mu.Lock()
	defer c.mu.Unlock()
	return len(c.hot)
}

// Clear evicts all entries from the in-memory hot cache.
// Badger L2 entries are not deleted (append-only design), but they
// will be ignored since keys are no longer indexed.
func (c *LRU[T]) Clear() {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.hot = make(map[string]*T)
	c.order = list.New()
	c.index = make(map[string]*list.Element)
}

// putHot adds or updates key/val in the hot map and order list, evicting the
// LRU entry when the hot cache is over capacity.
// Caller must hold c.mu.
func (c *LRU[T]) putHot(key string, val *T) {
	if elem, exists := c.index[key]; exists {
		// Already in hot — update value and promote.
		c.order.MoveToFront(elem)
		c.hot[key] = val
		return
	}

	// Insert new entry at front.
	elem := c.order.PushFront(key)
	c.hot[key] = val
	c.index[key] = elem

	// Evict LRU entries until within capacity.
	for len(c.hot) > c.maxHot {
		back := c.order.Back()
		if back == nil {
			break
		}
		evictKey := back.Value.(string)
		c.order.Remove(back)
		delete(c.hot, evictKey)
		delete(c.index, evictKey)
	}
}
