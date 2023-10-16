package log

import (
	"log/slog"
	"os"
	"testing"
)

func TestInit(t *testing.T) {
	tests := []struct {
		name        string
		fileName    string
		logType     string
		level       slog.Level
		addSource   bool
		expectError bool
	}{
		{"ValidTextLog", "test.log", "text", slog.LevelInfo, false, false},
		{"ValidJSONLog", "test.log", "json", slog.LevelInfo, false, false},
		{"InvalidLogType", "test.log", "invalid", slog.LevelInfo, false, false},
		{"StderrTextLog", "", "text", slog.LevelInfo, false, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := Init(tt.fileName, tt.logType, tt.level, tt.addSource)
			if (err != nil) != tt.expectError {
				t.Errorf("Init() error = %v, expectError %v", err, tt.expectError)
				return
			}

			// Clean up if a log file is created
			if tt.fileName != "" {
				os.Remove(tt.fileName)
			}
		})
	}
}

func TestLevel(t *testing.T) {
	tests := []struct {
		name        string
		input       string
		expected    slog.Level
		expectError bool
	}{
		{"ValidDebug", "DEBUG", slog.LevelDebug, false},
		{"ValidInfo", "INFO", slog.LevelInfo, false},
		{"ValidWarn", "WARN", slog.LevelWarn, false},
		{"InvalidLevel", "INVALID", slog.LevelInfo, true}, // Assuming it returns an error
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := Level(tt.input)
			if (err != nil) != tt.expectError {
				t.Errorf("Level() error = %v, expectError %v", err, tt.expectError)
				return
			}
			if got != tt.expected {
				t.Errorf("Level() = %v, want %v", got, tt.expected)
			}
		})
	}
}
