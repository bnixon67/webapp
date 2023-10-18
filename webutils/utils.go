// Package webutils provides utility functions.
package webutils

import (
	"net/http"
	"slices"
	"strings"
)

// ValidMethod checks if the HTTP method of request is one of the allowed methods.
// If the method is not allowed, it sets the appropriate headers, HTTP status and writes an error response.
// It returns true if the method is allowed, false otherwise.
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
