// Copyright 2024 Bill Nixon. All rights reserved.
// Use of this source code is governed by the license found in the LICENSE file.

package webauth

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"

	"github.com/bnixon67/required"
	"github.com/bnixon67/webapp/webapp"
	"github.com/bnixon67/webapp/webutil"
)

// ConfigAuth holds settings specific to the auth app.
type ConfigAuth struct {
	BaseURL      string `required:"true"` // Base URL of the application.
	LoginExpires string `required:"true"` // Duration string for expiry.
}

// ConfigSQL hold SQL database connection settings.
type ConfigSQL struct {
	DriverName     string `required:"true"` // Database driver name.
	DataSourceName string `required:"true"` // Database connection string.
}

// Config represents the overall application configuration.
type Config struct {
	webapp.Config                    // Inherit webapp.Config
	Auth          ConfigAuth         // Auth app configuration.
	SQL           ConfigSQL          // SQL Database configuration.
	SMTP          webutil.SMTPConfig // SMTP server configuration.
}

var (
	ErrConfigRead  = errors.New("failed to read config file")
	ErrConfigParse = errors.New("failed to parse config file")
)

// LoadConfigFromJSON loads configuration settings from a JSON file.
func LoadConfigFromJSON(filepath string) (*Config, error) {
	data, err := os.ReadFile(filepath)
	if err != nil {
		return nil, fmt.Errorf("%w: %v", ErrConfigRead, err)
	}

	var config Config
	if err := json.Unmarshal(data, &config); err != nil {
		return nil, fmt.Errorf("%w: %v", ErrConfigParse, err)
	}

	return &config, nil
}

// MissingFields identifies which required fields are absent in Config.
// It returns a slice of missing fields. If an error occurs during the check,
// an empty slice and the error are returned.
func (c *Config) MissingFields() ([]string, error) {
	return required.MissingFields(c)
}

// RedactedConfig provides a redacted copy of Config for secure logging.
type RedactedConfig Config

// redact creates a redacted copy of Config to hide sensitive data.
func (c *Config) redact() RedactedConfig {
	if c == nil {
		return RedactedConfig{}
	}

	r := RedactedConfig(*c)
	r.SQL.DataSourceName = "[REDACTED]"
	return r
}

// MarshalJSON customizes JSON marshalling to redact sensitive Config data.
func (c *Config) MarshalJSON() ([]byte, error) {
	return json.Marshal(c.redact())
}

// String returns string representation of Config with sensitive data redacted.
func (c *Config) String() string {
	if c == nil {
		return "<nil>"
	}

	return fmt.Sprintf("%+v", c.redact())
}
