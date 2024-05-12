// Copyright 2024 Bill Nixon. All rights reserved.
// Use of this source code is governed by the license found in the LICENSE file.

package webapp

import (
	"log/slog"
	"net/http"

	"github.com/bnixon67/webapp/webhandler"
	"github.com/bnixon67/webapp/webutil"
)

// RootPageName specifies the template file for the root page.
const RootPageName = "root.html"

// RootPageData encapsulates data to be passed to the root page template.
type RootPageData struct {
	Title string // Title of the page.
}

// RootHandlerGet handles GET requests to the root ("/") route.
func (app *WebApp) RootHandlerGet(w http.ResponseWriter, r *http.Request) {
	logger := webhandler.NewRequestLoggerWithFuncName(r)

	if !webutil.IsMethodOrError(w, r, http.MethodGet) {
		logger.Error("invalid method")
		return
	}

	// Ensure the request is to the exact root path.
	if r.URL.Path != "/" {
		logger.Error("invalid path")
		http.NotFound(w, r)
		return
	}

	data := RootPageData{Title: app.Config.App.Name}

	err := webutil.RenderTemplateOrError(app.Tmpl, w, RootPageName, data)
	if err != nil {
		logger.Error("failed to RenderTemplate", "err", err)
		return
	}

	logger.Info("done", slog.Any("data", data))
}
