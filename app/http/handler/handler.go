package handler

import (
	"net/http"
	"time"

	"github.com/amerkurev/doku/app/store"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	log "github.com/sirupsen/logrus"
)

func echo(w http.ResponseWriter, req *http.Request) {
}

func version(w http.ResponseWriter, req *http.Request) {
	revision, _ := store.Get().Get("revision")
	_, _ = w.Write([]byte(revision.(string)))
}

// CreateRouter creates an HTTP route multiplexer.
func CreateRouter(longPollingTimeout time.Duration) *chi.Mux {
	r := chi.NewRouter()

	// a good base middleware stack
	r.Use(middleware.RequestID)
	r.Use(middleware.RealIP)
	r.Use(NewStructuredLogger(log.StandardLogger()))
	r.Use(middleware.Recoverer)

	// long polling routes
	r.Group(func(r chi.Router) {
		r.Use(LongPolling(longPollingTimeout))
		r.Get("/_/echo", echo)
	})

	r.Get("/version", version)
	return r
}
