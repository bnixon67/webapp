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
	ErrOpenLogFile     = errors.New("open log file failed")

	levelMap = map[string]slog.Level{
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

// Init validates config and initializes the default slog logger.
func Init(config Config) error {
	if err := validateType(config.Type); err != nil {
		return err
	}

	if err := validateLevel(config.Level); err != nil {
		return err
	}

	writer, err := writer(config.Filename)
	if err != nil {
		return err
	}

	level, err := ParseLevel(config.Level)
	if err != nil {
		return err
	}

	initLogger(writer, config.Type, level, config.AddSource)

	slog.Debug("initialized logger",
		slog.Group("config",
			slog.String("Filename", filename(writer)),
			slog.String("LogType", config.Type),
			slog.String("Level", level.String()),
			slog.Bool("AddSource", config.AddSource),
		),
	)

	return nil
}

// initLogger initializes the default slog logger.
func initLogger(w io.Writer, logType string, level slog.Level, addSource bool) {
	handlerOptions := &slog.HandlerOptions{
		AddSource: addSource,
		Level:     level,
	}

	var handler slog.Handler
	if logType == "json" {
		handler = slog.NewJSONHandler(w, handlerOptions)
	} else { // default to text
		handler = slog.NewTextHandler(w, handlerOptions)
	}

	slog.SetDefault(slog.New(handler))
}

// ParseLevel converts a string to its corresponding slog.Level.
// If s is empty, the slog.Level zero value is returned.
// If s is invalid, an error is returned.
func ParseLevel(s string) (slog.Level, error) {
	var defaultLevel slog.Level

	if s == "" {
		s = defaultLevel.String()
	}

	level, ok := levelMap[strings.ToUpper(s)]
	if !ok {
		return defaultLevel, fmt.Errorf("%w: %q, valid levels are %q", ErrInvalidLogLevel, s, Levels())
	}

	return level, nil
}

// Levels generates a sorted string slice of valid log levels.
// The levels are sorted by their severity as defined in slog.Level.
func Levels() []string {
	// Pre-allocate the slice with the exact size needed.
	levels := make([]string, 0, len(levelMap))
	for level := range levelMap {
		levels = append(levels, level)
	}

	// Sort the slice based on the severity defined in LevelMap.
	sort.Slice(levels, func(i, j int) bool {
		return levelMap[levels[i]] < levelMap[levels[j]]
	})

	return levels
}

// validateType checks if logType is valid. An empty type is valid.
func validateType(logType string) error {
	if logType == "" {
		return nil
	}

	validTypes := []string{"json", "text"}
	for _, v := range validTypes {
		if logType == v {
			return nil
		}
	}

	return fmt.Errorf("%w: %q, valid types are %q", ErrInvalidLogType, logType, validTypes)
}

// validateLevel checks if level is valid.
func validateLevel(level string) error {
	_, err := ParseLevel(level)
	return err
}

// writer opens and returns a file for the provided name.
// If name is empty, os.Stderr is used.
func writer(name string) (io.Writer, error) {
	if name == "" {
		return os.Stderr, nil
	}

	// Append to file, create if it doesn't exist, open only for writing.
	const flag = os.O_APPEND | os.O_CREATE | os.O_WRONLY

	// Sets permission to read/writeable only by owner.
	const perm = 0o600

	file, err := os.OpenFile(name, flag, perm)
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
