// Copyright 2024 Bill Nixon. All rights reserved.
// Use of this source code is governed by the license found in the LICENSE file.

package webhandler

import (
	"context"
	"log/slog"
	"net/http"

	"github.com/bnixon67/webapp/webutil"
)

// loggerKeyType is a custom type to avoid key collisions in context values.
type loggerKeyType struct{}

// loggerKey is a unique identifier for retrieving a logger from a context.
var loggerKey = loggerKeyType{}

// NewRequestLogger creates and configures a logger specifically for logging
// HTTP request details, such as the method, URL, and client IP. It optionally
// includes a request ID if present.
func NewRequestLogger(r *http.Request) *slog.Logger {
	attributes := []any{
		slog.String("method", r.Method),
		slog.String("url", r.URL.String()),
		slog.String("ip", webutil.ClientIP(r)),
	}

	// Append request ID to the log attributes if available.
	if id := RequestID(r.Context()); id != "" {
		attributes = append(attributes, slog.String("id", id))
	}

	return slog.With(slog.Group("request", attributes...))
}

// NewRequestLoggerWithFuncName augments a request logger by adding the
// caller function's name to the log attributes.
func NewRequestLoggerWithFuncName(r *http.Request) *slog.Logger {
	return NewRequestLogger(r).With(slog.String("func", FuncNameParent()))
}

// MiddlewareLogger creates middleware that injects a logger into the request
// context, enabling subsequent handlers in the chain to log request-specific
// information.
func MiddlewareLogger(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		logger := NewRequestLogger(r)
		ctx := context.WithValue(r.Context(), loggerKey, logger)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

// Logger attempts to retrieve a logger from the provided context.
// It returns a default logger if no custom logger is found or if the context
// is nil.
func Logger(ctx context.Context) *slog.Logger {
	if ctx == nil {
		return slog.Default()
	}

	if logger, ok := ctx.Value(loggerKey).(*slog.Logger); ok {
		return logger
	}

	return slog.Default()
}
