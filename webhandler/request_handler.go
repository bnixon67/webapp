// Copyright 2023 Bill Nixon. All rights reserved.
// Use of this source code is governed by the license found in the LICENSE file.

package webhandler

import (
	"fmt"
	"log/slog"
	"net/http"
	"net/http/httputil"

	"github.com/bnixon67/webapp/webutil"
)

// RequestHandler responds with a dump of the HTTP request.
func (h *Handler) RequestHandler(w http.ResponseWriter, r *http.Request) {
	// Get the logger from the request context and add calling function name.
	logger := LoggerFromContext(r.Context()).With(slog.String("func", FuncName()))

	// Check if the HTTP method is valid.
	if !webutil.ValidMethod(w, r, http.MethodGet) {
		logger.Error("invalid method")
		return
	}

	// Get the request information.
	b, err := httputil.DumpRequest(r, true)
	if err != nil {
		http.Error(w, fmt.Sprintf("Error:\n%v\n", err), http.StatusInternalServerError)
		logger.Error("failed to DumpRequest", "err", err)
		return
	}

	// Log that the handler is executing.
	logger.Debug("response", slog.Any("request", string(b)))

	// Set the content type to plain text.
	webutil.SetTextContentType(w)

	// Set no-cache headers to prevent caching.
	webutil.SetNoCacheHeaders(w)

	// Write the "hello" message to the response with the application name.
	fmt.Fprintln(w, string(b))
}
