// Copyright 2024 Bill Nixon. All rights reserved.
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

// LoggerKey is the key used to store and retrieve the logger from the context.
var LoggerKey = LoggerKeyType{}

// NewRequestLogger creates a logger to log HTTP requests with attributes
// like method, URL, and IP.  This function can be a substitute to AddLogger
// middleware.
func NewRequestLogger(r *http.Request) *slog.Logger {
	attributes := []any{
		slog.String("method", r.Method),
		slog.String("url", r.URL.String()),
		slog.String("ip", webutil.ClientIP(r)),
	}

	if id := RequestID(r.Context()); id != "" {
		attributes = append(attributes, slog.String("id", id))
	}

	return slog.With(slog.Group("request", attributes...))
}

// NewRequestLoggerWithFuncName enhances RequestLogger with the function name.
func NewRequestLoggerWithFuncName(r *http.Request) *slog.Logger {
	return NewRequestLogger(r).With(slog.String("func", FuncNameParent()))
}

// MiddlewareLogger adds a logger to the request context and passes it down
// the middleware chain.
func MiddlewareLogger(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		logger := NewRequestLogger(r)
		ctx := context.WithValue(r.Context(), LoggerKey, logger)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

// ExtractLogger retrieves the logger from the context. Returns a default
// logger if none found.
func ExtractLogger(ctx context.Context) *slog.Logger {
	if ctx == nil {
		return slog.Default()
	}

	if logger, ok := ctx.Value(LoggerKey).(*slog.Logger); ok {
		return logger
	}

	return slog.Default()
}
