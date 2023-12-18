// Copyright 2023 Bill Nixon. All rights reserved.
// Use of this source code is governed by the license found in the LICENSE file.

package webapp

import (
	"fmt"
	"net/http"

	"github.com/bnixon67/webapp/assets"
	"github.com/bnixon67/webapp/webhandler"
	"github.com/bnixon67/webapp/webutil"
)

// HelloTextHandler responds with a simple "hello" message in plain text format.
func (app *WebApp) HelloTextHandler(w http.ResponseWriter, r *http.Request) {
	// Get logger with request info and function name.
	logger := webhandler.GetRequestLoggerWithFunc(r)

	// Check if the HTTP method is valid.
	if !webutil.ValidMethod(w, r, http.MethodGet) {
		logger.Error("invalid method")
		return
	}

	// Set the content type to plain text.
	webutil.SetTextContentType(w)

	// Set no-cache headers to prevent caching.
	webutil.SetNoCacheHeaders(w)

	// Write the "hello" message to the response with the application name.
	fmt.Fprintln(w, "hello from", app.Name)

	// Log sucecss of the handler.
	logger.Info("success")
}

// HelloHTMLHandler responds with a simple "hello" message in HTML format.
func (app *WebApp) HelloHTMLHandler(w http.ResponseWriter, r *http.Request) {
	// Get logger with request info and function name.
	logger := webhandler.GetRequestLoggerWithFunc(r)

	// Check if the HTTP method is valid.
	if !webutil.ValidMethod(w, r, http.MethodGet) {
		logger.Error("invalid method")
		return
	}

	// Log that the handler is executing.
	logger.Debug("response")

	// Set no-cache headers to prevent caching.
	webutil.SetNoCacheHeaders(w)

	// Write the HTML content to the response from the assets package.
	fmt.Fprint(w, assets.HelloHTML)

	// Log sucecss of the handler.
	logger.Info("success")
}
