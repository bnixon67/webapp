package webhandler

import (
	"context"
	"log/slog"
	"net/http"
)

// loggerKeyType is a unique key type to avoid key collisions with other context values.
type loggerKeyType struct{}

// loggerKey is as a key for storing and retrieving the logger from the context.
var loggerKey = loggerKeyType{}

// AttachRequestLogger is middleware adds a specialized logger to the
// request's context. This logger, enriched with request-specific details,
// can be retrieved in downstream handlers using the Logger function.
func (h Handler) AddLogger(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Create a slice to hold the basic log attributes.
		attrValues := []interface{}{
			"method", r.Method,
			"url", r.URL.String(),
		}

		// If request ID is not zero, add it to the log attributes.
		id := RequestID(r.Context())
		if id != "" {
			attrValues = append(attrValues, "id", id)
		}

		// Create a new logger instance with the specified attributes.
		logger := slog.With(slog.Group("request", attrValues...))

		// Add the logger to the request's context.
		ctx := context.WithValue(r.Context(), loggerKey, logger)

		// Call the next handler in the chain with the updated context.
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

// Logger retrieves the logger from the given context.
// If context is nil or does not contain a logger, the default logger is returned.
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
