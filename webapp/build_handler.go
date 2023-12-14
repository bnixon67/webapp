// Copyright 2023 Bill Nixon. All rights reserved.
// Use of this source code is governed by the license found in the LICENSE file.

package webapp

import (
	"fmt"
	"log/slog"
	"net/http"

	"github.com/bnixon67/webapp/webhandler"
	"github.com/bnixon67/webapp/webutil"
)

// BuildDateTimeFormat can be used to format a time as "YYYY-MM-DD HH:MM:SS"
const BuildDateTimeFormat = "2006-01-02 15:04:05"

// BuildHandler responds with the executable modification date and time.
func (app *WebApp) BuildHandler(w http.ResponseWriter, r *http.Request) {
	// Get logger with request info and function name.
	logger := webhandler.GetRequestLoggerWithFunc(r)

	// Check if the HTTP method is valid.
	if !webutil.ValidMethod(w, r, http.MethodGet) {
		logger.Error("invalid method")
		return
	}

	// Set no-cache headers to prevent caching of the response.
	webutil.SetNoCacheHeaders(w)

	// Format the time as a string.
	build := app.BuildDateTime.Format(BuildDateTimeFormat)

	// Set the content type of the response to text.
	webutil.SetTextContentType(w)

	// Write the build time to the response.
	fmt.Fprintln(w, build)

	// Log success of the handler.
	logger.Info("success", slog.String("build", build))
}
