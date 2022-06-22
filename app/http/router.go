package http

import (
	"net/http"
	"path"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	log "github.com/sirupsen/logrus"

	"github.com/amerkurev/doku/app/http/handler"
	"github.com/amerkurev/doku/app/store"
)

var emptyObject = []byte("{}")

func version(w http.ResponseWriter, _ *http.Request) {
	v, ok := store.Get("revision")
	if !ok {
		v = emptyObject
	}
	w.Write(v.([]byte)) // nolint:gosec
}

func dockerVersion(w http.ResponseWriter, _ *http.Request) {
	v, ok := store.Get("dockerVersion")
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

func dockerLogSize(w http.ResponseWriter, _ *http.Request) {
	v, ok := store.Get("dockerLogSize")
	if !ok {
		v = emptyObject
	}
	w.Write(v.([]byte)) // nolint:gosec
}

func dockerBindMounts(w http.ResponseWriter, _ *http.Request) {
	v, ok := store.Get("dockerBindMounts")
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

	// Misc
	r.Get("/favicon.ico", handler.ServeFile(path.Join(s.StaticFolder, "favicon.ico")))
	r.Get("/manifest.json", handler.ServeFile(path.Join(s.StaticFolder, "manifest.json")))

	// Protected routes
	r.Group(func(r chi.Router) {
		if s.BasicAuthEnabled {
			log.Debugln("basic auth is enabled")
			r.Use(handler.BasicAuthentication(s.BasicAuthAllowed))
		}

		// API for frontend
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
				r.Get("/version", dockerVersion)
				r.Get("/disk-usage", dockerDiskUsage)
				r.Get("/log-size", dockerLogSize)
				r.Get("/bind-mounts", dockerBindMounts)
			})
		})

		// Static
		r.Group(func(r chi.Router) {
			r.Use(middleware.Compress(5, "text/html", "text/css", "text/javascript", "application/javascript"))

			// SPA
			indexHTML := path.Join(s.StaticFolder, "index.html")
			r.Get("/", handler.SinglePageApplication(indexHTML, s.UITitle, s.UIHeader))

			// Everything else falls back on static content
			// https://github.com/go-chi/chi/issues/403#issuecomment-900144943
			fileServer := http.FileServer(http.Dir(s.StaticFolder))
			r.Handle("/static/*", http.StripPrefix("/static", fileServer))
		})
	})
	return r
}
