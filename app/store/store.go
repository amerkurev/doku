package store

import (
	"context"
	"sync"
	"time"

	log "github.com/sirupsen/logrus"
)

// Store is the interface that represents storage functionality.
type Store interface {
	Set(key string, value any)
	Get(key string) (value any, ok bool)

	Wait(ctx context.Context, d time.Duration) <-chan struct{}
	NotifyAll()
}

var (
	store Store
	once  sync.Once
)

const errStoreUninitialized = "store requested before it was initialized"

// Initialize instantiates the store.
func Initialize() error {
	var err error
	once.Do(func() {
		store, err = NewMemStore()
	})
	return err
}

// Set sets the value for a key.
func Set(key string, value any) {
	if store == nil {
		log.Panic(errStoreUninitialized)
	}
	store.Set(key, value)
}

// Get returns the value stored in the store for a key, or nil if no value is present.
// The ok result indicates whether value was found in the store.
func Get(key string) (value any, ok bool) {
	if store == nil {
		log.Panic(errStoreUninitialized)
	}
	return store.Get(key)
}

// Wait returns a closed channel only after notification or when the timeout elapses.
func Wait(ctx context.Context, d time.Duration) <-chan struct{} {
	if store == nil {
		log.Panic(errStoreUninitialized)
	}
	return store.Wait(ctx, d)
}

// NotifyAll wakes up those who are waiting.
func NotifyAll() {
	if store == nil {
		log.Panic(errStoreUninitialized)
	}
	store.NotifyAll()
}
