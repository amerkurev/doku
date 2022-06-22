package http

import (
	"context"
	"fmt"
	"net/http"
	"time"

	log "github.com/sirupsen/logrus"

	"github.com/amerkurev/doku/app/store"
)

// Server is a server for http.
type Server struct {
	Address          string
	Timeouts         Timeouts
	BasicAuthEnabled bool
	BasicAuthAllowed []string
	StaticFolder     string
	UITitle          string
	UIHeader         string
}

// Timeouts consolidate timeouts for both server and transport
type Timeouts struct {
	Write       time.Duration
	Read        time.Duration
	Idle        time.Duration
	Shutdown    time.Duration
	LongPolling time.Duration
}

// Run creates and starts an HTTP server.
func (s *Server) Run(ctx context.Context) error {
	router := CreateRouter(s)

	httpServer := &http.Server{
		Addr:         s.Address,
		Handler:      router,
		ReadTimeout:  s.Timeouts.Read,
		WriteTimeout: s.Timeouts.Write,
		IdleTimeout:  s.Timeouts.Idle,
	}

	done := make(chan struct{}, 1)

	go func() {
		<-ctx.Done()

		// shutdown signal with grace period of `shutdownTimeout` seconds
		ctx, cancel := context.WithTimeout(context.Background(), s.Timeouts.Shutdown)
		defer func() {
			cancel()
			done <- struct{}{}
		}()

		go func() {
			store.NotifyAll() // unlock all long-polling requests before shut down
		}()

		err := httpServer.Shutdown(ctx)
		if err != nil {
			log.WithField("err", err).Error("failed to shut down the http server")
			return
		}
		log.Info("gracefully http server shutdown")
	}()

	err := httpServer.ListenAndServe()
	if err != nil && err == http.ErrServerClosed {
		<-done
		return nil
	}

	return fmt.Errorf("http server failed: %w", err)
}
