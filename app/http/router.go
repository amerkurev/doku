package http

import (
	"net/http"
	"os"
	"path"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	log "github.com/sirupsen/logrus"

	"github.com/amerkurev/doku/app/http/handler"
	"github.com/amerkurev/doku/app/store"
)

func getDataFromStore(key string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		v, ok := store.Get(key)
		if !ok {
			v = []byte("{}")
		}
		w.Write(v.([]byte)) // nolint:gosec
	}
}

// CreateRouter creates an HTTP route multiplexer.
func CreateRouter(s *Server) *chi.Mux {
	r := chi.NewRouter()

	// a good base middleware stack
	r.Use(middleware.RequestID)
	r.Use(middleware.RealIP)
	r.Use(handler.NewStructuredLogger(log.StandardLogger()))
	r.Use(middleware.Recoverer)

	if os.Getenv("ENVIRONMENT") == "dev" {
		r.Use(handler.DevCORS)
	}

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
		r.Route("/v0", func(r chi.Router) {
			r.Use(handler.ContentTypeJSON)
			r.Get("/version", getDataFromStore("revision"))
			r.Get("/disk-usage", getDataFromStore("diskUsage"))

			// long polling routes
			r.Group(func(r chi.Router) {
				r.Use(handler.LongPolling(s.Timeouts.LongPolling))
				r.Get("/_/docker/disk-usage", getDataFromStore("dockerDiskUsage"))
			})

			r.Route("/docker", func(r chi.Router) {
				r.Get("/version", getDataFromStore("dockerVersion"))
				r.Get("/containers", getDataFromStore("dockerContainerList"))
				r.Get("/disk-usage", getDataFromStore("dockerDiskUsage"))
				r.Get("/log-size", getDataFromStore("dockerLogSize"))
				r.Get("/bind-mounts", getDataFromStore("dockerBindMounts"))
			})
		})

		// Static
		r.Group(func(r chi.Router) {
			r.Use(middleware.Compress(5, "text/html", "text/css", "text/javascript", "application/javascript"))

			// Everything else falls back on static content
			// https://github.com/go-chi/chi/issues/403#issuecomment-900144943
			fileServer := http.FileServer(http.Dir(s.StaticFolder))
			r.Handle("/static/*", http.StripPrefix("/static", fileServer))

			// SPA
			indexHTML := path.Join(s.StaticFolder, "index.html")
			r.Get("/*", handler.SinglePageApplication(indexHTML, s.UITitle, s.UIHeader))
		})
	})
	return r
}
