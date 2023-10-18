package webhandler

import (
	"context"
	"log/slog"
	"net/http"
)

// LoggerKey is a type used for context keys associated with request loggers.
type LoggerKey int

// loggerKey is used as a unique key for the logger to avoid key collisions.
const loggerKey LoggerKey = iota

// AttachRequestLogger is middleware that both logs the details of
// incoming HTTP requests and adds a specialized logger to the request's
// context. This logger, enriched with request-specific details, can be
// retrieved in downstream handlers using the Logger function.
func (h Handler) AttachRequestLogger(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Create a logger with fields for the request's method and URL.
		logger := slog.With(
			slog.Group("request",
				slog.String("method", r.Method),
				slog.String("url", r.URL.String()),
			),
		)

		// Log the incoming request.
		logger.Info("received request")

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
