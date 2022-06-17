package http

import (
	"fmt"
	"github.com/amerkurev/doku/app/http/handler"
	"github.com/amerkurev/doku/app/store"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	log "github.com/sirupsen/logrus"
	"net/http"
)

var emptyObject = []byte("{}")

func version(w http.ResponseWriter, _ *http.Request) {
	v, _ := store.Get("revision")
	b := []byte(fmt.Sprintf(`{"version": "%s"}`, v.(string)))
	w.Write(b) // nolint:gosec
}

func dockerInfo(w http.ResponseWriter, _ *http.Request) {
	v, ok := store.Get("dockerInfo")
	if !ok {
		v = emptyObject
	}
	w.Write(v.([]byte)) // nolint:gosec
}

func dockerDiskUsage(w http.ResponseWriter, _ *http.Request) {
	v, ok := store.Get("dockerDiskUsage")
	if !ok {
		v = emptyObject
	}
	w.Write(v.([]byte)) // nolint:gosec
}

func hostDiskUsage(w http.ResponseWriter, _ *http.Request) {
	v, ok := store.Get("hostDiskUsage")
	if !ok {
		v = emptyObject
	}
	w.Write(v.([]byte)) // nolint:gosec
}

// CreateRouter creates an HTTP route multiplexer.
func CreateRouter(s *Server) *chi.Mux {
	r := chi.NewRouter()

	// a good base middleware stack
	r.Use(middleware.RequestID)
	r.Use(middleware.RealIP)
	r.Use(handler.NewStructuredLogger(log.StandardLogger()))
	r.Use(middleware.Recoverer)

	if s.BasicAuthEnabled {
		log.Debugln("basic auth is enabled")
		r.Use(handler.BasicAuthHandler(s.BasicAuthAllowed))
	}

	r.Route("/api", func(r chi.Router) {
		r.Use(handler.ContentTypeJSON)
		r.Get("/version", version)

		// long polling routes
		r.Group(func(r chi.Router) {
			r.Use(handler.LongPolling(s.Timeouts.LongPolling))
			r.Get("/_/docker/disk-usage", dockerDiskUsage)
		})

		r.Route("/docker", func(r chi.Router) {
			r.Get("/info", dockerInfo)
			r.Get("/disk-usage", dockerDiskUsage)
		})

		r.Route("/host", func(r chi.Router) {
			r.Get("/disk-usage", hostDiskUsage)
		})
	})
	return r
}
