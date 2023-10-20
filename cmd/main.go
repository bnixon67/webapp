package main

import (
	"context"
	"flag"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"strings"

	"github.com/bnixon67/webapp/webhandler"
	"github.com/bnixon67/webapp/weblog"
	"github.com/bnixon67/webapp/webserver"
)

const (
	ExitUsage   = iota + 1 // ExitUsage indicates a usage error.
	ExitLog                // ExitLog indicates a log error.
	ExitHandler            // ExitHandler indicates a handler error.
	ExitServer             // ExitServer indicates a server error.
)

// Flags struct holds the command line flags values.
type Flags struct {
	LogFile   string
	LogType   string
	LogLevel  string
	LogSource bool
	Addr      string
}

// parseFlags parses the command line flags and returns them in a Flags struct.
func parseFlags() (*Flags, error) {
	flags := &Flags{}

	flag.StringVar(&flags.LogFile, "logfile", "", "Path to log file.")
	flag.StringVar(&flags.LogType, "logtype", "text", "Log type. Valid types are: "+strings.Join(weblog.Types, ","))
	flag.StringVar(&flags.LogLevel, "loglevel", "INFO", "Logging level. Valid levels are: "+weblog.Levels())
	flag.BoolVar(&flags.LogSource, "logsource", false, "Add source code position to log statement.")
	flag.StringVar(&flags.Addr, "addr", ":8080", "Address for server.")

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
		weblog.WithFileName(flags.LogFile),
		weblog.WithLogType(flags.LogType),
		weblog.WithLevel(flags.LogLevel),
		weblog.WithSource(flags.LogSource),
	)
	if err != nil {
		fmt.Fprintln(os.Stderr, "Error initializing logger:", err)
		os.Exit(ExitLog)
	}

	// Create the web handler.
	h, err := webhandler.New(webhandler.WithAppName("Web Server"))
	if err != nil {
		fmt.Fprintln(os.Stderr, "Error creating new handler:", err)
		os.Exit(ExitHandler)
	}

	// Create a new ServeMux to handle HTTP requests.
	mux := http.NewServeMux()
	mux.HandleFunc("/hello", h.HelloHandler)
	mux.HandleFunc("/build", h.BuildHandler)

	// Create the web server.
	srv, err := webserver.New(
		webserver.WithAddr(flags.Addr),
		webserver.WithHandler(
			h.AddRequestID(h.AddLogger(h.LogRequest(mux)))),
	)
	if err != nil {
		fmt.Fprintln(os.Stderr, "Error creating server:", err)
		os.Exit(ExitServer)
	}

	// Create a new context.
	ctx := context.Background()

	// Run the web server.
	err = webserver.Run(ctx, srv)
	if err != nil {
		slog.Error("error running server", "err", err)
		fmt.Fprintln(os.Stderr, "Error running server:", err)
		os.Exit(ExitServer)
	}
}
