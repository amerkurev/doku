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

	done := make(chan bool)

	go func() {
		<-ctx.Done()
		go func() {
			store.Get().NotifyAll() // unlock all long-polling requests
		}()
		err := httpServer.Shutdown(context.Background())
		if err != nil {
			log.WithField("err", err).Error("failed to shut down the http server")
		}
		done <- true
	}()

	err := httpServer.ListenAndServe()
	if err != nil && err == http.ErrServerClosed {
		<-done
		log.Info("gracefully http server shutdown")
		return nil
	}

	return fmt.Errorf("http server failed: %w", err)
}
