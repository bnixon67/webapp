// Package weblog provides a logging system for the webapp based on slog.
package weblog

import (
	"errors"
	"fmt"
	"io"
	"os"
	"slices"
	"sort"
	"strings"

	"log/slog"
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

// Config contains configuration options to initialize the logging system.
type Config struct {
	FileName  string // The name of the log file.
	LogType   string // The type of logging, e.g., json or text.
	Level     string // The log level, e.g., DEBUG, INFO, etc.
	AddSource bool   // Indicates if source code position is included in log.
}

// Option is a function type that modifies the Config.
type Option func(*Config)

// WithFileName returns an Option that sets the Config's FileName field.
func WithFileName(fileName string) Option {
	return func(c *Config) {
		c.FileName = fileName
	}
}

// WithLogType returns an Option that sets the Config's LogType field.
func WithLogType(logType string) Option {
	return func(c *Config) {
		c.LogType = logType
	}
}

// WithLevel returns an Option that sets the Config's Level field.
func WithLevel(level string) Option {
	return func(c *Config) {
		c.Level = level
	}
}

// WithSource returns an Option that sets the Config's AddSource field.
func WithSource(addSource bool) Option {
	return func(c *Config) {
		c.AddSource = addSource
	}
}

// Init initializes the logging system using the provided options.
func Init(opts ...Option) error {
	config := &Config{
		LogType: "text", // Default to text log type
		Level:   "INFO", // Default to INFO log level
	}

	// Apply the provided options to override the defaults if needed.
	for _, opt := range opts {
		opt(config)
	}

	// Validate the log type against the supported types.
	if !slices.Contains(Types, config.LogType) {
		return fmt.Errorf("%w: %v, valid log types: %s",
			ErrInvalidLogType,
			config.LogType,
			strings.Join(Types, ","))
	}

	// Convert and validate the log level.
	level, err := Level(config.Level)
	if err != nil {
		return fmt.Errorf("%w: %s, valid log levels: %s",
			ErrInvalidLogLevel,
			config.Level,
			Levels())
	}

	writer, err := getWriter(config.FileName)
	if err != nil {
		return err
	}

	initLogger(writer, config, level)

	return nil
}

// getWriter opens and returns a writer based on the provided fileName.
// If fileName is empty, return os.Stderr.
func getWriter(fileName string) (io.Writer, error) {
	if fileName == "" {
		return os.Stderr, nil
	}

	file, err := os.OpenFile(fileName, logFileFlag, logFilePerm)
	if err != nil {
		return nil, fmt.Errorf("%w: %v", ErrLogFileOpenError, err)
	}

	// Do not close file as it is used outside function to write log entries.
	return file, nil
}

// initLogger configures and initializes the logger.
func initLogger(writer io.Writer, config *Config, level slog.Level) {
	// Set up log handler options.
	hOpts := &slog.HandlerOptions{
		AddSource: config.AddSource,
		Level:     level,
	}

	// Select the log handler based on the log type.
	var handler slog.Handler
	switch config.LogType {
	case "json":
		handler = slog.NewJSONHandler(writer, hOpts)
	default:
		handler = slog.NewTextHandler(writer, hOpts)
	}

	// Create new logger instance and set as default.
	logger := slog.New(handler)
	slog.SetDefault(logger)

	slog.Info("log initialized",
		slog.Group("config",
			slog.String("fileName", config.FileName),
			slog.String("logType", config.LogType),
			slog.String("level", level.String()),
			slog.Bool("addSource", config.AddSource),
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
	//assume LogLevelMap keys are uppercase
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
