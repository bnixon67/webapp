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

// AddLogger is middleware that adds a specialized logger to the request's context.
// This logger is enriched with request-specific attributes and can be retrieved in downstream handlers using the Logger function.
func (h Handler) AddLogger(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Create a slice to hold the basic log attributes.
		// A slice is used to simplify adding other attributes, such as RequestID,
		// to the "request" group based on other logic.
		attrValues := []interface{}{
			"method", r.Method,
			"url", r.URL.String(),
			"ip", webutil.RealRemoteAddr(r),
		}

		// If request ID is not empty, add it to the log attributes.
		id := RequestID(r.Context())
		if id != "" {
			attrValues = append(attrValues, "id", id)
		}

		// Create a new logger instance with the specified attributes.
		logger := slog.With(slog.Group("request", attrValues...))

		// Add the logger to the request's context.
		ctx := context.WithValue(r.Context(), loggerKey, logger)

		// Call the next handler in the chain using the updated context.
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

// Logger retrieves the logger from the given context.
// If the context is nil or does not contain a logger, the default logger is returned.
func Logger(ctx context.Context) *slog.Logger {
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
