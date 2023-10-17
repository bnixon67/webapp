package main

import (
	"flag"
	"fmt"
	"os"
	"strings"

	"github.com/bnixon67/webapp/weblog"
)

const (
	ExitUsage = iota + 1 // ExitUsage indicates a usage error.
	ExitLog              // ExitLog indicates a log error.
)

type Flags struct {
	LogFile   string
	LogType   string
	LogLevel  string
	LogSource bool
}

func parseFlags() (*Flags, error) {
	flags := &Flags{}

	flag.StringVar(&flags.LogFile, "logfile", "", "Path to log file.")
	flag.StringVar(&flags.LogType, "logtype", "text", "Log type. Valid types are: "+strings.Join(weblog.Types, ","))
	flag.StringVar(&flags.LogLevel, "loglevel", "INFO", "Logging level. Valid levels are: "+weblog.Levels())
	flag.BoolVar(&flags.LogSource, "logsource", false, "Add source code position to log statement.")

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

	// Initialize and validate logging.
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
}
