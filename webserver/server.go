// Package webserver provides utilities for creating and managing web servers.
package webserver

import (
	"context"
	"net"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"log/slog"
)

// Option is a function type for configuring the HTTP server.
type Option func(*http.Server)

// WithAddr returns an Option to set the address the server will bind to.
func WithAddr(addr string) Option {
	return func(s *http.Server) {
		s.Addr = addr
	}
}

// WithHandler returns an Option to set the HTTP handler of the server.
func WithHandler(h http.Handler) Option {
	return func(s *http.Server) {
		s.Handler = h
	}
}

// New creates a new HTTP server with the given options and returns it.
func New(opts ...Option) (*http.Server, error) {
	s := &http.Server{}

	// Apply configuration options to the server.
	for _, opt := range opts {
		opt(s)
	}

	return s, nil
}

// Run starts the HTTP server and waits for a shutdown signal.
// It returns an error if there's an issue starting the server.
func Run(ctx context.Context, srv *http.Server) error {
	ln, err := net.Listen("tcp", srv.Addr)
	if err != nil {
		return err
	}

	errCh := make(chan error, 1)

	// Start server in a goroutine so it doesn't block the signal listening.
	go func() {
		err = srv.Serve(ln)
		if err != nil && err != http.ErrServerClosed {
			slog.Error("failed to serve", "err", err)
			errCh <- err
			return
		}
	}()

	slog.Info("started server", slog.String("addr", ln.Addr().String()))

	// Listen for shutdown signals to gracefully shutdown the server.
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)

	select {
	case sig := <-sigChan:
		signal.Stop(sigChan)
		slog.Info("shutting down server", "signal", sig)

		// Create context with timeout for server Shutdown.
		timeoutCtx, cancel := context.WithTimeout(ctx, 10*time.Second)
		defer cancel()

		// Attemp to gracefully shutdown the server.
		err := srv.Shutdown(timeoutCtx)
		if err != nil {
			slog.Error("server shutdown error", "err", err)
			return err
		}

		slog.Info("server shutdown")
		return nil
	case err := <-errCh:
		return err
	case <-ctx.Done():
		return ctx.Err()
	}
}
