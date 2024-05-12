// Copyright 2024 Bill Nixon. All rights reserved.
// Use of this source code is governed by the license found in the LICENSE file.

package webhandler

import (
	"log/slog"
	"net/http"
	"time"
)

// LogRequest is a middleware function that logs the start and completion
// of each HTTP request, along with the duration it took to handle the
// request. It enhances HTTP handlers by providing detailed logging for
// debugging and monitoring.
func LogRequest(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()

		logger := NewRequestLogger(r)
		logger.Info("HTTP request received")

		// Create a response writer to capture the status code.
		lw := newLoggingResponseWriter(w)

		// Process the request.
		next.ServeHTTP(lw, r)

		duration := time.Since(start)
		logger.Info("HTTP request done",
			slog.String("duration", duration.String()),
			slog.Int("status", lw.statusCode),
		)
	})
}

// loggingResponseWriter is a wrapper around http.ResponseWriter that captures
// the HTTP status code for logging purposes.
type loggingResponseWriter struct {
	http.ResponseWriter
	statusCode int
}

// newLoggingResponseWriter creates a new loggingResponseWriter instance.
func newLoggingResponseWriter(w http.ResponseWriter) *loggingResponseWriter {
	// Default to 200 OK, since WriteHeader may not be called explicitly
	// if there is no error.
	return &loggingResponseWriter{ResponseWriter: w, statusCode: http.StatusOK}
}

// WriteHeader captures the status code and delegates to the original
// ResponseWriter.
func (lrw *loggingResponseWriter) WriteHeader(statusCode int) {
	lrw.statusCode = statusCode
	lrw.ResponseWriter.WriteHeader(statusCode)
}
