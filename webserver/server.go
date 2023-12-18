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

// WebServer represents an HTTP server.
type WebServer struct {
	HTTPServer       http.Server
	CertFile         string
	KeyFile          string
	shutdownComplete chan struct{} // Channel to signal shutdown completion

}

// Option is a function type for configuring the HTTP server.
// This follows the Option pattern from https://commandcenter.blogspot.com/2014/01/self-referential-functions-and-design.html and elsewhere.
type Option func(*WebServer)

// WithAddr returns an Option to set the address the server will bind to.
func WithAddr(addr string) Option {
	return func(s *WebServer) {
		s.HTTPServer.Addr = addr
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
		s.CertFile = certFile
		s.KeyFile = keyFile
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

// New creates a new HTTP server with the given options and returns it.
func New(opts ...Option) (*WebServer, error) {
	s := &WebServer{
		HTTPServer: http.Server{
			ReadTimeout:  10 * time.Second,
			WriteTimeout: 10 * time.Second,
			IdleTimeout:  120 * time.Second,
		},
	}

	// Apply configuration options to the server.
	for _, opt := range opts {
		opt(s)
	}

	return s, nil
}

var ErrServerStart = errors.New("failed to start server")

// Start starts the HTTP server and waits for a shutdown signal.
// It returns an error if there's an issue starting the server.
func (s *WebServer) Start(ctx context.Context) error {
	ln, err := net.Listen("tcp", s.HTTPServer.Addr)
	if err != nil {
		return fmt.Errorf("%w: %v", ErrServerStart, err)
	}

	// Initialize the channels
	errCh := make(chan error, 1)
	s.shutdownComplete = make(chan struct{})

	// Start the server in a separate goroutine.
	go func() {
		var serverErr error
		if s.CertFile != "" && s.KeyFile != "" {
			slog.Info("starting https server",
				slog.String("addr", ln.Addr().String()))

			serverErr = s.HTTPServer.ServeTLS(ln, s.CertFile, s.KeyFile)
		} else {
			slog.Info("starting http server",
				slog.String("addr", ln.Addr().String()))

			serverErr = s.HTTPServer.Serve(ln)
		}

		if serverErr != nil && serverErr != http.ErrServerClosed {
			errCh <- serverErr
		}
	}()

	// Ask for notification of shutdown signals to shut down server.
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)

	select {
	case <-ctx.Done():
		// If the context is done, return its error.
		return ctx.Err()
	case sig := <-sigChan:
		// If a shutdown signal, gracefully shut down the server.
		err := s.shutdownServer(sig, ctx)
		if err != nil {
			return err
		}
	case err := <-errCh:
		// Handle the error that occurred during server startup.
		return fmt.Errorf("%w: %v", ErrServerStart, err)
	}

	<-s.shutdownComplete // Wait for shutdown to complete
	return nil
}

// shutdownServer attempts to gracefully shut down the server.
func (s *WebServer) shutdownServer(sig os.Signal, ctx context.Context) error {
	slog.Info("shutting down server", "signal", sig)

	// Create a context with timeout to shut down within a reasonable time.
	shutdownCtx, cancel := context.WithTimeout(ctx, 10*time.Second)
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
