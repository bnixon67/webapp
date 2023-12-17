// Copyright 2023 Bill Nixon. All rights reserved.
// Use of this source code is governed by the license found in the LICENSE file.

// Package weblog provides a logging system for the webapp based on slog.
package weblog

import (
	"errors"
	"fmt"
	"io"
	"log/slog"
	"os"
	"slices"
	"sort"
	"strings"
)

// Log configures the parameters for the logging system.
type Log struct {
	Filename    string     // Filename specifies the file to write logs.
	Type        string     // Type is the log format, e.g., json or text.
	LevelString string     // LevelString is the log level as a string.
	Level       slog.Level // Level is the log level as a slog.Level type.
	AddSource   bool       // AddSource, if true, adds source code position.
}

// Option is a function type that modifies the Config.
// This follows the Option pattern from https://commandcenter.blogspot.com/2014/01/self-referential-functions-and-design.html and elsewhere.
type Option func(*Log)

// WithFilename creates an Option to set the Log's Filename field.
func WithFilename(filename string) Option {
	return func(l *Log) {
		l.Filename = filename
	}
}

// WithType creates an Option to set the Log's Type field.
// If logType is empty, it preserves the default value.
func WithType(logType string) Option {
	return func(l *Log) {
		if logType != "" {
			l.Type = logType
		}
	}
}

// WithLevel creates an Option to sets the Log's Level field.
// If level is empty, it preserves the default value.
func WithLevel(level string) Option {
	return func(l *Log) {
		if level != "" { // Ignore empty to allow defaults.
			l.LevelString = level
		}
	}
}

// WithSource creates an Option that sets the Log's AddSource field.
func WithSource(addSource bool) Option {
	return func(l *Log) {
		l.AddSource = addSource
	}
}

var (
	ErrInvalidLogType   = errors.New("invalid log type")
	ErrInvalidLevel     = errors.New("invalid log level")
	ErrLogFileOpenError = errors.New("error opening log file")
)

// Init initializes the logging system with the given options.
func Init(opts ...Option) error {
	// Set default values.
	l := &Log{
		Type:        "text",
		LevelString: "INFO",
	}

	// Apply each Option to override defaults.
	for _, opt := range opts {
		opt(l)
	}

	// Validate the log type.
	logTypes := []string{"json", "text"}
	if !slices.Contains(logTypes, l.Type) {
		return fmt.Errorf("%w: %v, valid log types: %s",
			ErrInvalidLogType,
			l.Type,
			strings.Join(logTypes, ","))
	}

	// Convert the string level to a log level type and validate.
	var err error
	l.Level, err = ParseLevel(l.LevelString)
	if err != nil {
		return fmt.Errorf("%w: %s, valid log levels: %s",
			ErrInvalidLevel,
			l.LevelString,
			Levels())
	}

	// Setup the log output writer.
	writer, err := getWriter(l.Filename)
	if err != nil {
		return err
	}

	// Initialize the logger with the configured settings.
	initLogger(writer, l)

	return nil
}

const (
	// logFileFlag defines the flags for opening the log file.
	// Append to file, create if it doesn't exist, open for writing only.
	logFileFlag = os.O_APPEND | os.O_CREATE | os.O_WRONLY

	// ownerReadWrite sets permission to read/writeable only by owner.
	ownerReadWrite = 0o600
)

// getWriter opens a file for the provided filename.
// If the filename is empty, it defaults to os.Stderr.
func getWriter(filename string) (io.Writer, error) {
	if filename == "" {
		return os.Stderr, nil
	}

	file, err := os.OpenFile(filename, logFileFlag, ownerReadWrite)
	if err != nil {
		return nil, fmt.Errorf("%w: %v", ErrLogFileOpenError, err)
	}

	// Return the opened file for logging without closing it here.
	return file, nil
}

// initLogger configures and initializes the logger.
func initLogger(writer io.Writer, config *Log) {
	// Prepare options for the log handler.
	hOpts := &slog.HandlerOptions{
		AddSource: config.AddSource,
		Level:     config.Level,
	}

	// Determine the appropriate log handler based on the log type.
	var handler slog.Handler
	switch config.Type {
	case "json":
		handler = slog.NewJSONHandler(writer, hOpts)
	default:
		handler = slog.NewTextHandler(writer, hOpts)
	}

	// Create new logger and set as default.
	logger := slog.New(handler)
	slog.SetDefault(logger)

	slog.Debug("initialized logger",
		slog.Group("config",
			slog.String("Filename", config.Filename),
			slog.String("LogType", config.Type),
			slog.String("Level", config.Level.String()),
			slog.Bool("AddSource", config.AddSource),
		),
	)
}

// LevelMap maps string representations of log levels with slog.Level.
// This is used to convert string inputs, like 'DEBUG' into slog.Level values.
var LevelMap = map[string]slog.Level{
	"DEBUG": slog.LevelDebug,
	"INFO":  slog.LevelInfo,
	"WARN":  slog.LevelWarn,
	"ERROR": slog.LevelError,
}

// ParseLevel converts s to its corresponding slog.Level.
// If s is not a valid log level, it returns an error and slog.LevelInfo.
func ParseLevel(s string) (slog.Level, error) {
	// assume LogLevelMap keys are uppercase
	s = strings.ToUpper(s)

	level, ok := LevelMap[s]
	if !ok {
		return slog.LevelInfo, fmt.Errorf("%w: %s", ErrInvalidLevel, s)
	}

	return level, nil
}

// Levels generates a sorted, comma-separated string of available log levels.
// The levels are sorted by their severity as defined in slog.Level.
func Levels() string {
	keys := make([]string, 0, len(LevelMap))

	// Extract keys (log level strings) from LevelMap.
	for key := range LevelMap {
		keys = append(keys, key)
	}

	// Sort the keys based on their corresponding slog.Level values.
	sort.Slice(keys, func(i, j int) bool {
		return LevelMap[keys[i]] < LevelMap[keys[j]]
	})

	// Join the sorted keys into a comma-separated string.
	return strings.Join(keys, ",")
}
