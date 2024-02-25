// Copyright 2023 Bill Nixon. All rights reserved.
// Use of this source code is governed by the license found in the LICENSE file.

package weblog_test

import (
	"errors"
	"log/slog"
	"os"
	"reflect"
	"testing"

	"github.com/bnixon67/webapp/weblog"
)

func TestInitFromConfig(t *testing.T) {
	tests := []struct {
		name    string
		cfg     weblog.Config
		wantErr error
	}{
		{
			name:    "Invalid Log Type",
			cfg:     weblog.Config{Type: "invalid"},
			wantErr: weblog.ErrInvalidLogType,
		},
		{
			name:    "Invalid Log Level",
			cfg:     weblog.Config{Level: "invalid"},
			wantErr: weblog.ErrInvalidLogLevel,
		},
		{
			name:    "Valid JSON Log Type",
			cfg:     weblog.Config{Type: "json"},
			wantErr: nil,
		},
		{
			name: "Valid JSON Log Type with Filename",
			cfg: weblog.Config{
				Type:     "json",
				Filename: "test_json.log",
			},
			wantErr: nil,
		},
		{
			name:    "Valid Text Log Type",
			cfg:     weblog.Config{Type: "text"},
			wantErr: nil,
		},
		{
			name: "Valid Text Log Type with Filename",
			cfg: weblog.Config{
				Type:     "text",
				Filename: "test_text.log",
			},
			wantErr: nil,
		},
		{
			name:    "Valid Filename",
			cfg:     weblog.Config{Filename: "test.log"},
			wantErr: nil,
		},
		{
			name:    "Invalid Filename",
			cfg:     weblog.Config{Filename: "/no/such/file"},
			wantErr: weblog.ErrOpenLogFile,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := weblog.Init(tt.cfg)
			if (err != nil) != (tt.wantErr != nil) {
				t.Errorf("Init() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if errors.Is(err, tt.wantErr) == false {
				t.Errorf("Init() error = %v, wantErr %v", err, tt.wantErr)
			}

			// TODO: test log output

			// Cleanup, remove log file if it was created
			if err == nil {
				if tt.cfg.Filename != "" {
					// ignore error
					os.Remove(tt.cfg.Filename)
				}
			}
		})
	}
}

func TestLevel(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		want    slog.Level
		wantErr error
	}{
		{
			name:    "Valid DEBUG Level",
			input:   "DEBUG",
			want:    slog.LevelDebug,
			wantErr: nil,
		},
		{
			name:    "Valid INFO Level",
			input:   "INFO",
			want:    slog.LevelInfo,
			wantErr: nil,
		},
		{
			name:    "Invalid Level",
			input:   "INVALID",
			want:    slog.LevelInfo,
			wantErr: weblog.ErrInvalidLogLevel,
		},
		{
			name:    "Lowercase Level",
			input:   "warn",
			want:    slog.LevelWarn,
			wantErr: nil,
		},
		{
			name:    "Empty Level",
			input:   "",
			wantErr: nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := weblog.ParseLevel(tt.input)
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("Level() got = %v, want %v", got, tt.want)
			}

			if (err != nil) != (tt.wantErr != nil) { // Simplifying error checking
				t.Errorf("Level() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if err != nil && !errors.Is(err, tt.wantErr) {
				t.Errorf("Level() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestLevels(t *testing.T) {
	want := []string{"DEBUG", "INFO", "WARN", "ERROR"}

	if got := weblog.Levels(); !reflect.DeepEqual(got, want) {
		t.Errorf("got %q, want %q", got, want)
	}
}
