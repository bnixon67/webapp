// Copyright 2024 Bill Nixon. All rights reserved.
// Use of this source code is governed by the license found in the LICENSE file.

// Package webutils provides utility functions.
package webutil

import (
	"net/http"
)

// SetHeaders applies headers to the HTTP response.
func SetHeaders(w http.ResponseWriter, headers map[string]string) {
	for key, value := range headers {
		w.Header().Set(key, value)
	}
}

// SetNoCacheHeaders instructs the client not to cache the response.
func SetNoCacheHeaders(w http.ResponseWriter) {
	headers := map[string]string{
		"Cache-Control": "no-cache, no-store, must-revalidate",
		"Pragma":        "no-cache",
		"Expires":       "0",
	}

	SetHeaders(w, headers)
}

// SetContentType sets the 'Content-Type' and associated security headers.
func SetContentType(w http.ResponseWriter, contentType string) {
	headers := map[string]string{
		"Content-Type":           contentType,
		"X-Content-Type-Options": "nosniff",
	}

	SetHeaders(w, headers)
}

// SetContentTypeText sets HTTP headers for plain text responses.
func SetContentTypeText(w http.ResponseWriter) {
	SetContentType(w, "text/plain;charset=utf-8")
}

// SetContentTypeHTML sets HTTP headers for HTML responses.
func SetContentTypeHTML(w http.ResponseWriter) {
	SetContentType(w, "text/html;charset=utf-8")
}

// ClientIP retrieves the client's IP address, preferring the X-Real-IP header.
func ClientIP(r *http.Request) string {
	if ip := r.Header.Get("X-Real-IP"); ip != "" {
		return ip
	}
	return r.RemoteAddr
}
