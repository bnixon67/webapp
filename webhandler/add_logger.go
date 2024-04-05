// Copyright 2023 Bill Nixon. All rights reserved.
// Use of this source code is governed by the license found in the LICENSE file.

package webhandler

import (
	"context"
	"log/slog"
	"net/http"

	"github.com/bnixon67/webapp/webutil"
)

// LoggerKeyType is used to avoid key collisions with other context values.
type LoggerKeyType struct{}

// LoggerKey is the key to store and retrieve the logger from the context.
var LoggerKey = LoggerKeyType{}

// RequestLogger creates and returns a logger with request-specific details.
// This function can be a substitute to AddLogger middleware.
func RequestLogger(r *http.Request) *slog.Logger {
	attributes := []interface{}{
		slog.String("method", r.Method),
		slog.String("url", r.URL.String()),
		slog.String("ip", webutil.ClientIP(r)),
	}

	if id := RequestID(r.Context()); id != "" {
		attributes = append(attributes, slog.String("id", id))
	}

	return slog.With(slog.Group("request", attributes...))
}

// RequestLoggerWithFunc enhances RequestLogger with the function name.
func RequestLoggerWithFunc(r *http.Request) *slog.Logger {
	return RequestLogger(r).With(slog.String("func", FuncNameParent()))
}

// AddLogger wraps an HTTP handler to include a request-specific logger in
// its context.
func AddLogger(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		logger := RequestLogger(r)
		ctx := context.WithValue(r.Context(), LoggerKey, logger)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

// LoggerFromContext extracts the logger from ctx.  If ctx is nil or
// does not contain a logger, the slog default logger returned.
func LoggerFromContext(ctx context.Context) *slog.Logger {
	if ctx == nil {
		return slog.Default()
	}

	logger, ok := ctx.Value(LoggerKey).(*slog.Logger)
	if !ok {
		return slog.Default()
	}

	return logger
}
