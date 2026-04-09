// Copyright 2023 Bill Nixon. All rights reserved.
// Use of this source code is governed by the license found in the LICENSE file.

// Package webserver provides utilities for creating and managing web servers.
package webserver

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"net"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"
)

// Config holds the web server configuration. Use Config.Create to construct
// a WebServer from it. Use New with Option functions for custom tuning.
type Config struct {
	Host        string        // host address to listen on
	Port        string        // port to listen on
	CertFile    string        // path to the TLS certificate file
	KeyFile     string        // path to the TLS key file
	IdleTimeout time.Duration // keep-alive timeout; defaults to 120s if zero.
}

// WebServer is an HTTP server that supports graceful shutdown and optional TLS.
type WebServer struct {
	certFile         string
	keyFile          string
	HTTPServer       http.Server
	shutdownComplete chan struct{} // closed when shutdown is complete
}

// Option is a function type for configuring the HTTP server.
type Option func(*WebServer)

// WithAddr returns an Option to set the listen address directly.
// Use WithHostPort instead when host and port are separate values,
// as it correctly handles IPv6 addresses via net.JoinHostPort.
func WithAddr(addr string) Option {
	return func(s *WebServer) {
		s.HTTPServer.Addr = addr
	}
}

// WithHostPort returns an Option to set the address the server will bind to.
// It uses net.JoinHostPort to correctly format IPv6 addresses.
func WithHostPort(host, port string) Option {
	return func(s *WebServer) {
		s.HTTPServer.Addr = net.JoinHostPort(host, port)
	}
}

// WithHandler returns an Option to set the HTTP handler of the server.
func WithHandler(h http.Handler) Option {
	return func(s *WebServer) {
		s.HTTPServer.Handler = h
	}
}

// WithTLS returns an Option to set TLS configuration for the server.
func WithTLS(certFile, keyFile string) Option {
	return func(s *WebServer) {
		s.certFile = certFile
		s.keyFile = keyFile
	}
}

// WithReadTimeout returns an Option to set the ReadTimeout of the server.
func WithReadTimeout(d time.Duration) Option {
	return func(s *WebServer) {
		s.HTTPServer.ReadTimeout = d
	}
}

// WithWriteTimeout returns an Option to set the WriteTimeout of the server.
func WithWriteTimeout(d time.Duration) Option {
	return func(s *WebServer) {
		s.HTTPServer.WriteTimeout = d
	}
}

// WithIdleTimeout returns an Option to set the IdleTimeout of the server.
func WithIdleTimeout(d time.Duration) Option {
	return func(s *WebServer) {
		s.HTTPServer.IdleTimeout = d
	}
}

// New creates a new WebServer with the given options and returns it.
func New(opts ...Option) *WebServer {
	s := &WebServer{
		HTTPServer: http.Server{
			ReadTimeout:  10 * time.Second,
			WriteTimeout: 10 * time.Second,
			IdleTimeout:  120 * time.Second,
		},
		shutdownComplete: make(chan struct{}),
	}

	// Apply configuration options to the server.
	for _, opt := range opts {
		opt(s)
	}

	return s
}

// Create constructs a WebServer from cfg. It is equivalent to calling New
// with WithHostPort, WithHandler, WithTLS, and (if non-zero) WithIdleTimeout,
// keeping Config as the single source of truth for those settings.
func (cfg Config) Create(h http.Handler) *WebServer {
	opts := []Option{
		WithHandler(h),
		WithHostPort(cfg.Host, cfg.Port),
		WithTLS(cfg.CertFile, cfg.KeyFile),
	}

	if cfg.IdleTimeout != 0 {
		opts = append(opts, WithIdleTimeout(cfg.IdleTimeout))
	}

	return New(opts...)
}

// ErrServerListen is returned when the server cannot bind to the requested
// address.
var ErrServerListen = errors.New("failed to listen")

// ErrServerServe is returned when the server encounters a fatal error
// while serving.
var ErrServerServe = errors.New("server error")

// Run starts the HTTP server and blocks until the server exits.
// It returns an error if there's an issue starting or stopping the server.
// Shutdown is triggered by ctx cancellation or an OS interrupt/SIGTERM signal.
func (s *WebServer) Run(ctx context.Context) error {
	ln, err := net.Listen("tcp", s.HTTPServer.Addr)
	if err != nil {
		return fmt.Errorf("%w: %w", ErrServerListen, err)
	}

	// Register for OS shutdown signals before starting the goroutine to
	// avoid a race where a signal arrives before Notify is called.
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)
	defer signal.Stop(sigChan)

	// Initialize the serve-error channel.
	errCh := make(chan error, 1)

	// Start the server in a separate goroutine.
	go func() {
		var serverErr error
		if s.certFile != "" && s.keyFile != "" {
			slog.Info("starting https server",
				slog.String("addr", ln.Addr().String()))
			serverErr = s.HTTPServer.ServeTLS(ln, s.certFile, s.keyFile)
		} else {
			slog.Info("starting http server",
				slog.String("addr", ln.Addr().String()))
			serverErr = s.HTTPServer.Serve(ln)
		}

		if serverErr != nil && serverErr != http.ErrServerClosed {
			errCh <- serverErr
		}
	}()

	select {

	case <-ctx.Done():
		// Context cancelled: shut down gracefully so in-flight requests drain.
		// context.WithoutCancel preserves trace/values but strips cancellation,
		// giving shutdownServer its own 5s window independent of the caller.
		slog.Info("context done, shutting down server")
		if err := s.shutdownServer(context.WithoutCancel(ctx)); err != nil {
			return err
		}

	case sig := <-sigChan:
		// OS signal received: gracefully shut down the server.
		slog.Info("shutting down server", "signal", sig)
		if err := s.shutdownServer(ctx); err != nil {
			return err
		}

	case err := <-errCh:
		// Fatal serve error.
		return fmt.Errorf("%w: %w", ErrServerServe, err)
	}

	<-s.shutdownComplete // Wait for shutdown to complete.

	return nil
}

// shutdownServer gracefully stops the server and signals completion.
func (s *WebServer) shutdownServer(ctx context.Context) error {
	// Create a context with timeout to shut down within a reasonable time.
	shutdownCtx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	err := s.HTTPServer.Shutdown(shutdownCtx)
	if err != nil {
		slog.Error("error shutting down server", slog.Any("err", err))
		return err
	}

	close(s.shutdownComplete) // Signal that shutdown is complete.

	slog.Info("server shutdown")

	return nil
}
