package httpserver

import (
	"net/http"
	"time"
)

type Option func(*http.Server)

func WithReadTimeout(timeout time.Duration) Option {
	return func(server *http.Server) {
		server.ReadTimeout = timeout
	}
}

func WithWriteTimeout(timeout time.Duration) Option {
	return func(server *http.Server) {
		server.WriteTimeout = timeout
	}
}

func WithReadHeaderTimeout(timeout time.Duration) Option {
	return func(server *http.Server) {
		server.ReadHeaderTimeout = timeout
	}
}

func WithIdleTimeout(timeout time.Duration) Option {
	return func(server *http.Server) {
		server.IdleTimeout = timeout
	}
}
