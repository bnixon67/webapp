// Copyright 2023 Bill Nixon. All rights reserved.
// Use of this source code is governed by the license found in the LICENSE file.

package webapp

import (
	"log/slog"
	"net/http"

	"github.com/bnixon67/webapp/webhandler"
	"github.com/bnixon67/webapp/webutil"
)

// RootPageName is the name of the HTTP template to execute.
const RootPageName = "root.html"

// RootPageData holds the data passed to the HTML template.
type RootPageData struct {
	Title string // Title of the page.
}

// RootHandler handles the root ("/") route.
func (app *WebApp) RootHandler(w http.ResponseWriter, r *http.Request) {
	// Get logger with request info and function name.
	logger := webhandler.GetRequestLoggerWithFunc(r)

	// Check if the HTTP method is valid.
	if !webutil.ValidMethod(w, r, http.MethodGet) {
		logger.Error("invalid method")
		return
	}

	// Check for valid URL path.
	if r.URL.Path != "/" {
		logger.Error("invalid path")
		http.NotFound(w, r)
		return
	}

	// Prepare the data for rendering the template.
	data := RootPageData{
		Title: "Request Headers",
	}

	// Render the template with the data.
	err := webutil.RenderTemplate(app.Tmpl, w, RootPageName, data)
	if err != nil {
		logger.Error("failed to RenderTemplate", "err", err)
		return
	}

	logger.Info("success", slog.Any("data", data))
}
