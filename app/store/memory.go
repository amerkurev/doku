package store

import (
	"context"
	"sync"
	"time"
)

type memoryStore struct {
	sync.Map

	mx        sync.Mutex
	cancelers []context.CancelFunc
}

func (s *memoryStore) Set(key string, value any) {
	s.Store(key, value)
}

func (s *memoryStore) Get(key string) (value any, ok bool) {
	value, ok = s.Load(key)
	return value, ok
}

func (s *memoryStore) Wait(ctx context.Context, d time.Duration) <-chan struct{} {
	ctx, cancelFunc := context.WithTimeout(ctx, d)
	s.mx.Lock()
	defer s.mx.Unlock()
	s.cancelers = append(s.cancelers, cancelFunc)
	return ctx.Done()
}

func (s *memoryStore) NotifyAll() {
	for _, cancelFunc := range s.loadAndDeleteCancelers() {
		cancelFunc()
	}
}

func (s *memoryStore) loadAndDeleteCancelers() []context.CancelFunc {
	s.mx.Lock()
	defer s.mx.Unlock()
	r := s.cancelers
	s.cancelers = nil
	return r
}

// NewMemStore creates a memory store that implements the Store interface.
func NewMemStore() (Store, error) {
	s := &memoryStore{}
	return s, nil
}
