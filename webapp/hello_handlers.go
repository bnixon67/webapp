// Copyright 2024 Bill Nixon. All rights reserved.
// Use of this source code is governed by the license found in the LICENSE file.

package webapp

import (
	"fmt"
	"net/http"

	"github.com/bnixon67/webapp/assets"
	"github.com/bnixon67/webapp/webhandler"
	"github.com/bnixon67/webapp/webutil"
)

// HelloTextHandlerGet responds with a hello message in plain text format.
func (app *WebApp) HelloTextHandlerGet(w http.ResponseWriter, r *http.Request) {
	logger := webhandler.RequestLoggerWithFunc(r)

	if !webutil.IsMethodValid(w, r, http.MethodGet) {
		logger.Error("invalid method")
		return
	}

	webutil.SetNoCacheHeaders(w)
	webutil.SetContentTypeText(w)

	_, err := fmt.Fprintln(w, "hello from", app.Config.App.Name)
	if err != nil {
		logger.Error("failed to write response", "err", err)
		return
	}

	logger.Info("done")
}

// HelloHTMLHandlerGet responds with a hello message in HTML format.
func (app *WebApp) HelloHTMLHandlerGet(w http.ResponseWriter, r *http.Request) {
	logger := webhandler.RequestLoggerWithFunc(r)

	if !webutil.IsMethodValid(w, r, http.MethodGet) {
		logger.Error("invalid method")
		return
	}

	webutil.SetNoCacheHeaders(w)
	webutil.SetContentTypeHTML(w)

	// Write the HTML content to the response from the assets package.
	_, err := fmt.Fprint(w, assets.HelloHTML)
	if err != nil {
		logger.Error("failed to write response", "err", err)
		return
	}

	logger.Info("done")
}
