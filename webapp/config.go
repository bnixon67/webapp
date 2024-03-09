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

// AppConfig holds the web app settings.
type AppConfig struct {
	Name        string // Name of the web application. (required)
	AssetsDir   string // AssetsDir is directory for web assets.
	TmplPattern string // TmplPattern identifies template files.
}

// Config represents the overall application configuration.
type Config struct {
	App    AppConfig        // App configuration.
	Server webserver.Config // Server configuration.
	Log    weblog.Config    // Log configuration.
}

var (
	ErrConfigOpen   = errors.New("failed to open config file")
	ErrConfigDecode = errors.New("failed to decode config file")
	ErrConfigClose  = errors.New("failed to close config file")
)

// ConfigFromJSONFile returns a Config with settings from a JSON file.
func ConfigFromJSONFile(filename string) (Config, error) {
	var config Config

	file, err := os.Open(filename)
	if err != nil {
		return Config{}, fmt.Errorf("%w: %v", ErrConfigOpen, err)
	}
	defer file.Close()

	if err := json.NewDecoder(file).Decode(&config); err != nil {
		return Config{}, fmt.Errorf("%w: %v", ErrConfigDecode, err)
	}

	return config, nil
}

// appendIfEmpty appends message to messages if value is empty, and
// returns the updated slice.
func appendIfEmpty(messages []string, value, message string) []string {
	if value == "" {
		messages = append(messages, message)
	}

	return messages
}

// Valid checks if all required Config fields are populated.
// Returns a boolean indicating validity and a slice of missing field messages.
func (c *Config) Valid() (bool, []string) {
	var missing []string

	// Append message for each missing field.
	missing = appendIfEmpty(missing, c.App.Name, "missing App.Name")

	return len(missing) == 0, missing
}
