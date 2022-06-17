package http

import (
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
	w.Write([]byte(v.(string))) // nolint:gosec
}

func dockerInfo(w http.ResponseWriter, _ *http.Request) {
	v, ok := store.Get("json.dockerInfo")
	if !ok {
		v = emptyObject
	}
	w.Write(v.([]byte)) // nolint:gosec
}

func dockerDiskUsage(w http.ResponseWriter, _ *http.Request) {
	v, ok := store.Get("json.dockerDiskUsage")
	if !ok {
		v = emptyObject
	}
	w.Write(v.([]byte)) // nolint:gosec
}

func volumeUsage(w http.ResponseWriter, _ *http.Request) {
	v, ok := store.Get("json.volumeUsage")
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
		// long polling routes
		r.Group(func(r chi.Router) {
			r.Use(handler.ContentTypeJSON)
			r.Use(handler.LongPolling(s.Timeouts.LongPolling))
			r.Get("/_/docker/disk-usage", dockerDiskUsage)
		})

		r.Get("/version", version)
		r.Route("/docker", func(r chi.Router) {
			r.Use(handler.ContentTypeJSON)
			r.Get("/info", dockerInfo)
			r.Get("/disk-usage", dockerDiskUsage)
		})

		r.Route("/host", func(r chi.Router) {
			r.Use(handler.ContentTypeJSON)
			r.Get("/volume-usage", volumeUsage)
		})
	})
	return r
}
