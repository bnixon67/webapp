package main

import (
	"flag"
	"fmt"
	"os"
	"slices"
	"strings"

	"github.com/bnixon67/webapp/log"
)

const (
	ExitUsage = iota + 1 // ExitUsage indicates a usage error.
	ExitLog              // ExitLog indicates a log error.
)

// initializeLog configures and initializes the logging mechanism based on the provided command-line flags.
func initializeLog(logFileName, logType, logLevel string, logAddSource bool) error {
	// Validate the log type against the supported types.
	if !slices.Contains(log.Types, logType) {
		return fmt.Errorf("invalid log type: %v. valid log types: %s", logType, strings.Join(log.Types, ", "))
	}

	// Convert and validate the log level.
	level, err := log.Level(logLevel)
	if err != nil {
		return fmt.Errorf("error: %v. valid log levels: %s", err, log.Levels())
	}

	// Initialize the logger with the specified parameters and handle any initialization error.
	if err := log.Init(logFileName, logType, level, logAddSource); err != nil {
		return fmt.Errorf("failed to initialize logger: %v", err)
	}

	return nil
}

func main() {
	// Define command-line flags with default values and descriptions.
	logFileNameFlag := flag.String("logfile", "", "Path to log file.")
	logTypeFlag := flag.String("logtype", "text", "Log type. Valid types are: "+strings.Join(log.Types, ", "))
	logLevelFlag := flag.String("loglevel", "INFO", "Logging level. Valid levels are: "+log.Levels())
	logAddSourceFlag := flag.Bool("logsource", false, "Add source code position to log statement.")

	// Parse the command-line flags.
	flag.Parse()

	// Check for unexpected command-line arguments.
	if flag.NArg() > 0 {
		fmt.Fprintln(os.Stderr, "Unexpected arguments.")
		flag.Usage()
		os.Exit(ExitUsage)
	}

	// Initialize and validate logging.
	if err := initializeLog(*logFileNameFlag, *logTypeFlag, *logLevelFlag, *logAddSourceFlag); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(ExitLog)
	}
}
