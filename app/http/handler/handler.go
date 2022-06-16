package handler

import (
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
)

func echo(w http.ResponseWriter, req *http.Request) {
}

// CreateRouter creates an HTTP route multiplexer.
func CreateRouter(longPollingTimeout time.Duration) *chi.Mux {
	r := chi.NewRouter()

	// a good base middleware stack
	r.Use(middleware.Recoverer)

	// long polling routes
	r.Route("/long-polling", func(r chi.Router) {
		r.Use(LongPolling(longPollingTimeout))
		r.Get("/echo", echo)
	})

	r.Get("/echo", echo)
	return r
}
