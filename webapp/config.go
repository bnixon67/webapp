// Copyright 2024 Bill Nixon. All rights reserved.
// Use of this source code is governed by the license found in the LICENSE file.

package webapp

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"

	"github.com/bnixon67/required"
	"github.com/bnixon67/webapp/weblog"
	"github.com/bnixon67/webapp/webserver"
)

// AppConfig holds settings related to the web application itself.
type AppConfig struct {
	Name        string `required:"true"` // Name of the web application.
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
	ErrConfigOpen  = errors.New("failed to open config file")
	ErrConfigParse = errors.New("failed to parse config file")
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
		return Config{}, fmt.Errorf("%w: %s", ErrConfigParse, err)
	}

	return config, nil
}

// IsValid checks the Config struct for any missing required fields.
// Returns true if required fields are present, otherwise returns false
// with slice of missing field messages.
func (c *Config) IsValid() (bool, []string, error) {
	missingFields, err := required.MissingFields(c)
	if err != nil {
		return false, []string{}, err
	}

	return len(missingFields) == 0, missingFields, nil
}
