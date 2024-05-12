// Copyright 2023 Bill Nixon. All rights reserved.
// Use of this source code is governed by the license found in the LICENSE file.

// Package webutils provides utility functions.
package webutil

import (
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"slices"
	"strings"
)

// IsMethodOrError checks if the HTTP request method matches the specified
// method and returns true if they match. Otherwise, it sends a 405 Method
// Not Allowed response and returns false. The caller should stop further
// processing if false is returned.
func IsMethodOrError(w http.ResponseWriter, r *http.Request, method string) bool {
	if r.Method != method {
		RespondWithError(w, http.StatusMethodNotAllowed)
		return false
	}
	return true
}

// CheckAllowedMethods validates the request's method against a list of
// allowed methods, automatically including the OPTIONS method.
// It sets the 'Allow' header and sends appropriate responses.
// The caller should stop further processing if false is returned.
func CheckAllowedMethods(w http.ResponseWriter, r *http.Request, allowed ...string) bool {
	if slices.Contains(allowed, r.Method) {
		return true
	}

	// Append OPTIONS to list of allowed methods for compliance.
	allowed = append(allowed, http.MethodOptions)

	// Inform the client about the allowed methods.
	w.Header().Set("Allow", strings.Join(allowed, ", "))
	if r.Method == http.MethodOptions {
		w.WriteHeader(http.StatusNoContent)
		return false
	}

	// Method is not allowed and not OPTIONS.
	txt := r.Method + " " + http.StatusText(http.StatusMethodNotAllowed)
	http.Error(w, txt, http.StatusMethodNotAllowed)
	return false
}

func setHeaders(w http.ResponseWriter, headers map[string]string) {
	for key, value := range headers {
		w.Header().Set(key, value)
	}
}

// SetNoCacheHeaders instructs the client to not cache the response.
func SetNoCacheHeaders(w http.ResponseWriter) {
	headers := map[string]string{
		"Cache-Control": "no-cache, no-store, must-revalidate",
		"Pragma":        "no-cache",
		"Expires":       "0",
	}

	setHeaders(w, headers)
}

// SetContentType configures the 'Content-Type' and related headers.
func SetContentType(w http.ResponseWriter, contentType string) {
	headers := map[string]string{
		"Content-Type":           contentType,
		"X-Content-Type-Options": "nosniff",
	}

	setHeaders(w, headers)
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

// RespondWithError sends an HTTP response with the specified error code and
// a corresponding error message.
func RespondWithError(w http.ResponseWriter, code int) {
	message := fmt.Sprintf("Error: %s", http.StatusText(code))
	http.Error(w, message, code)
}
