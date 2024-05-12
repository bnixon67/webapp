// Copyright 2024 Bill Nixon. All rights reserved.
// Use of this source code is governed by the license found in the LICENSE file.

// Package webutils provides utility functions.
package webutil

import (
	"fmt"
	"net/http"
)

// RespondWithError sends an HTTP response with the specified error code and
// a corresponding error message.
func RespondWithError(w http.ResponseWriter, code int) {
	message := fmt.Sprintf("Error: %s", http.StatusText(code))
	http.Error(w, message, code)
}
