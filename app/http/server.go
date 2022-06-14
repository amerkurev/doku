package http

import (
	"context"
	"net"
	"net/http"
	"time"

	log "github.com/sirupsen/logrus"
)

const (
	ReadTimeout     = 15 * time.Second
	WriteTimeout    = 15 * time.Second
	IdleTimeout     = 15 * time.Second
	ShutdownTimeout = 15 * time.Second
)

// Server represents an interface of control over the HTTP server.
type Server interface {
	Stop() error
}

type server struct {
	*http.Server
}

// NewServer creates an HTTP server.
func NewServer(addr string) (Server, error) {
	ln, err := net.Listen("tcp", addr)
	if err != nil {
		return nil, err
	}

	s := &server{
		Server: &http.Server{
			Addr:         addr,
			ReadTimeout:  ReadTimeout,
			WriteTimeout: WriteTimeout,
			IdleTimeout:  IdleTimeout,
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
	ctx, cancel := context.WithTimeout(context.Background(), ShutdownTimeout)
	defer cancel()
	err := s.Server.Shutdown(ctx)
	return err
}
