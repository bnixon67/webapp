// Copyright 2023 Bill Nixon. All rights reserved.
// Use of this source code is governed by the license found in the LICENSE file.

package main

import (
	"context"
	"flag"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"path/filepath"
	"strings"

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

// Flags struct holds the command line flags values.
type Flags struct {
	LogFile   string
	LogType   string
	LogLevel  string
	LogSource bool
	Addr      string
	CertFile  string
	KeyFile   string
	TmplDir   string
}

// parseFlags parses the command line flags and returns them in a Flags struct.
func parseFlags() (*Flags, error) {
	flags := &Flags{}

	flag.StringVar(&flags.LogFile, "logfile", "", "Path to log file.")
	flag.StringVar(&flags.LogType, "logtype", "json", "Log type. Valid types are: "+strings.Join(weblog.Types, ","))
	flag.StringVar(&flags.LogLevel, "loglevel", "INFO", "Logging level. Valid levels are: "+weblog.Levels())
	flag.BoolVar(&flags.LogSource, "logsource", false, "Add source code position to log statement.")
	flag.StringVar(&flags.Addr, "addr", ":8443", "Address for server.")
	flag.StringVar(&flags.CertFile, "cert", "cert/cert.pem", "Path to cert file.")
	flag.StringVar(&flags.KeyFile, "key", "cert/key.pem", "Path to key file.")
	flag.StringVar(&flags.TmplDir, "tmpldir", "", "Path to template directory.")

	flag.Parse()

	if flag.NArg() > 0 {
		return nil, fmt.Errorf("unexpected arguments")
	}

	return flags, nil
}

func main() {
	flags, err := parseFlags()
	if err != nil {
		fmt.Fprintln(os.Stderr, "Error:", err)
		flag.Usage()
		os.Exit(ExitUsage)
	}

	// Initialize logging.
	err = weblog.Init(
		weblog.WithFilename(flags.LogFile),
		weblog.WithLogType(flags.LogType),
		weblog.WithLevel(flags.LogLevel),
		weblog.WithSource(flags.LogSource),
	)
	if err != nil {
		fmt.Fprintln(os.Stderr, "Error initializing logger:", err)
		os.Exit(ExitLog)
	}

	// Initialize config
	cfg, err := weblogin.GetConfigFromFile("config.json")
	if err != nil {
		fmt.Fprintln(os.Stderr, "failed to initialize config:", err)
		os.Exit(ExitConfig)
	}

	// Initialize templates
	tmpl, err := webutil.InitTemplates(cfg.ParseGlobPattern)
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
		webapp.WithAppName("Web Login"), webapp.WithTemplate(tmpl),
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
	mux.Handle("/",
		http.RedirectHandler("/user", http.StatusMovedPermanently))
	mux.HandleFunc("/w3.css", webutil.ServeFileHandler(cssFile))
	mux.HandleFunc("/favicon.ico", webutil.ServeFileHandler(icoFile))
	mux.HandleFunc("/user", app.UserHandler)
	mux.HandleFunc("/login", app.LoginHandler)
	mux.HandleFunc("/logout", app.LogoutHandler)
	mux.HandleFunc("/register", app.RegisterHandler)
	mux.HandleFunc("/users", app.UsersHandler)
	mux.HandleFunc("/forgot", app.ForgotHandler)
	mux.HandleFunc("/reset", app.ResetHandler)

	// Create the web server.
	srv, err := webserver.New(
		webserver.WithAddr(flags.Addr),
		webserver.WithHandler(webhandler.AddRequestID(webhandler.AddRequestLogger(webhandler.LogRequest(mux)))),
		webserver.WithTLS(flags.CertFile, flags.KeyFile),
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