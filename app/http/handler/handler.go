package handler

import (
	"net/http"
	"time"

	"github.com/amerkurev/doku/app/store"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	log "github.com/sirupsen/logrus"
)

var emptyObject = []byte("{}")

func version(w http.ResponseWriter, _ *http.Request) {
	v, _ := store.Get("revision")
	w.Write([]byte(v.(string))) // nolint gosec
}

func dockerInfo(w http.ResponseWriter, _ *http.Request) {
	v, ok := store.Get("json.dockerInfo")
	if !ok {
		v = emptyObject
	}
	w.Write(v.([]byte)) // nolint gosec
}

func dockerDiskUsage(w http.ResponseWriter, _ *http.Request) {
	v, ok := store.Get("json.dockerDiskUsage")
	if !ok {
		v = emptyObject
	}
	w.Write(v.([]byte)) // nolint gosec
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
		r.Use(ContentTypeJSON)
		r.Use(LongPolling(longPollingTimeout))
		r.Get("/_/docker/disk-usage", dockerDiskUsage)
	})

	r.Get("/version", version)
	r.Route("/docker", func(r chi.Router) {
		r.Use(ContentTypeJSON)
		r.Get("/info", dockerInfo)
		r.Get("/disk-usage", dockerDiskUsage)
	})

	return r
}
