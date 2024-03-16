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

// AppConfig holds settings related to the web application itself.
type AppConfig struct {
	Name        string // Required name of the web application.
	AssetsDir   string // Directory for static web assets.
	TmplPattern string // Glob pattern for template files.
}

// Config consolidates configs, including app, server, and log settings.
type Config struct {
	App    AppConfig        // Web application-specific configuration.
	Server webserver.Config // HTTP server configuration.
	Log    weblog.Config    // Logging configuration.
}

// Predefined errors for common configuration issues.
var (
	ErrConfigOpen      = errors.New("failed to open config file")
	ErrConfigUnmarshal = errors.New("failed to parse config file")
)

// LoadConfigFromJSON loads app config from a specified JSON file path.
// It returns a populated Config or error if reading or parsing file fails.
func LoadConfigFromJSON(filepath string) (Config, error) {
	data, err := os.ReadFile(filepath)
	if err != nil {
		return Config{}, fmt.Errorf("%w: %s", ErrConfigOpen, err)
	}

	var config Config
	if err := json.Unmarshal(data, &config); err != nil {
		return Config{}, fmt.Errorf("%w: %s", ErrConfigUnmarshal, err)
	}

	return config, nil
}

// appendIfMissing adds and returns updated messages slice if message is empty.
// It can be used to accumulate error messages for missing configuration values.
func appendIfMissing(messages []string, value, message string) []string {
	if value == "" {
		messages = append(messages, message)
	}

	return messages
}

// Validate checks the Config struct for any missing required fields.
// Returns true if required fields are present, otherwise returns false
// with slice of missing field messages.
func (c *Config) Validate() (bool, []string) {
	var missing []string

	// Append message for each missing field.
	missing = appendIfMissing(missing, c.App.Name, "App.Name is required")

	return len(missing) == 0, missing
}
