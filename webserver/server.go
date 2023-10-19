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

	slog.Info("starting server", slog.String("addr", ln.Addr().String()))

	// Start the server in a separate goroutine, and push any errors to errCh.
	go func() {
		//err = http.ErrNotSupported
		err = srv.Serve(ln)
		if err != nil && err != http.ErrServerClosed {
			errCh <- err
		}
	}()

	// Ask for notification of shutdown signals to gracefully shutdown server.
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)

	select {
	case <-ctx.Done():
		// If the context is done, return its error.
		return ctx.Err()
	case err := <-errCh:
		// If there's an error from the server, return it.
		return err
	case sig := <-sigChan:
		// If a shutdown signal received, gracefully shut down server.
		return shutdownServer(sig, sigChan, srv, ctx)
	}
}

// shutdownServer attempts to gracefully shut down the server.
func shutdownServer(sig os.Signal, sigChan chan os.Signal, srv *http.Server, ctx context.Context) error {
	signal.Stop(sigChan)
	slog.Info("shutting down server", "signal", sig)

	// Create a context with a timeout to shut down within a reasonable time.
	shutdownCtx, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()

	err := srv.Shutdown(shutdownCtx)
	if err != nil {
		return err
	}

	slog.Info("server shutdown")
	return nil
}
