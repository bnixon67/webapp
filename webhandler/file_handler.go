// Copyright 2024 Bill Nixon. All rights reserved.
// Use of this source code is governed by the license found in the LICENSE file.

package webhandler

import (
	"log/slog"
	"net/http"
	"os"
)

// FileHandler returns a HTTP handler that serves a specified file from the
// provided name. This handler uses http.ServeFile to serve the file directly.
//
// If the file specified by name does not exist or is not accessible, the
// handler logs the error and returns an HTTP 404 (Not Found) response for
// all incoming requests.
func FileHandler(name string) http.HandlerFunc {
	// Check if the file exists and is accessible.
	if _, err := os.Stat(name); err != nil {
		slog.Error("file check failed",
			slog.String("filePath", name),
			slog.String("error", err.Error()))

		// Return a handler that issues an HTTP 404 response.
		return func(w http.ResponseWriter, r *http.Request) {
			http.NotFound(w, r)
		}
	}

	return func(w http.ResponseWriter, r *http.Request) {
		http.ServeFile(w, r, name)
	}
}
