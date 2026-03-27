package cache

import (
	"fmt"
	"sync"
	"testing"
	"time"

	"github.com/dgraph-io/badger/v4"
)

// openTestDB opens a Badger DB in the given temp directory with error logging
// suppressed so tests produce clean output.
func openTestDB(t *testing.T) *badger.DB {
	t.Helper()
	dir := t.TempDir()
	db, err := badger.Open(badger.DefaultOptions(dir).WithLoggingLevel(badger.ERROR))
	if err != nil {
		t.Fatalf("failed to open badger: %v", err)
	}
	t.Cleanup(func() { db.Close() })
	return db
}

// strMarshal / strUnmarshal are simple helpers used throughout the tests.
func strMarshal(s string) []byte   { return []byte(s) }
func strUnmarshal(b []byte) string { return string(b) }

// newStrLRU builds an LRU[string] for tests.
func newStrLRU(db *badger.DB, maxHot int) *LRU[string] {
	return NewLRU(db, "test:", maxHot, strMarshal, strUnmarshal)
}

// --- 1. Put / Get -----------------------------------------------------------

func TestPutGet(t *testing.T) {
	db := openTestDB(t)
	c := newStrLRU(db, 10)

	c.Put("hello", "world")

	val, ok := c.Get("hello")
	if !ok {
		t.Fatal("expected Get to return true")
	}
	if *val != "world" {
		t.Fatalf("expected 'world', got %q", *val)
	}
}

func TestGetMiss(t *testing.T) {
	db := openTestDB(t)
	c := newStrLRU(db, 10)

	val, ok := c.Get("missing")
	if ok || val != nil {
		t.Fatal("expected miss for unknown key")
	}
}

// --- 2. Eviction ------------------------------------------------------------

func TestEviction(t *testing.T) {
	db := openTestDB(t)
	c := newStrLRU(db, 3)

	// Insert 3 items (fills hot cache).
	c.Put("a", "alpha")
	c.Put("b", "beta")
	c.Put("c", "gamma")

	if c.Len() != 3 {
		t.Fatalf("expected Len 3, got %d", c.Len())
	}

	// "a" is the LRU entry; inserting "d" should evict it.
	c.Put("d", "delta")

	if c.Len() != 3 {
		t.Fatalf("expected Len 3 after eviction, got %d", c.Len())
	}

	_, ok := c.Get("a")
	if ok {
		t.Fatal("expected 'a' to be evicted from hot cache")
	}

	// "d" should be present.
	val, ok := c.Get("d")
	if !ok || *val != "delta" {
		t.Fatalf("expected 'd' in hot cache")
	}
}

// --- 3. Promote on Get ------------------------------------------------------

func TestPromoteOnGet(t *testing.T) {
	db := openTestDB(t)
	c := newStrLRU(db, 3)

	c.Put("a", "alpha")
	c.Put("b", "beta")
	c.Put("c", "gamma")

	// Access "a" to promote it to MRU.
	c.Get("a")

	// Insert "d" — now "b" should be the LRU and get evicted (not "a").
	c.Put("d", "delta")

	_, aOk := c.Get("a")
	_, bOk := c.Get("b")

	if !aOk {
		t.Fatal("'a' should still be in hot cache after being promoted")
	}
	if bOk {
		t.Fatal("'b' should have been evicted as the LRU entry")
	}
}

// --- 4. GetAsync recovers evicted item from Badger --------------------------

func TestGetAsyncRecovery(t *testing.T) {
	db := openTestDB(t)
	c := newStrLRU(db, 2)

	// Fill hot cache to capacity.
	c.Put("x", "xray")
	c.Put("y", "yankee")
	// "x" is now LRU.

	// Inserting "z" evicts "x" from hot — but it was persisted to Badger.
	c.Put("z", "zulu")

	if _, ok := c.Get("x"); ok {
		t.Fatal("'x' should have been evicted from hot cache")
	}

	// GetAsync should find "x" in Badger and call the callback.
	done := make(chan string, 1)
	c.GetAsync("x", func(v string) {
		done <- v
	})

	select {
	case got := <-done:
		if got != "xray" {
			t.Fatalf("expected 'xray' from Badger, got %q", got)
		}
	case <-time.After(3 * time.Second):
		t.Fatal("GetAsync timed out")
	}

	// After GetAsync, "x" should be back in hot cache.
	val, ok := c.Get("x")
	if !ok || *val != "xray" {
		t.Fatal("'x' should have been promoted back into hot cache by GetAsync")
	}
}

func TestGetAsyncMiss(t *testing.T) {
	db := openTestDB(t)
	c := newStrLRU(db, 10)

	called := false
	c.GetAsync("nonexistent", func(v string) { called = true })

	// Give goroutine time to run.
	time.Sleep(100 * time.Millisecond)
	if called {
		t.Fatal("callback should not be called for a missing key")
	}
}

// --- 5. Thread safety -------------------------------------------------------

func TestThreadSafety(t *testing.T) {
	db := openTestDB(t)
	c := newStrLRU(db, 20)

	const goroutines = 50
	const ops = 100

	var wg sync.WaitGroup
	wg.Add(goroutines)

	for i := range goroutines {
		go func(id int) {
			defer wg.Done()
			for j := range ops {
				key := fmt.Sprintf("k%d-%d", id, j%10)
				c.Put(key, fmt.Sprintf("v%d", j))
				c.Get(key)
				c.Len()
			}
		}(i)
	}

	// Also exercise GetAsync concurrently.
	wg.Add(goroutines)
	for i := range goroutines {
		go func(id int) {
			defer wg.Done()
			for j := range ops {
				key := fmt.Sprintf("k%d-%d", id, j%10)
				c.GetAsync(key, func(_ string) {})
			}
		}(i)
	}

	wg.Wait()
	// If we reached here without a data-race or panic the test passes.
}
