// Copyright 2023 Bill Nixon. All rights reserved.
// Use of this source code is governed by the license found in the LICENSE file.

package webapp

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

// ConfigLog holds the log settings.
type ConfigLog struct {
	Filename   string // Filename of log file.
	Type       string // Type of log, e.g., json or text.
	Level      string // Level of log, e.g., DEBUG, INFO, WARN, ERROR.
	WithSource bool   // WithSource add source info to log.
}

// ConfigServer holds the settings for the web server.
type ConfigServer struct {
	Host     string // Server host address.
	Port     string // Server port.
	CertFile string // CertFile is path to the cert file.
	KeyFile  string // KeyFile is path to the key file.
}

// Config represents the overall application configuration.
type Config struct {
	Name      string       // Name of the application.
	AssetsDir string       // AssetsDir is directory with server asets.
	Server    ConfigServer // Server configuration.
	Log       ConfigLog    // Log configuration.
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
	missing = appendIfEmpty(missing, c.Name, "missing Name")

	return len(missing) == 0, missing
}
