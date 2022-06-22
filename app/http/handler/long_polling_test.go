package handler

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/amerkurev/doku/app/store"
)

func Test_LongPolling(t *testing.T) {
	err := store.Initialize()
	require.NoError(t, err)

	wr := httptest.NewRecorder()
	handler := LongPolling(time.Minute)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {}))

	start := time.Now()
	go func() {
		time.Sleep(time.Millisecond * 100)
		store.NotifyAll()
	}()
	req, err := http.NewRequest("GET", "http://example.com", http.NoBody)
	require.NoError(t, err)

	handler.ServeHTTP(wr, req)
	assert.Less(t, time.Since(start), time.Second)
}
