// Copyright 2023 Bill Nixon. All rights reserved.
// Use of this source code is governed by the license found in the LICENSE file.

package main

import (
	"context"
	"fmt"
	"html/template"
	"log"
	"log/slog"
	"net/http"
	_ "net/http/pprof"
	"os"
	"path/filepath"

	"github.com/bnixon67/webapp/assets"
	"github.com/bnixon67/webapp/webapp"
	"github.com/bnixon67/webapp/webhandler"
	"github.com/bnixon67/webapp/weblog"
	"github.com/bnixon67/webapp/webserver"
	"github.com/bnixon67/webapp/websse"
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
	go func() {
		log.Println(http.ListenAndServe("localhost:6060", nil))
	}()

	// Check command line for config file.
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
		weblog.WithType(cfg.Log.Type),
		weblog.WithLevel(cfg.Log.Level),
		weblog.WithSource(cfg.Log.WithSource),
	)
	if err != nil {
		fmt.Fprintln(os.Stderr, "Error initializing logger:", err)
		os.Exit(ExitLog)
	}

	// Get directory for assets, using a default if not specified in config.
	if cfg.App.AssetsDir == "" {
		cfg.App.AssetsDir = assets.AssetPath()
	}
	assetsDir := cfg.App.AssetsDir

	// Show config in log.
	slog.Info("using config", slog.Any("config", cfg))

	// Define custom template functions.
	funcMap := template.FuncMap{
		"ToTimeZone": webutil.ToTimeZone,
		"Join":       webutil.Join,
	}

	// Parse templates.
	pattern := filepath.Join(assetsDir, "tmpl", "*.html")
	tmpl, err := webutil.TemplatesWithFuncs(pattern, funcMap)
	if err != nil {
		fmt.Fprintln(os.Stderr, "Error initializing templates:", err)
		os.Exit(ExitTemplate)
	}

	// Create the web app.
	app, err := webapp.New(
		webapp.WithName(cfg.App.Name), webapp.WithTemplate(tmpl))
	if err != nil {
		fmt.Fprintln(os.Stderr, "Error creating new handler:", err)
		os.Exit(ExitHandler)
	}

	cssFile := filepath.Join(assetsDir, "css", "w3.css")
	icoFile := filepath.Join(assetsDir, "ico", "webapp.ico")

	// Create a new ServeMux to handle HTTP requests.
	mux := http.NewServeMux()

	// Add middleware to mux.
	// Functions are executed in reverse, so last added is called first.
	h := webhandler.LogRequest(mux)
	h = webhandler.AddLogger(h)
	h = webhandler.AddRequestID(h)

	sseServer := websse.NewServer()

	mux.HandleFunc("/", app.RootHandler)
	mux.HandleFunc("/w3.css", webutil.ServeFileHandler(cssFile))
	mux.HandleFunc("/favicon.ico", webutil.ServeFileHandler(icoFile))
	mux.HandleFunc("/event", sseServer.EventStreamHandler)
	mux.HandleFunc("/send", sseServer.SendMessageHandler)

	// Create the web server.
	srv, err := webserver.New(
		webserver.WithAddr(cfg.Server.Host+":"+cfg.Server.Port),
		webserver.WithHandler(h),
		webserver.WithTLS(cfg.Server.CertFile, cfg.Server.KeyFile),
		webserver.WithWriteTimeout(0),
	)
	if err != nil {
		fmt.Fprintln(os.Stderr, "Error creating server:", err)
		os.Exit(ExitServer)
	}

	// Run the SSE server.
	sseServer.Run()

	// Create a new context.
	ctx := context.Background()

	// Run the web server.
	err = srv.Run(ctx)
	if err != nil {
		slog.Error("web server error", slog.Any("err", err))
		fmt.Fprintln(os.Stderr, "Error running web server:", err)
		os.Exit(ExitServer)
	}
}
