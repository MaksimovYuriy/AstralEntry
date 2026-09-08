package httpserver

import (
	"context"
	"errors"
	"fmt"
	"net/http"
)

type Server struct {
	server *http.Server
}

func New(address string, handler http.Handler, options ...Option) *Server {
	httpServer := &http.Server{
		Addr:    address,
		Handler: handler,
	}

	for _, option := range options {
		option(httpServer)
	}

	return &Server{server: httpServer}
}

func (s *Server) Run() error {
	if err := s.server.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
		return fmt.Errorf("http server: listen and serve: %w", err)
	}

	return nil
}

func (s *Server) Shutdown(ctx context.Context) error {
	if err := s.server.Shutdown(ctx); err != nil {
		return fmt.Errorf("http server: shutdown: %w", err)
	}

	return nil
}
