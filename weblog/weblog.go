// Copyright 2024 Bill Nixon. All rights reserved.
// Use of this source code is governed by the license found in the LICENSE file.

// Package weblog provides a configurable logging system based on slog
// by providing convenient configuration and initialization options.
package weblog

import (
	"errors"
	"fmt"
	"io"
	"log/slog"
	"os"
	"sort"
	"strings"
)

var (
	ErrInvalidLogType  = errors.New("invalid log type")
	ErrInvalidLogLevel = errors.New("invalid log level")
	ErrOpenLogFile     = errors.New("failed to open log file")

	logLevelMap = map[string]slog.Level{
		"DEBUG": slog.LevelDebug,
		"INFO":  slog.LevelInfo,
		"WARN":  slog.LevelWarn,
		"ERROR": slog.LevelError,
	}
)

// Config defines the configuration options for logging.
type Config struct {
	Filename  string // Path to the log file. Uses stderr if empty.
	Type      string // Log format: 'json' or 'text'.
	Level     string // Log level as a string.
	AddSource bool   // If true, includes source code position in logs.
}

func isValidLogType(logType string) bool {
	switch logType {
	case "", "json", "text":
		return true
	default:
		return false
	}
}

// ParseLogLevel converts a log level string to its corresponding slog.Level.
// An empty string returns the default value of slog.Level with no error.
// If the level string is invalid, it returns an error.
func ParseLogLevel(level string) (slog.Level, error) {
	var defaultLevel slog.Level

	if level == "" {
		level = defaultLevel.String()
	}

	parsedLevel, exists := logLevelMap[strings.ToUpper(level)]
	if !exists {
		return defaultLevel, fmt.Errorf("%w: %q, valid levels are %q", ErrInvalidLogLevel, level, Levels())
	}

	return parsedLevel, nil
}

// Init validates config and initializes the default slog logger.
func Init(config Config) error {
	if !isValidLogType(config.Type) {
		return fmt.Errorf("%w: %q", ErrInvalidLogType, config.Type)
	}

	level, err := ParseLogLevel(config.Level)
	if err != nil {
		return err
	}

	logWriter, err := writer(config.Filename)
	if err != nil {
		return err
	}

	setupLogger(logWriter, config.Type, level, config.AddSource)

	slog.Debug("initialized logger",
		slog.Group("config",
			slog.String("Filename", filename(logWriter)),
			slog.String("LogType", config.Type),
			slog.String("Level", level.String()),
			slog.Bool("AddSource", config.AddSource),
		),
	)

	return nil
}

func chooseLogHandler(writer io.Writer, logType string, options *slog.HandlerOptions) slog.Handler {
	if logType == "json" {
		return slog.NewJSONHandler(writer, options)
	}
	return slog.NewTextHandler(writer, options)
}

func setupLogger(writer io.Writer, logType string, level slog.Level, addSource bool) {
	options := &slog.HandlerOptions{
		AddSource: addSource,
		Level:     level,
	}

	handler := chooseLogHandler(writer, logType, options)
	slog.SetDefault(slog.New(handler))
}

// Levels generates a sorted list of valid log levels from the logLevelMap.
func Levels() []string {
	levels := make([]string, 0, len(logLevelMap))
	for level := range logLevelMap {
		levels = append(levels, level)
	}

	// Sort the slice based on the severity defined in LevelMap.
	sort.Slice(levels, func(i, j int) bool {
		return logLevelMap[levels[i]] < logLevelMap[levels[j]]
	})

	return levels
}

// writer opens and returns a file for the provided name.
// If name is empty, os.Stderr is used.
func writer(filename string) (io.Writer, error) {
	if filename == "" {
		return os.Stderr, nil
	}

	// Append to file, create if it doesn't exist, open only for writing.
	const flag = os.O_APPEND | os.O_CREATE | os.O_WRONLY

	// Sets permission to read/writeable only by owner.
	const perm = 0o600

	file, err := os.OpenFile(filename, flag, perm)
	if err != nil {
		return nil, fmt.Errorf("%w: %v", ErrOpenLogFile, err)
	}

	return file, nil
}

// filename returns the filename from an *os.File writer.
// If writer is not an *os.File, then an empty string returned.
func filename(writer io.Writer) string {
	if file, ok := writer.(*os.File); ok {
		return file.Name()
	}
	return ""
}
