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
	Notify()
}

var (
	store Store
	once  sync.Once
)

// Initialize instantiates the store.
func Initialize() error {
	var err error
	once.Do(func() {
		store, err = NewMemStore()
	})
	return err
}

// Get returns current initialized store.
func Get() Store {
	if store == nil {
		// This only happens in tests
		log.Warning("store requested before it was initialized, automatically initializing")
		err := Initialize()
		if err != nil {
			log.WithField("err", err).Fatal("failed to automatically initialize store")
		}
	}
	return store
}
