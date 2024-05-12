// Copyright 2023 Bill Nixon. All rights reserved.
// Use of this source code is governed by the license found in the LICENSE file.

package webhandler

import (
	"net/http"
)

// LogRequest is a middleware function that logs the details of incoming
// HTTP requests. It enhances HTTP handlers by logging each request before
// passing control to the next handler.
func LogRequest(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		NewRequestLogger(r).Info("HTTP request received")

		next.ServeHTTP(w, r)
	})
}
