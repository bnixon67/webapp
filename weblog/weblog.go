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

const (
	logFileFlag    = os.O_APPEND | os.O_CREATE | os.O_WRONLY
	ownerReadWrite = 0o600
	logFilePerm    = ownerReadWrite
)

var (
	ErrInvalidLogType   = errors.New("invalid log type")
	ErrInvalidLogLevel  = errors.New("invalid log level")
	ErrLogFileOpenError = errors.New("error opening log file")
)

// Types is a slice of supported log types.
var Types = []string{"json", "text"}

// Log contains configuration options to initialize the logging system.
type Log struct {
	Filename    string // The name of the log file.
	HandlerType string // The type of logging, e.g., json or text.
	Level       string // The log level, e.g., DEBUG, INFO, etc.
	AddSource   bool   // Indicates if source code position is included in log.
}

// Option is a function type that modifies the Config.
// This follows the Option pattern from https://commandcenter.blogspot.com/2014/01/self-referential-functions-and-design.html and elsewhere.
type Option func(*Log)

// WithFilename returns an Option that sets the Config's Filename field.
func WithFilename(filename string) Option {
	return func(l *Log) {
		l.Filename = filename
	}
}

// WithLogType returns an Option that sets the Config's LogType field.
func WithLogType(logType string) Option {
	return func(l *Log) {
		l.HandlerType = logType
	}
}

// WithLevel returns an Option that sets the Config's Level field.
func WithLevel(level string) Option {
	return func(l *Log) {
		l.Level = level
	}
}

// WithSource returns an Option that sets the Config's AddSource field.
func WithSource(addSource bool) Option {
	return func(l *Log) {
		l.AddSource = addSource
	}
}

// Init initializes the logging system using the provided options.
func Init(opts ...Option) error {
	l := &Log{
		HandlerType: "text", // Default to text log type
		Level:       "INFO", // Default to INFO log level
	}

	// Apply the provided options to override the defaults if needed.
	for _, opt := range opts {
		opt(l)
	}

	// Validate the log type against the supported types.
	if !slices.Contains(Types, l.HandlerType) {
		return fmt.Errorf("%w: %v, valid log types: %s",
			ErrInvalidLogType,
			l.HandlerType,
			strings.Join(Types, ","))
	}

	// Convert and validate the log level.
	level, err := Level(l.Level)
	if err != nil {
		return fmt.Errorf("%w: %s, valid log levels: %s",
			ErrInvalidLogLevel,
			l.Level,
			Levels())
	}

	writer, err := getWriter(l.Filename)
	if err != nil {
		return err
	}

	initLogger(writer, l, level)

	return nil
}

// getWriter opens and returns a writer based on the provided filename.
// If filename is empty, return os.Stderr.
func getWriter(filename string) (io.Writer, error) {
	if filename == "" {
		return os.Stderr, nil
	}

	file, err := os.OpenFile(filename, logFileFlag, logFilePerm)
	if err != nil {
		return nil, fmt.Errorf("%w: %v", ErrLogFileOpenError, err)
	}

	// Do not close file since it is used outside function to log entries.
	return file, nil
}

// initLogger configures and initializes the logger.
func initLogger(writer io.Writer, config *Log, level slog.Level) {
	// Set up log handler options.
	hOpts := &slog.HandlerOptions{
		AddSource: config.AddSource,
		Level:     level,
	}

	// Select the log handler based on the log type.
	var handler slog.Handler
	switch config.HandlerType {
	case "json":
		handler = slog.NewJSONHandler(writer, hOpts)
	default:
		handler = slog.NewTextHandler(writer, hOpts)
	}

	// Create new logger instance and set as default.
	logger := slog.New(handler)
	slog.SetDefault(logger)

	slog.Info("log initialized",
		slog.Group("log",
			slog.String("Filename", config.Filename),
			slog.String("LogType", config.HandlerType),
			slog.String("Level", level.String()),
			slog.Bool("AddSource", config.AddSource),
		),
	)
}

// LevelMap maps between string representations and slog.Level values.
var LevelMap = map[string]slog.Level{
	"DEBUG": slog.LevelDebug,
	"INFO":  slog.LevelInfo,
	"WARN":  slog.LevelWarn,
	"ERROR": slog.LevelError,
}

// Level converts a string representation of a log level to a slog.Level.
// It returns an error if the string does not represent a valid log level.
func Level(s string) (slog.Level, error) {
	// assume LogLevelMap keys are uppercase
	s = strings.ToUpper(s)

	level, ok := LevelMap[s]
	if !ok {
		return slog.LevelInfo, fmt.Errorf("%w: %s", ErrInvalidLogLevel, s)
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

	return strings.Join(keys, ",")
}
