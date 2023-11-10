// Copyright 2023 Bill Nixon. All rights reserved.
// Use of this source code is governed by the license found in the LICENSE file.

package weblogin

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
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
	Host     string // SMTP server host address.
	Port     string // SMTP server port number.
	User     string // SMTP server username.
	Password string // SMTP server password.
}

// ConfigServer holds the settings for the web server.
type ConfigServer struct {
	Host string // Server host address.
	Port string // Server port.
}

// Config represents the overall application configuration.
type Config struct {
	Title            string       // Title of the application.
	BaseURL          string       // Base URL of the application (e.g., https://example.com).
	ParseGlobPattern string       // Glob pattern for parsing template files.
	SessionExpires   string       // Duration string for session expiry. See time#ParseDuration.
	Server           ConfigServer // Web server configuration.
	SQL              ConfigSQL    // SQL Database configuration.
	SMTP             ConfigSMTP   // SMTP server configuration.
}

// GetConfigFromFile loads configuration settings from a JSON file.
func GetConfigFromFile(filename string) (Config, error) {
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

	// Append errors for each missing mandatory field to help identify which are missing.
	missing = appendIfEmpty(missing, c.Title, "missing Title")
	missing = appendIfEmpty(missing, c.BaseURL, "missing BaseURL")
	missing = appendIfEmpty(missing, c.ParseGlobPattern, "missing ParseGlobPattern")
	missing = appendIfEmpty(missing, c.SessionExpires, "missing SessionExpires")
	missing = appendIfEmpty(missing, c.Server.Host, "missing Server.Host")
	missing = appendIfEmpty(missing, c.Server.Port, "missing Server.Port")
	missing = appendIfEmpty(missing, c.SQL.DriverName, "missing SQL.DriverName")
	missing = appendIfEmpty(missing, c.SQL.DataSourceName, "missing SQL.DataSourceName")
	missing = appendIfEmpty(missing, c.SMTP.Host, "missing SMTP.Host")
	missing = appendIfEmpty(missing, c.SMTP.Port, "missing SMTP.Port")
	missing = appendIfEmpty(missing, c.SMTP.User, "missing SMTP.User")
	missing = appendIfEmpty(missing, c.SMTP.Password, "missing SMTP.Password")

	return len(missing) == 0, missing
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

// String returns a string representation of Config with sensitive data redacted.
func (c Config) String() string {
	// Create a copy of Config that will contain the redacted data.
	r := RedactedConfig(c)

	// Redact sensitive data fields.
	r.SQL.DataSourceName = "[REDACTED]"
	r.SMTP.Password = "[REDACTED]"

	return fmt.Sprintf("%+v", r)
}

// TODO: slog.LogValuer for Config
