// Copyright 2024 Bill Nixon. All rights reserved.
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

// RootHandlerGet handles the root ("/") route.
func (app *WebApp) RootHandlerGet(w http.ResponseWriter, r *http.Request) {
	logger := webhandler.RequestLoggerWithFunc(r)

	if !webutil.IsMethodValid(w, r, http.MethodGet) {
		logger.Error("invalid method")
		return
	}

	// Check for valid URL path.
	if r.URL.Path != "/" {
		logger.Error("invalid path")
		http.NotFound(w, r)
		return
	}

	data := RootPageData{Title: app.Config.App.Name}

	err := webutil.RenderTemplate(app.Tmpl, w, RootPageName, data)
	if err != nil {
		logger.Error("failed to RenderTemplate", "err", err)
		return
	}

	logger.Info("done", slog.Any("data", data))
}
