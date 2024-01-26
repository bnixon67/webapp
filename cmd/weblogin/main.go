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

	_ "github.com/go-sql-driver/mysql"

	"github.com/bnixon67/webapp/assets"
	"github.com/bnixon67/webapp/webapp"
	"github.com/bnixon67/webapp/webhandler"
	"github.com/bnixon67/webapp/weblog"
	"github.com/bnixon67/webapp/weblogin"
	"github.com/bnixon67/webapp/webserver"
	"github.com/bnixon67/webapp/webutil"
)

const (
	ExitUsage    = iota + 1 // ExitUsage indicates a usage error.
	ExitLog                 // ExitLog indicates a log error.
	ExitServer              // ExitServer indicates a server error.
	ExitTemplate            // ExitTemplate indicates a template error.
	ExitConfig              // ExitConfig indicates a config error.
	ExitDB                  // ExitConfig indicates a database error.
	ExitApp                 // ExitHandler indicates an app error.
)

func main() {
	// Check for command line argument with config file.
	if len(os.Args) != 2 {
		fmt.Fprintf(os.Stderr, "Usage: %s [config file]\n", os.Args[0])
		os.Exit(ExitUsage)
	}

	// Read config.
	cfg, err := weblogin.ConfigFromJSONFile(os.Args[1])
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

	// Define the custom function
	funcMap := template.FuncMap{
		"ToTimeZone": webutil.ToTimeZone,
		"Join":       webutil.Join,
	}

	// Initialize templates
	tmpl, err := webutil.TemplatesWithFuncs(cfg.ParseGlobPattern, funcMap)
	if err != nil {
		fmt.Fprintln(os.Stderr, "Error initializing templates:", err)
		os.Exit(ExitTemplate)
	}

	// Initialize db
	db, err := weblogin.InitDB(cfg.SQL.DriverName, cfg.SQL.DataSourceName)
	if err != nil {
		fmt.Fprintln(os.Stderr, "failed to initialize database:", err)
		os.Exit(ExitDB)
	}

	// Create the web login app.
	app, err := weblogin.New(
		webapp.WithName(cfg.App.Name), webapp.WithTemplate(tmpl),
		weblogin.WithConfig(cfg), weblogin.WithDB(db),
	)
	if err != nil {
		fmt.Fprintln(os.Stderr, "failed to create new weblogin:", err)
		os.Exit(ExitApp)
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

	mux.Handle("/",
		http.RedirectHandler("/user", http.StatusMovedPermanently))
	mux.HandleFunc("/w3.css", webutil.ServeFileHandler(cssFile))
	mux.HandleFunc("/favicon.ico", webutil.ServeFileHandler(icoFile))
	mux.HandleFunc("/user", app.UserHandler)
	mux.HandleFunc("/login", app.LoginHandler)
	mux.HandleFunc("/logout", app.LogoutHandler)
	mux.HandleFunc("/register", app.RegisterHandler)
	mux.HandleFunc("/users", app.UsersHandler)
	mux.HandleFunc("/userscsv", app.UsersCSVHandler)
	mux.HandleFunc("/forgot", app.ForgotHandler)
	mux.HandleFunc("/confirm_request", app.ConfirmRequestHandler)
	mux.HandleFunc("/confirm", app.ConfirmHandler)
	mux.HandleFunc("/reset", app.ResetHandler)
	mux.HandleFunc("/events", app.EventsHandler)
	mux.HandleFunc("/eventscsv", app.EventsCSVHandler)

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
	err = srv.Run(ctx)
	if err != nil {
		slog.Error("error running server", "err", err)
		fmt.Fprintln(os.Stderr, "Error running server:", err)
		os.Exit(ExitServer)
	}
}
