// Copyright 2024 Bill Nixon. All rights reserved.
// Use of this source code is governed by the license found in the LICENSE file.

// Package webutils provides utility functions.
package webutil

import (
	"net/http"
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
