// Copyright 2024 Bill Nixon. All rights reserved.
// Use of this source code is governed by the license found in the LICENSE file.

package webhandler

import (
	"log/slog"
	"net/http"
	"os"
)

// ServeFileHandler creates an HTTP handler that serves a static file from
// the specified filePath. If the file does not exist or cannot be accessed,
// it logs an error and returns an HTTP handler that responds with an HTTP
// 404 (Not Found) error for all requests.
func ServeFileHandler(filePath string) http.HandlerFunc {
	if _, err := os.Stat(filePath); err != nil {
		slog.Error("file does not exist or not accessible",
			slog.String("filePath", filePath),
			slog.Any("error", err))

		// Return a handler that issues an HTTP 404 response.
		return func(w http.ResponseWriter, r *http.Request) {
			http.NotFound(w, r)
		}
	}

	return func(w http.ResponseWriter, r *http.Request) {
		http.ServeFile(w, r, filePath)
	}
}
