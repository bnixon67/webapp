// Copyright 2023 Bill Nixon. All rights reserved.
// Use of this source code is governed by the license found in the LICENSE file.

package main

import (
	"context"
	"fmt"
	"log/slog"
	"net/http"
	"os"

	_ "github.com/go-sql-driver/mysql"

	"github.com/bnixon67/webapp/webapp"
	"github.com/bnixon67/webapp/webauth"
	"github.com/bnixon67/webapp/webutil"
)

const (
	ExitUsage = iota + 1
	ExitConfig
	ExitInit
	ExitApp
	ExitServer
	ExitEmail
)

// SendStartingEmail sends an email indicating the server is starting.
func SendStartingEmail(to string, cfg webutil.SMTPConfig) error {
	hostName, err := os.Hostname()
	if err != nil {
		return err
	}
	subj := "starting webauth"
	body := "starting webauth on" + hostName

	err = cfg.SendMessage(cfg.User, []string{to}, subj, body)
	if err != nil {
		return err
	}

	return nil
}

func main() {
	// Check for command line argument with config file.
	if len(os.Args) != 2 {
		fmt.Fprintf(os.Stderr, "usage: %s [config file]\n", os.Args[0])
		os.Exit(ExitUsage)
	}

	// Read config.
	cfg, err := webauth.ConfigFromJSONFile(os.Args[1])
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(ExitConfig)
	}

	// Initialize logging, templates, database.
	tmpl, db, err := Init(cfg)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(ExitInit)
	}

	// Create the app.
	app, err := webauth.NewApp(
		webapp.WithName(cfg.App.Name), webapp.WithTemplate(tmpl),
		webauth.WithConfig(cfg), webauth.WithDB(db),
	)
	if err != nil {
		fmt.Fprintln(os.Stderr, "failed to create app:", err)
		os.Exit(ExitApp)
	}

	// Create new ServeMux for HTTP requests and add routes and middleware.
	mux := http.NewServeMux()
	AddRoutes(mux, app)
	handler := AddMiddleware(mux)

	// Create the web server.
	srv, err := cfg.Server.Create(handler)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(ExitServer)
	}

	// Send starting email to confirm SMTP configuration is valid.
	err = SendStartingEmail("bnixon67@gmail.com", cfg.SMTP)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(ExitEmail)
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
