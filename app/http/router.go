package http

import (
	"context"
	"net/http"
	"os"
	"path"

	"github.com/amerkurev/doku/app/docker"
	"github.com/go-chi/chi/v5"
	chiMiddleware "github.com/go-chi/chi/v5/middleware"
	log "github.com/sirupsen/logrus"

	"github.com/amerkurev/doku/app/http/handler"
	"github.com/amerkurev/doku/app/http/middleware"
)

// CreateRouter creates an HTTP route multiplexer.
func CreateRouter(ctx context.Context, s *Server, d *docker.Client) *chi.Mux {
	r := chi.NewRouter()

	// a good base middleware stack
	r.Use(chiMiddleware.RequestID)
	r.Use(chiMiddleware.RealIP)
	r.Use(middleware.NewStructuredLogger(log.StandardLogger()))
	r.Use(chiMiddleware.Recoverer)

	if os.Getenv("ENVIRONMENT") == "dev" {
		r.Use(middleware.DevCORS)
	}

	// Misc
	r.Get("/favicon.ico", middleware.ServeFile(path.Join(s.StaticFolder, "favicon.ico")))
	r.Get("/manifest.json", middleware.ServeFile(path.Join(s.StaticFolder, "manifest.json")))

	// Protected routes
	r.Group(func(r chi.Router) {
		if s.BasicAuthEnabled {
			log.Debugln("basic auth is enabled")
			r.Use(middleware.BasicAuthentication(s.BasicAuthAllowed))
		}

		// API for frontend
		r.Route("/v0", func(r chi.Router) {
			r.Use(middleware.ContentTypeJSON)
			r.Get("/version", handler.Version(ctx))
			r.Get("/disk-usage", handler.DiskUsage(ctx))

			r.Route("/docker", func(r chi.Router) {
				r.Get("/version", handler.DockerVersion(ctx, d))
				r.Get("/containers", handler.DockerContainerList(ctx, d))
				r.Get("/disk-usage", handler.DockerDiskUsage(ctx, d))
				r.Get("/log-size", handler.DockerLogSize(ctx, d))
				r.Get("/bind-mounts", handler.DockerBindMounts())
			})
		})

		// Static
		r.Group(func(r chi.Router) {
			r.Use(chiMiddleware.Compress(5, "text/html", "text/css", "text/javascript", "application/javascript"))

			// Everything else falls back on static content
			// https://github.com/go-chi/chi/issues/403#issuecomment-900144943
			fileServer := http.FileServer(http.Dir(s.StaticFolder))
			r.Handle("/static/*", http.StripPrefix("/static", fileServer))

			// SPA
			indexHTML := path.Join(s.StaticFolder, "index.html")
			r.Get("/*", middleware.SinglePageApplication(indexHTML, s.UITitle, s.UIHeader))
		})
	})
	return r
}
