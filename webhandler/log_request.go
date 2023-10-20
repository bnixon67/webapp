package webhandler

import (
	"net/http"
)

// LogRequest is middleware that logs the details of incoming HTTP requests.
func (h Handler) LogRequest(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		logger := Logger(r.Context())

		// Log the incoming request.
		logger.Info("received request")

		// Call the next handler in the chain with the updated context.
		next.ServeHTTP(w, r)
	})
}
