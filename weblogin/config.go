// Copyright 2023 Bill Nixon. All rights reserved.
// Use of this source code is governed by the license found in the LICENSE file.

package weblogin

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"

	"github.com/bnixon67/webapp/webapp"
)

var (
	ErrConfigOpen   = errors.New("failed to open config file")
	ErrConfigDecode = errors.New("failed to decode config file")
)

// ConfigSQL hold SQL database connection settings.
type ConfigSQL struct {
	DriverName     string // Database driver name.
	DataSourceName string // Database connection string.
}

// ConfigSMTP holds SMTP server settings for email functionality.
type ConfigSMTP struct {
	Host     string // Host address.
	Port     string // Port number.
	User     string // Server username.
	Password string // Server password.
}

// Config represents the overall application configuration.
type Config struct {
	webapp.Config // Inherit webapp.Config

	BaseURL          string // Base URL of the application.
	ParseGlobPattern string // Glob pattern for parsing template files.
	LoginExpires     string // Duration string for login expiry.

	SQL  ConfigSQL  // SQL Database configuration.
	SMTP ConfigSMTP // SMTP server configuration.
}

// ConfigFromJSONFile loads configuration settings from a JSON file.
func ConfigFromJSONFile(filename string) (Config, error) {
	file, err := os.Open(filename)
	if err != nil {
		return Config{}, fmt.Errorf("%w: %v", ErrConfigOpen, err)
	}
	defer file.Close()

	var c Config
	if err := json.NewDecoder(file).Decode(&c); err != nil {
		return Config{}, fmt.Errorf("%w: %v", ErrConfigDecode, err)
	}

	return c, nil
}

// appendIfEmpty appends message to a slice if value is empty.
func appendIfEmpty(messages []string, value, message string) []string {
	if value == "" {
		messages = append(messages, message)
	}

	return messages
}

// IsValid checks if all required Config fields are populated.
// Returns a boolean and a slice of messages indicating the issue(s).
func (c *Config) IsValid() (bool, []string) {
	isValid, missing := c.Config.IsValid()

	// Append errors for each missing mandatory field.
	missing = appendIfEmpty(missing, c.BaseURL, "missing BaseURL")
	missing = appendIfEmpty(missing, c.ParseGlobPattern, "missing ParseGlobPattern")
	missing = appendIfEmpty(missing, c.LoginExpires, "missing LoginExpires")
	missing = appendIfEmpty(missing, c.Server.Host, "missing Server.Host")
	missing = appendIfEmpty(missing, c.Server.Port, "missing Server.Port")
	missing = appendIfEmpty(missing, c.SQL.DriverName, "missing SQL.DriverName")
	missing = appendIfEmpty(missing, c.SQL.DataSourceName, "missing SQL.DataSourceName")
	missing = appendIfEmpty(missing, c.SMTP.Host, "missing SMTP.Host")
	missing = appendIfEmpty(missing, c.SMTP.Port, "missing SMTP.Port")
	missing = appendIfEmpty(missing, c.SMTP.User, "missing SMTP.User")
	missing = appendIfEmpty(missing, c.SMTP.Password, "missing SMTP.Password")

	return isValid && len(missing) == 0, missing
}

// RedactedConfig provides a redacted copy of Config for secure logging.
type RedactedConfig Config

// MarshalJSON customizes JSON marshalling to redact sensitive Config data.
func (c Config) MarshalJSON() ([]byte, error) {
	// Create a copy of Config that will contain the redacted data.
	r := RedactedConfig(c)

	// Redact sensitive data fields.
	r.SQL.DataSourceName = "[REDACTED]"
	r.SMTP.Password = "[REDACTED]"

	return json.Marshal(r)
}

// String returns string representation of Config with sensitive data redacted.
func (c Config) String() string {
	// Create a copy of Config that will contain the redacted data.
	r := RedactedConfig(c)

	// Redact sensitive data fields.
	r.SQL.DataSourceName = "[REDACTED]"
	r.SMTP.Password = "[REDACTED]"

	return fmt.Sprintf("%+v", r)
}

// TODO: slog.LogValuer for Config
