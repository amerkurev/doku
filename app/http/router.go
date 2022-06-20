package http

import (
	"fmt"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	log "github.com/sirupsen/logrus"

	"github.com/amerkurev/doku/app/http/handler"
	"github.com/amerkurev/doku/app/store"
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

func dockerLogInfo(w http.ResponseWriter, _ *http.Request) {
	v, ok := store.Get("dockerLogInfo")
	if !ok {
		v = emptyObject
	}
	w.Write(v.([]byte)) // nolint:gosec
}

func dockerMountsBind(w http.ResponseWriter, _ *http.Request) {
	v, ok := store.Get("dockerMountsBind")
	if !ok {
		v = emptyObject
	}
	w.Write(v.([]byte)) // nolint:gosec
}

func sizeCalcProgress(w http.ResponseWriter, _ *http.Request) {
	v, ok := store.Get("sizeCalcProgress")
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
		r.Use(handler.BasicAuthentication(s.BasicAuthAllowed))
	}

	r.Group(func(r chi.Router) {
		r.Use(handler.ContentTypeJSON)
		r.Get("/version", version)
		r.Get("/size-calc-progress", sizeCalcProgress)

		// long polling routes
		r.Group(func(r chi.Router) {
			r.Use(handler.LongPolling(s.Timeouts.LongPolling))
			r.Get("/_/docker/disk-usage", dockerDiskUsage)
		})

		r.Route("/docker", func(r chi.Router) {
			r.Get("/info", dockerInfo)
			r.Get("/disk-usage", dockerDiskUsage)
			r.Get("/log-info", dockerLogInfo)
			r.Get("/mounts-bind", dockerMountsBind)
		})
	})
	return r
}
