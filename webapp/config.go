// Copyright 2023 Bill Nixon. All rights reserved.
// Use of this source code is governed by the license found in the LICENSE file.

package webapp

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"

	"github.com/bnixon67/webapp/weblog"
	"github.com/bnixon67/webapp/webserver"
)

// ConfigApp holds the web app settings.
type ConfigApp struct {
	Name        string // Name of the application.
	AssetsDir   string // AssetsDir is directory for web asets.
	TmplPattern string // TmplPattern identifies template files.
}

/*
// ConfigLog holds the log settings.
type ConfigLog struct {
	Filename  string // Filename of log file.
	Type      string // Type of log, e.g., json or text.
	Level     string // Level of log, e.g., DEBUG, INFO, WARN, ERROR.
	AddSource bool   // WithSource add source info to log.
}
*/

// Config represents the overall application configuration.
type Config struct {
	App    ConfigApp        // App configuration.
	Server webserver.Config // Server configuration.
	Log    *weblog.Config   // Log configuration.
}

var (
	ErrConfigOpen   = errors.New("failed to open config file")
	ErrConfigDecode = errors.New("failed to decode config file")
)

// ConfigFromJSONFile returns a Config with settings from a JSON file.
func ConfigFromJSONFile(filename string) (Config, error) {
	var config Config

	file, err := os.Open(filename)
	if err != nil {
		return config, fmt.Errorf("%w: %v", ErrConfigOpen, err)
	}
	defer file.Close()

	// decode json from config
	if err := json.NewDecoder(file).Decode(&config); err != nil {
		return config, fmt.Errorf("%w: %v", ErrConfigDecode, err)
	}

	return config, nil
}

// appendIfEmpty appends a message to a slice if a string is empty.
func appendIfEmpty(messages []string, value, message string) []string {
	if value == "" {
		messages = append(messages, message)
	}

	return messages
}

// IsValid checks if all required Config fields are populated.
// Returns a boolean indicating validity and a slice of missing field messages.
func (c *Config) IsValid() (bool, []string) {
	var missing []string

	// Append message for each missing field.
	missing = appendIfEmpty(missing, c.App.Name, "missing Name")

	return len(missing) == 0, missing
}
