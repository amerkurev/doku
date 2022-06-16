package http

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/amerkurev/doku/app/http/handler"
	"github.com/amerkurev/doku/app/store"
	log "github.com/sirupsen/logrus"
)

const (
	longPollingTimeout = 30 * time.Second // it must be less than writeTimeout!
	shutdownTimeout    = 5 * time.Second
	readTimeout        = 5 * time.Second
	writeTimeout       = 2 * longPollingTimeout
)

// Run creates and starts an HTTP server.
func Run(ctx context.Context, addr string) error {
	router := handler.CreateRouter(longPollingTimeout)

	httpServer := &http.Server{
		Addr:         addr,
		Handler:      router,
		ReadTimeout:  readTimeout,
		WriteTimeout: writeTimeout,
	}

	done := make(chan bool, 1)

	go func() {
		<-ctx.Done()

		go func() {
			store.Get().NotifyAll() // unlock all long-polling requests
		}()

		// shutdown signal with grace period of `shutdownTimeout` seconds
		ctx, cancel := context.WithTimeout(context.Background(), shutdownTimeout)
		defer func() {
			cancel()
			done <- true
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
