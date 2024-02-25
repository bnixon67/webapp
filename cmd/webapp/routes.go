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
	if app.AssetsDir == "" {
		app.AssetsDir = assets.AssetPath()
	}
	assetsDir := app.AssetsDir

	cssFile := filepath.Join(assetsDir, "css", "w3.css")
	icoFile := filepath.Join(assetsDir, "ico", "webapp.ico")

	mux.HandleFunc("/", app.RootHandler)
	mux.HandleFunc("/w3.css", webutil.ServeFileHandler(cssFile))
	mux.HandleFunc("/favicon.ico", webutil.ServeFileHandler(icoFile))
	mux.HandleFunc("/hello", app.HelloTextHandler)
	mux.HandleFunc("/hellohtml", app.HelloHTMLHandler)
	mux.HandleFunc("/build", app.BuildHandler)
	mux.HandleFunc("/headers", app.HeadersHandler)
	mux.HandleFunc("/remote", webhandler.RemoteHandler)
	mux.HandleFunc("/request", webhandler.RequestHandler)
}

func AddMiddleware(h http.Handler) http.Handler {
	// Functions are executed in reverse, so last added is called first.
	h = webhandler.LogRequest(h)
	h = webhandler.AddLogger(h)
	h = webhandler.AddRequestID(h)

	return h
}