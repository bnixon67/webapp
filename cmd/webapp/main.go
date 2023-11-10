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

	"github.com/bnixon67/webapp/webapp"
	"github.com/bnixon67/webapp/webhandler"
	"github.com/bnixon67/webapp/weblog"
	"github.com/bnixon67/webapp/webserver"
	"github.com/bnixon67/webapp/webutil"
)

const (
	ExitUsage    = iota + 1 // ExitUsage indicates a usage error.
	ExitLog                 // ExitLog indicates a log error.
	ExitHandler             // ExitHandler indicates a handler error.
	ExitServer              // ExitServer indicates a server error.
	ExitTemplate            // ExitTemplate indicates a template error.
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
	flag.StringVar(&flags.LogType, "logtype", "text", "Log type. Valid types are: "+strings.Join(weblog.Types, ","))
	flag.StringVar(&flags.LogLevel, "loglevel", "INFO", "Logging level. Valid levels are: "+weblog.Levels())
	flag.BoolVar(&flags.LogSource, "logsource", false, "Add source code position to log statement.")
	flag.StringVar(&flags.Addr, "addr", ":8080", "Address for server.")
	flag.StringVar(&flags.CertFile, "cert", "", "Path to cert file.")
	flag.StringVar(&flags.KeyFile, "key", "", "Path to key file.")
	flag.StringVar(&flags.TmplDir, "tmpldir", "../../assets/tmpl", "Path to remplate directory.")

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

	// Initialize templates
	tmpl, err := webutil.InitTemplates(filepath.Join(flags.TmplDir, "*.html"))
	if err != nil {
		fmt.Fprintln(os.Stderr, "Error initializing templates:", err)
		os.Exit(ExitTemplate)
	}

	// Create the web app.
	app, err := webapp.New(webapp.WithAppName("Web Server"), webapp.WithTemplate(tmpl))
	if err != nil {
		fmt.Fprintln(os.Stderr, "Error creating new handler:", err)
		os.Exit(ExitHandler)
	}

	// Create a new ServeMux to handle HTTP requests.
	mux := http.NewServeMux()
	mux.HandleFunc("/", app.RootHandler)
	mux.HandleFunc("/hello", app.HelloTextHandler)
	mux.HandleFunc("/hellohtml", app.HelloHTMLHandler)
	mux.HandleFunc("/build", app.BuildHandler)
	mux.HandleFunc("/headers", app.HeadersHandler)
	mux.HandleFunc("/remote", webhandler.RemoteHandler)
	mux.HandleFunc("/request", webhandler.RequestHandler)

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