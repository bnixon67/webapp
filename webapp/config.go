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
	ErrConfigRead  = errors.New("failed to read config file")
	ErrConfigParse = errors.New("failed to parse config file")
)

// LoadConfigFromJSON loads app config from a specified JSON file path.
// It returns a populated Config or error if reading or parsing file fails.
func LoadConfigFromJSON(filepath string) (*Config, error) {
	data, err := os.ReadFile(filepath)
	if err != nil {
		return nil, fmt.Errorf("%w: %s", ErrConfigRead, err)
	}

	var config Config
	if err := json.Unmarshal(data, &config); err != nil {
		return nil, fmt.Errorf("%w: %s", ErrConfigParse, err)
	}

	return &config, nil
}

// MissingFields identifies which required fields are absent in Config.
// It returns a slice of missing fields. If an error occurs during the check,
// an empty slice and the error are returned.
func (c *Config) MissingFields() ([]string, error) {
	return required.MissingFields(c)
}
