// Copyright 2024 Bill Nixon. All rights reserved.
// Use of this source code is governed by the license found in the LICENSE file.

package main

import (
	"net/http"
	"path/filepath"

	"github.com/bnixon67/webapp/assets"
	"github.com/bnixon67/webapp/webapp"
	"github.com/bnixon67/webapp/webhandler"
	"github.com/bnixon67/webapp/webutil"
)

func AddRoutes(mux *http.ServeMux, app *webapp.WebApp) {

	// Get directory for assets, using a default if not specified in config.
	if app.Config.App.AssetsDir == "" {
		app.Config.App.AssetsDir = assets.AssetPath()
	}
	assetsDir := app.Config.App.AssetsDir

	cssFile := filepath.Join(assetsDir, "css", "w3.css")
	icoFile := filepath.Join(assetsDir, "ico", "webapp.ico")

	mux.HandleFunc("GET /w3.css", webutil.ServeFileHandler(cssFile))
	mux.HandleFunc("GET /favicon.ico", webutil.ServeFileHandler(icoFile))
	mux.HandleFunc("GET /hello", app.HelloTextHandlerGet)
	mux.HandleFunc("GET /hellohtml", app.HelloHTMLHandlerGet)
	mux.HandleFunc("GET /build", app.BuildHandlerGet)
	mux.HandleFunc("GET /headers", app.HeadersHandlerGet)
	mux.HandleFunc("GET /remote", webhandler.RemoteHandler)
	mux.HandleFunc("GET /request", webhandler.RequestHandler)
	mux.HandleFunc("GET /", app.RootHandlerGet)
}

func AddMiddleware(h http.Handler) http.Handler {
	// Functions are executed in reverse, so last added is called first.
	h = webhandler.LogRequest(h)
	h = webhandler.AddLogger(h)
	h = webhandler.AddRequestID(h)

	return h
}
