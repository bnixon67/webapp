// Copyright 2023 Bill Nixon. All rights reserved.
// Use of this source code is governed by the license found in the LICENSE file.

// Package webutils provides utility functions.
package webutil

import (
	"net/http"
)

func addHeaders(w http.ResponseWriter, headers map[string]string) {
	for key, value := range headers {
		w.Header().Add(key, value)
	}
}

// SetNoCacheHeaders instructs the client to not cache the response.
func SetNoCacheHeaders(w http.ResponseWriter) {
	headers := map[string]string{
		"Cache-Control": "no-cache, no-store, must-revalidate",
		"Pragma":        "no-cache",
		"Expires":       "0",
	}

	addHeaders(w, headers)
}

// SetContentType configures the 'Content-Type' and related headers.
func SetContentType(w http.ResponseWriter, contentType string) {
	headers := map[string]string{
		"Content-Type":           contentType,
		"X-Content-Type-Options": "nosniff",
	}

	addHeaders(w, headers)
}

// SetContentTypeText sets headers for client to interpret response as plain text.
func SetContentTypeText(w http.ResponseWriter) {
	SetContentType(w, "text/plain;charset=utf-8")
}

// SetContentTypeHTML sets headers for client to interpret response as HTML.
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
