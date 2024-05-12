// Copyright 2024 Bill Nixon. All rights reserved.
// Use of this source code is governed by the license found in the LICENSE file.

// Package webutils provides utility functions.
package webutil

import (
	"log/slog"
	"net/http"
	"os"
)

// ServeFileHandler returns an HTTP handler that serves a specified file.
func ServeFileHandler(filePath string) http.HandlerFunc {
	if _, err := os.Stat(filePath); err != nil {
		slog.Error("does not exist or not accessible",
			slog.String("filePath", filePath),
			slog.Any("error", err))
		return nil
	}

	return func(w http.ResponseWriter, r *http.Request) {
		http.ServeFile(w, r, filePath)
	}
}
