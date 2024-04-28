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

// IsMethodValid verifies if the HTTP request method matches the specified
// method. It returns true if they match. Otherwise, it sends a 405 Method
// Not Allowed response and returns false. The caller should stop further
// processing if false is returned.
func IsMethodValid(w http.ResponseWriter, r *http.Request, method string) bool {
	if r.Method != method {
		RespondWithError(w, http.StatusMethodNotAllowed)
		return false
	}
	return true
}

// CheckAllowedMethods validates the request's method against a list of
// allowed methods. It automatically supports the OPTIONS method. If the
// method is allowed, it returns true. If the method is not allowed, it sets
// the 'Allow' header, responds appropriately, and returns false. The caller
// is advised to halt further processing if false is returned.
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

// SetNoCacheHeaders sets headers for client to not cache the response content.
func SetNoCacheHeaders(w http.ResponseWriter) {
	w.Header().Set("Cache-Control", "no-cache, no-store, must-revalidate")
	w.Header().Set("Pragma", "no-cache")
	w.Header().Set("Expires", "0")
}

// SetContentType sets headers for client to interpret response as contentType.
func SetContentType(w http.ResponseWriter, contentType string) {
	w.Header().Set("Content-Type", contentType)
	w.Header().Set("X-Content-Type-Options", "nosniff")
}

// SetContentTypeText sets headers for client to interpret response as plain text.
func SetContentTypeText(w http.ResponseWriter) {
	SetContentType(w, "text/plain;charset=utf-8")
}

// SetContentTypeHTML sets headers for client to interpret response as HTML.
func SetContentTypeHTML(w http.ResponseWriter) {
	SetContentType(w, "text/html;charset=utf-8")
}

// ClientIP retrieves the client's IP address from the request. It prioritizes
// the X-Real-IP header value if present; otherwise, it falls back to the
// request's RemoteAddr.
func ClientIP(r *http.Request) string {
	clientIP := r.Header.Get("X-Real-IP")
	if clientIP == "" {
		clientIP = r.RemoteAddr
	}
	return clientIP
}

// ServeFileHandler creates an HTTP handler function that serves a specified
// file. It verifies the file's existence and accessibility before creating
// the handler. Returns nil if the file does not exist or is not accessible.
func ServeFileHandler(filePath string) http.HandlerFunc {
	// Check if the file exists and is accessible
	_, err := os.Stat(filePath)
	if err != nil {
		slog.Error("does not exist or not accessible", "file", filePath)
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
