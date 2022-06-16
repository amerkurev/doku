package handler

import (
	"context"
	"net/http"
	"time"

	"github.com/amerkurev/doku/app/store"
)

// LongPolling holds a client's connection open for as long as possible,
// delivering a response only after data becomes available or within a timeout threshold.
func LongPolling(longPollingTimeout time.Duration) func(next http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		fn := func(w http.ResponseWriter, r *http.Request) {
			<-store.Wait(context.Background(), longPollingTimeout)
			next.ServeHTTP(w, r)
		}
		return http.HandlerFunc(fn)
	}
}
