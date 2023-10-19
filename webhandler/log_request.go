package webhandler

import (
	"log/slog"
	"net/http"
)

// LogRequest is middleware that logs the details of incoming HTTP requests.
func (h Handler) LogRequest(next http.Handler) http.Handler {
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

		// Call the next handler in the chain with the updated context.
		next.ServeHTTP(w, r)
	})
}
