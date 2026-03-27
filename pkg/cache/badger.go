package cache

import (
	"os"

	"github.com/dgraph-io/badger/v4"
)

func NewBadger(overridePath ...string) (*badger.DB, error) {
	path := or(os.Getenv("HOME"), os.Getenv("XDG_HOME")) + "/.spqt/internal_cache"
	if len(overridePath) > 0 {
		path = overridePath[0]
	}

	return badger.Open(badger.DefaultOptions(path))
}

func or[T comparable](ts ...T) T {
	var z T
	for _, t := range ts {
		if t != z {
			return t
		}
	}
	return z
}
