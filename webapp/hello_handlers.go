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

// HelloTextHandlerGet responds with a hello message in plain text format.
func (app *WebApp) HelloTextHandlerGet(w http.ResponseWriter, r *http.Request) {
	logger := webhandler.RequestLoggerWithFunc(r)

	if !webutil.EnforceMethod(w, r, http.MethodGet) {
		logger.Error("invalid method")
		return
	}

	webutil.SetNoCacheHeaders(w)

	webutil.SetTextContentType(w)
	fmt.Fprintln(w, "hello from", app.Name)

	logger.Info("done")
}

// HelloHTMLHandlerGet responds with a hello message in HTML format.
func (app *WebApp) HelloHTMLHandlerGet(w http.ResponseWriter, r *http.Request) {
	logger := webhandler.RequestLoggerWithFunc(r)

	if !webutil.EnforceMethod(w, r, http.MethodGet) {
		logger.Error("invalid method")
		return
	}

	webutil.SetNoCacheHeaders(w)

	// Write the HTML content to the response from the assets package.
	fmt.Fprint(w, assets.HelloHTML)

	logger.Info("done")
}
