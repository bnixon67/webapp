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
	ErrConfigRead      = errors.New("failed to read config file")
	ErrConfigUnmarshal = errors.New("failed to unmarshal config file")
)

// ConfigFromJSONFile loads configuration settings from a JSON file.
func ConfigFromJSONFile(filename string) (Config, error) {
	// Read the entire file.
	data, err := os.ReadFile(filename)
	if err != nil {
		return Config{}, fmt.Errorf("%w: %v", ErrConfigRead, err)
	}

	// Decode JSON data into Config struct.
	var c Config
	if err := json.Unmarshal(data, &c); err != nil {
		return Config{}, fmt.Errorf("%w: %v", ErrConfigUnmarshal, err)
	}

	return c, nil
}

// appendIfEmpty appends message to messages if value is empty, and
// returns the updated slice.
func appendIfEmpty(messages []string, value, message string) []string {
	if value == "" {
		return append(messages, message)
	}

	return messages
}

// IsValid checks if all required Config fields are populated.
// Returns a boolean and a slice of messages indicating the issue(s).
func (c *Config) IsValid() (bool, []string, error) {
	missingFields, err := required.MissingFields(c)
	if err != nil {
		return false, []string{}, err
	}

	return len(missingFields) == 0, missingFields, nil
}

// RedactedConfig provides a redacted copy of Config for secure logging.
type RedactedConfig Config

// redact creates a redacted copy of Config to hide sensitive data.
func (c *Config) redact() RedactedConfig {
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
	return fmt.Sprintf("%+v", c.redact())
}

// TODO: slog.LogValuer for Config
