// Copyright 2023 Bill Nixon. All rights reserved.
// Use of this source code is governed by the license found in the LICENSE file.

package webhandler

import (
	"log/slog"
	"net/http"
)

// LogRequest is a middleware function that logs the details of incoming HTTP requests.
// It is designed to wrap around other HTTP handlers, adding logging functionality to them.
func LogRequest(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Retrieve a logger pre-configured with request information.
		// GetRequestLogger is used instead of obtaining the logger directly from the context.
		// This ensures compatibility if AddLogger middleware was not.
		logger := RequestLogger(r).With(slog.String("func", "LogRequest"))

		// Log the incoming request.
		logger.Info("received")

		// Call the next handler in the chain with the updated context.
		next.ServeHTTP(w, r)
	})
}
