// Package webutils provides utility functions.
package webutil

import (
	"net/http"
	"slices"
	"strings"
)

// ValidMethod checks if the HTTP method of the request is one of the allowed methods.
// Returns true if the method is allowed, false otherwise.
// If the method is not allowed, w is updated with appropriate headers, HTTP status, and error message in the body. It does not otherwise end the request; the caller should ensure no further writes are done to w.
func ValidMethod(w http.ResponseWriter, r *http.Request, allowed ...string) bool {
	// Check if the request's method is in the list of allowed methods.
	if slices.Contains(allowed, r.Method) {
		return true
	}

	// Append OPTIONS method to allowed methods to adhere to the HTTP standard.
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

func RealRemoteAddr(r *http.Request) string {
	realIP := r.Header.Get("X-Real-IP")
	if realIP == "" {
		realIP = r.RemoteAddr
	}

	return realIP
}
