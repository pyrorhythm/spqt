package cache

import (
	"github.com/dgraph-io/badger/v4"
)

func NewBadger(overridePath ...string) (*badger.DB, error) {
	path := "/tmp/cache"
	if len(overridePath) > 0 {
		path = overridePath[0]
	}

	return badger.Open(badger.DefaultOptions(path))
}
