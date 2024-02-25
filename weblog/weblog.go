// Copyright 2024 Bill Nixon. All rights reserved.
// Use of this source code is governed by the license found in the LICENSE file.

// Package weblog provides a configurable logging system based on slog.
//
// This package simplifies integrating a robust logging solution into
// web applications or services by providing convenient configuration and
// initialization functions.
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
	ErrOpenLogFile     = errors.New("error opening log file")

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

// Init validates and initializes the logger using the Config.
func (c *Config) Init() error {
	if err := validateLogType(c.Type); err != nil {
		return err
	}

	if err := validateLogLevel(c.Level); err != nil {
		return err
	}

	writer, err := getWriter(c.Filename)
	if err != nil {
		return err
	}

	level, err := LevelFromString(c.Level)
	if err != nil {
		return err
	}

	initLogger(writer, c.Type, level, c.AddSource)

	slog.Debug("initialized logger",
		slog.Group("config",
			slog.String("Filename", fileNameFromWriter(writer)),
			slog.String("LogType", c.Type),
			slog.String("Level", level.String()),
			slog.Bool("AddSource", c.AddSource),
		),
	)

	return nil
}

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

// LevelFromString converts a string to its corresponding slog.Level.
// It defaults to the slog.Level zero value if s is empty.
// An error is returned if s is not a valid level.
func LevelFromString(s string) (slog.Level, error) {
	var defaultLevel slog.Level

	if s == "" {
		s = defaultLevel.String()
	}

	level, ok := levelMap[strings.ToUpper(s)]
	if !ok {
		return defaultLevel, fmt.Errorf("%w: %s", ErrInvalidLogLevel, s)
	}

	return level, nil
}

// Levels generates a sorted, comma-separated string of available log levels.
// The levels are sorted by their severity as defined in slog.Level.
func Levels() string {
	// Pre-allocate the slice with the exact size needed.
	levels := make([]string, 0, len(levelMap))
	for level := range levelMap {
		levels = append(levels, level)
	}

	// Sort the levels slice based on the severity defined in LevelMap.
	sort.Slice(levels, func(i, j int) bool {
		return levelMap[levels[i]] < levelMap[levels[j]]
	})

	return strings.Join(levels, ",")
}

// validateLogType checks if the provided log type is valid.
func validateLogType(t string) error {
	validTypes := []string{"", "json", "text"}
	for _, v := range validTypes {
		if t == v {
			return nil
		}
	}
	return fmt.Errorf("%w: %v, valid log types: %v",
		ErrInvalidLogType, t, validTypes)
}

// validateLogLevel checks if the provided log level is valid.
func validateLogLevel(l string) error {
	_, err := LevelFromString(l)
	return err
}

// getWriter opens a file for the provided filename.
// If the filename is empty, it defaults to os.Stderr.
func getWriter(filename string) (io.Writer, error) {
	if filename == "" {
		return os.Stderr, nil
	}

	// Append to file, create if it doesn't exist, open for writing only.
	const logFileFlag = os.O_APPEND | os.O_CREATE | os.O_WRONLY

	// ownerReadWrite sets permission to read/writeable only by owner.
	const ownerReadWrite = 0o600

	file, err := os.OpenFile(filename, logFileFlag, ownerReadWrite)
	if err != nil {
		return nil, fmt.Errorf("%w: %v", ErrOpenLogFile, err)
	}

	// Return the opened file for logging without closing it here.
	return file, nil
}

// fileNameFromWriter returns the filename from an *os.File writer.
// If writer is not an *os.File, then empty string returned.
func fileNameFromWriter(writer io.Writer) string {
	if file, ok := writer.(*os.File); ok {
		return file.Name()
	}
	return ""
}
