// Copyright 2023 Bill Nixon. All rights reserved.
// Use of this source code is governed by the license found in the LICENSE file.

package webhandler

import (
	"log/slog"
	"net/http"
)

// LogRequest is middleware that logs the details of incoming HTTP requests.
// It assumes that AddLogger was called prior to enrich the logger with request-specific attributes.
func (h Handler) LogRequest(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Get logger with request info from request context and add calling function name.
		logger := Logger(r.Context()).With(slog.String("func", "LogRequest"))

		// Log the incoming request.
		logger.Info("received")

		// Call the next handler in the chain with the updated context.
		next.ServeHTTP(w, r)
	})
}
