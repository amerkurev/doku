package store

import (
	"sync"
)

var (
	store sync.Map
)

// Set sets the value for a key.
func Set(key string, value any) {
	store.Store(key, value)
}

// Get returns the value stored in the store for a key, or nil if no value is present.
// The ok result indicates whether value was found in the store.
func Get(key string) (value any, ok bool) {
	value, ok = store.Load(key)
	return value, ok
}
