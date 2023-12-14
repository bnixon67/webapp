// Copyright 2023 Bill Nixon. All rights reserved.
// Use of this source code is governed by the license found in the LICENSE file.

package main

import (
	"context"
	"fmt"
	"html/template"
	"log/slog"
	"net/http"
	"os"
	"path/filepath"

	"github.com/bnixon67/webapp/assets"
	"github.com/bnixon67/webapp/webapp"
	"github.com/bnixon67/webapp/webhandler"
	"github.com/bnixon67/webapp/weblog"
	"github.com/bnixon67/webapp/webserver"
	"github.com/bnixon67/webapp/webutil"
)

const (
	ExitUsage    = iota + 1 // ExitUsage indicates a usage error.
	ExitConfig              // ExitConfig indicates a config error.
	ExitLog                 // ExitLog indicates a log error.
	ExitHandler             // ExitHandler indicates a handler error.
	ExitServer              // ExitServer indicates a server error.
	ExitTemplate            // ExitTemplate indicates a template error.
)

func main() {
	// Check for command line argument with config file.
	if len(os.Args) != 2 {
		fmt.Fprintf(os.Stderr, "Usage: %s [config file]\n", os.Args[0])
		os.Exit(ExitUsage)
	}

	// Read config.
	cfg, err := webapp.ConfigFromJSONFile(os.Args[1])
	if err != nil {
		fmt.Fprintln(os.Stderr, "Failed to get config:", err)
		os.Exit(ExitConfig)
	}

	// Initialize logging.
	err = weblog.Init(
		weblog.WithFilename(cfg.Log.Filename),
		weblog.WithLogType(cfg.Log.Type),
		weblog.WithLevel(cfg.Log.Level),
		weblog.WithSource(cfg.Log.WithSource),
	)
	if err != nil {
		fmt.Fprintln(os.Stderr, "Error initializing logger:", err)
		os.Exit(ExitLog)
	}

	// Define the custom function
	funcMap := template.FuncMap{
		"ToTimeZone": webutil.ToTimeZone,
		"Join":       webutil.Join,
	}

	// Initialize templates
	if cfg.AssetsDir == "" {
		cfg.AssetsDir = assets.AssetPath()
	}
	tmpl, err := webutil.InitTemplatesWithFuncMap(filepath.Join(cfg.AssetsDir, "tmpl", "*.html"), funcMap)
	if err != nil {
		fmt.Fprintln(os.Stderr, "Error initializing templates:", err)
		os.Exit(ExitTemplate)
	}

	// Create the web app.
	app, err := webapp.New(webapp.WithAppName(cfg.Name), webapp.WithTemplate(tmpl))
	if err != nil {
		fmt.Fprintln(os.Stderr, "Error creating new handler:", err)
		os.Exit(ExitHandler)
	}

	assetDir := assets.AssetPath()
	cssFile := filepath.Join(assetDir, "css", "w3.css")
	icoFile := filepath.Join(assetDir, "ico", "favicon.ico")

	// Create a new ServeMux to handle HTTP requests.
	mux := http.NewServeMux()

	// Add middleware to mux.
	// Functions are executed in reverse, so last added is called first.
	h := webhandler.AddSecurityHeaders(mux)
	h = webhandler.LogRequest(h)
	h = webhandler.AddLogger(h)
	h = webhandler.AddRequestID(h)

	mux.HandleFunc("/", app.RootHandler)
	mux.HandleFunc("/w3.css", webutil.ServeFileHandler(cssFile))
	mux.HandleFunc("/favicon.ico", webutil.ServeFileHandler(icoFile))
	mux.HandleFunc("/hello", app.HelloTextHandler)
	mux.HandleFunc("/hellohtml", app.HelloHTMLHandler)
	mux.HandleFunc("/build", app.BuildHandler)
	mux.HandleFunc("/headers", app.HeadersHandler)
	mux.HandleFunc("/remote", webhandler.RemoteHandler)
	mux.HandleFunc("/request", webhandler.RequestHandler)

	// Create the web server.
	srv, err := webserver.New(
		webserver.WithAddr(cfg.Server.Host+":"+cfg.Server.Port),
		webserver.WithHandler(h),
		webserver.WithTLS(cfg.Server.CertFile, cfg.Server.KeyFile),
	)
	if err != nil {
		fmt.Fprintln(os.Stderr, "Error creating server:", err)
		os.Exit(ExitServer)
	}

	// Create a new context.
	ctx := context.Background()

	// Start the web server.
	err = srv.Start(ctx)
	if err != nil {
		slog.Error("error running server", "err", err)
		fmt.Fprintln(os.Stderr, "Error running server:", err)
		os.Exit(ExitServer)
	}
}
