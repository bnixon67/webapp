// Copyright 2023 Bill Nixon. All rights reserved.
// Use of this source code is governed by the license found in the LICENSE file.

package webhandler

import (
	"fmt"
	"log/slog"
	"net/http"

	"github.com/bnixon67/webapp/webutil"
)

// RemoteGetHandler responds with RemoteAddr and common headers for the
// actual RemoteAddr if a proxy, load balancer, or similar is used to route
// the request.
func RemoteGetHandler(w http.ResponseWriter, r *http.Request) {
	// Get logger with request info from request context and add calling function name.
	logger := ExtractLogger(r.Context()).With(slog.String("func", FuncName()))

	// Check if the HTTP method is valid.
	if !webutil.CheckAllowedMethods(w, r, http.MethodGet) {
		logger.Error("invalid method")
		return
	}

	logger.Debug("response")

	// Set the content type of the response to text.
	webutil.SetContentTypeText(w)

	// Set no-cache headers to prevent caching of the response.
	webutil.SetNoCacheHeaders(w)

	// Write the RemoteAddr from the Request.
	fmt.Fprintf(w, "RemoteAddr: %v\n", r.RemoteAddr)

	// Common headers that may contain the actual remote address.
	headers := []string{
		"Cf-Connecting-Ip",
		"X-Client-Ip",
		"X-Forwarded-For",
		"X-Real-Ip",
	}

	// Write common headers that may contain the actual remote address.
	for _, header := range headers {
		val := r.Header.Get(header)
		if val != "" {
			fmt.Fprintf(w, "%s: %v\n", header, val)
		}
	}
}
