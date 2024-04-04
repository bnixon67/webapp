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

// IsMethodValid checks if the request's method matches the specified
// method. It return true if the method matches; otherwise, it responds with
// StatusMethodNotAllowed (405) and returns false. The caller is responsible
// for not proceeding with further writes to w if false is returned.
func IsMethodValid(w http.ResponseWriter, r *http.Request, method string) bool {
	if r.Method != method {
		RespondWithError(w, http.StatusMethodNotAllowed)
		return false
	}
	return true
}

// ValidMethod checks the HTTP method of the request against allowed methods.
//
// Returns true if the method is in allowed methods list, false otherwise.
//
// It adds the OPTIONS method, so clients can determine which methods are valid.
//
// If the method is not allowed,
//   - It sets the 'Allow' header wth the allowed methods.
//   - For methods not allowed, it responds with StatusMethodNotAllowed (405).
//   - For OPTIONS, it respondes with a StatusNoContent (204).
//
// Note: When false is returned, it does not otherwise end the request;
// the caller should ensure no further writes are done to w.
func ValidMethod(w http.ResponseWriter, r *http.Request, allowed ...string) bool {
	// Check if the request's method is in the list of allowed methods.
	if slices.Contains(allowed, r.Method) {
		return true
	}

	// Append OPTIONS method to allowed methods to adhere to HTTP standard.
	allowed = append(allowed, http.MethodOptions)

	// Set the 'Allow' header to inform client about allowed methods.
	w.Header().Set("Allow", strings.Join(allowed, ", "))

	// If request's method is OPTIONS, respond with list of allowed methods.
	if r.Method == http.MethodOptions {
		w.WriteHeader(http.StatusNoContent)
		return false
	}

	// Request's method is not allowed and not OPTIONS.
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

// SetTextContentType sets headers for client to interpret response as plain text.
func SetTextContentType(w http.ResponseWriter) {
	w.Header().Set("Content-Type", "text/plain; charset=utf-8")
	w.Header().Set("X-Content-Type-Options", "nosniff")
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
