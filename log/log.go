package log

import (
	"fmt"
	"io"
	"os"
	"sort"
	"strings"

	"log/slog"
)

const (
	// logFileFlag defines the file opening flags for the log file.
	logFileFlag = os.O_APPEND | os.O_CREATE | os.O_WRONLY

	// ownerReadWrite is read and write by the owner only.
	ownerReadWrite = 0o600

	// logFilePerm holds the permission setting for the log file.
	logFilePerm = ownerReadWrite
)

// Init initializes the logging system.
func Init(fileName, logType string, level slog.Level, addSource bool) error {
	// Default to Stderr if no log file is provided.
	var w io.Writer = os.Stderr

	// If log file name is provided, attempt to open file for logging.
	if fileName != "" {
		file, err := os.OpenFile(fileName, logFileFlag, logFilePerm)
		if err != nil {
			return fmt.Errorf("error opening log file: %w", err)
		}
		// Don't defer file.Close() since file must remain open.
		w = file
	}

	// Set up logging options.
	opts := &slog.HandlerOptions{
		AddSource: addSource,
		Level:     level,
	}

	// Determine the log handler based on the specified log type.
	var handler slog.Handler
	switch logType {
	case "json":
		handler = slog.NewJSONHandler(w, opts)
	default: // Default to text if an unknown log type is specified.
		handler = slog.NewTextHandler(w, opts)
	}

	// Initialize the logger and set it as default.
	logger := slog.New(handler)
	slog.SetDefault(logger)

	slog.Info("log initialized",
		slog.String("fileName", fileName),
		slog.String("logType", logType),
		slog.String("level", level.String()),
		slog.Bool("addSource", addSource),
	)

	return nil
}

// LevelMap maps between string representations and slog.Level values.
var LevelMap = map[string]slog.Level{
	"DEBUG": slog.LevelDebug,
	"INFO":  slog.LevelInfo,
	"WARN":  slog.LevelWarn,
	"ERROR": slog.LevelError,
}

// Types represents the supported log types.
var Types = []string{"json", "text"}

// Level converts a string representation of a log level to slog.Level.
func Level(s string) (slog.Level, error) {
	//assume LogLevelMap keys are uppercase
	s = strings.ToUpper(s)

	level, ok := LevelMap[s]
	if !ok {
		return slog.LevelInfo, fmt.Errorf("invalid level: %s", s)
	}

	return level, nil
}

// Levels returns a comma-separated string of log levels sorted by severity.
func Levels() string {
	keys := make([]string, 0, len(LevelMap))

	// get all keys
	for key := range LevelMap {
		keys = append(keys, key)
	}

	// sort by key value (slog.Level)
	sort.Slice(keys, func(i, j int) bool {
		return LevelMap[keys[i]] < LevelMap[keys[j]]
	})

	return strings.Join(keys, ", ")
}
