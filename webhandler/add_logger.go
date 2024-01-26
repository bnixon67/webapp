// Copyright 2023 Bill Nixon. All rights reserved.
// Use of this source code is governed by the license found in the LICENSE file.

package webhandler

import (
	"context"
	"log/slog"
	"net/http"

	"github.com/bnixon67/webapp/webutil"
)

// loggerKeyType is a unique key type to avoid key collisions with other context values.
type loggerKeyType struct{}

// loggerKey is used as a key for storing and retrieving the logger from the context.
var loggerKey = loggerKeyType{}

// RequestLogger returns a logger that includes a "request" group with
// request-specific information.
// This function can be a substitute to AddLogger middleware.
func RequestLogger(r *http.Request) *slog.Logger {
	// Create a slice with basic attributes of the request.
	// This allows addition of attributes to the group based on
	// additional logic or conditions. I could not determine another
	// method to do this with the existing slog package.
	attrValues := []interface{}{
		slog.String("method", r.Method),
		slog.String("url", r.URL.String()),
		slog.String("ip", webutil.RealRemoteAddr(r)),
	}

	// If request ID is not empty, add it to the log attributes.
	if id := RequestID(r.Context()); id != "" {
		attrValues = append(attrValues, slog.String("id", id))
	}

	// Return a new logger instance with the specified attributes.
	return slog.With(slog.Group("request", attrValues...))
}

// RequestLoggerWithFunc returns a RequestLogger that includes function name.
func RequestLoggerWithFunc(r *http.Request) *slog.Logger {
	return RequestLogger(r).With(slog.String("func", FuncNameParent()))
}

// AddLogger is middleware that adds a logger to the request's context.
// This logger is enriched with request-specific attributes and can be
// retrieved in downstream handlers using the Logger function.
func AddLogger(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Create a new logger instance with the specified attributes.
		logger := RequestLogger(r)

		// Add the logger to the request's context.
		ctx := context.WithValue(r.Context(), loggerKey, logger)

		// Call the next handler in the chain with the updated context.
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

// LoggerFromContext retrieves the logger from the given context.
// If the context is nil or does not contain a logger, the default logger
// is returned.
func LoggerFromContext(ctx context.Context) *slog.Logger {
	// Return the default logger if the context is nil.
	if ctx == nil {
		return slog.Default()
	}

	// Attempt to retrieve the logger from the context.
	logger, ok := ctx.Value(loggerKey).(*slog.Logger)

	// If the logger is not found in the context, return the default logger.
	if !ok {
		return slog.Default()
	}

	// Return the logger retrieved from the context.
	return logger
}
