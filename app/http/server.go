package http

import (
	"context"
	"net"
	"net/http"
	"time"

	"github.com/amerkurev/doku/app/store"
	log "github.com/sirupsen/logrus"
)

const (
	longPollingTimeout = 30 * time.Second // must be less than writeTimeout!
	shutdownTimeout    = 5 * time.Second
	readTimeout        = 5 * time.Second
	writeTimeout       = 2 * longPollingTimeout
)

// Server represents an interface of control over the HTTP server.
type Server interface {
	Stop() error
}

type server struct {
	*http.Server
}

func longPolling(w http.ResponseWriter, req *http.Request) {
	<-store.Get().Wait(context.Background(), longPollingTimeout)
}

// NewServer creates an HTTP server.
func NewServer(addr string) (Server, error) {
	ln, err := net.Listen("tcp", addr)
	if err != nil {
		return nil, err
	}

	handler := http.NewServeMux()
	handler.HandleFunc("/poll", longPolling)

	s := &server{
		Server: &http.Server{
			Addr:         addr,
			Handler:      handler,
			ReadTimeout:  readTimeout,
			WriteTimeout: writeTimeout,
		},
	}

	go func() {
		if err := s.Serve(ln); err != http.ErrServerClosed {
			log.WithField("err", err).Error("error when the HTTP request handling")
		}
	}()
	return s, nil
}

func (s *server) Serve(ln net.Listener) error {
	return s.Server.Serve(ln)
}

// Stop stops HTTP server.
func (s *server) Stop() error {
	ctx, cancel := context.WithTimeout(context.Background(), shutdownTimeout)
	defer cancel()
	err := s.Server.Shutdown(ctx)
	return err
}
