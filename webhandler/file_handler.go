// Copyright 2025 Bill Nixon. All rights reserved.
// Use of this source code is governed by the license found in the LICENSE file.

package webhandler

import (
	"log/slog"
	"net/http"
	"os"
	"path/filepath"
	"strings"
)

// FileHandler returns an http.HandlerFunc that serves the single file at path.
// If the file does not exist or cannot be accessed at startup, the returned
// handler always responds with 404 Not Found and logs the error once.
//
// Typical use is to serve a static page such as /robots.txt or /favicon.ico.
func FileHandler(path string) http.HandlerFunc {
	if _, err := os.Stat(path); err != nil {
		slog.Error("file check failed",
			slog.String("path", path),
			slog.String("error", err.Error()))

		return func(w http.ResponseWriter, r *http.Request) {
			http.NotFound(w, r)
		}
	}

	return func(w http.ResponseWriter, r *http.Request) {
		http.ServeFile(w, r, path)
	}
}

// FilesFromDir returns an http.HandlerFunc that serves files from dir at the
// given urlPrefix (e.g., "/css/"), rejecting direct directory requests.
//
// The returned handler:
//   - Strips urlPrefix from the request path to map to the file under dir.
//   - Responds 404 if the request points to a directory (ends with "/").
//   - Uses http.ServeFile so content type, caching headers, and range requests
//     work as usual.
//
// Example:
//
//	mux.Handle("/css/", FilesFromDir("/css/", "assets/css"))
func FilesFromDir(urlPrefix, dir string) http.HandlerFunc {
	if !strings.HasSuffix(urlPrefix, "/") {
		urlPrefix += "/"
	}

	return func(w http.ResponseWriter, r *http.Request) {
		// Ensure the request path starts with our prefix
		if !strings.HasPrefix(r.URL.Path, urlPrefix) {
			http.NotFound(w, r)
			return
		}

		rel := strings.TrimPrefix(r.URL.Path, urlPrefix)
		if rel == "" || strings.HasSuffix(rel, "/") {
			http.NotFound(w, r)
			return
		}

		// Clean to avoid ".." path traversal
		rel = filepath.Clean(rel)
		if strings.Contains(rel, "..") {
			http.NotFound(w, r)
			return
		}

		http.ServeFile(w, r, filepath.Join(dir, rel))
	}
}
