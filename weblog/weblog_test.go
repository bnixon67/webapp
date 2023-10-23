package weblog_test

import (
	"errors"
	"log/slog"
	"os"
	"reflect"
	"testing"

	"github.com/bnixon67/webapp/weblog"
)

func TestInit(t *testing.T) {
	tests := []struct {
		name    string
		opts    []weblog.Option
		wantErr error
	}{
		{
			name: "Invalid Log Type",
			opts: []weblog.Option{
				weblog.WithLogType("invalid"),
			},
			wantErr: weblog.ErrInvalidLogType,
		},
		{
			name: "Invalid Log Level",
			opts: []weblog.Option{
				weblog.WithLevel("invalid"),
			},
			wantErr: weblog.ErrInvalidLogLevel,
		},
		{
			name: "Valid JSON Log Type",
			opts: []weblog.Option{
				weblog.WithLogType("json"),
			},
			wantErr: nil,
		},
		{
			name: "Valid JSON Log Type with Filename",
			opts: []weblog.Option{
				weblog.WithLogType("json"),
				weblog.WithFilename("test_json.log"),
			},
			wantErr: nil,
		},
		{
			name: "Valid Text Log Type",
			opts: []weblog.Option{
				weblog.WithLogType("text"),
			},
			wantErr: nil,
		},
		{
			name: "Valid Text Log Type with Filename",
			opts: []weblog.Option{
				weblog.WithLogType("text"),
				weblog.WithFilename("test_text.log"),
			},
			wantErr: nil,
		},
		{
			name: "Valid Filename",
			opts: []weblog.Option{
				weblog.WithFilename("test.log"),
			},
			wantErr: nil,
		},
		{
			name: "Invalid Filename",
			opts: []weblog.Option{
				weblog.WithFilename("/no/such/file"),
			},
			wantErr: weblog.ErrLogFileOpenError,
		},
		// Add more test cases as needed.
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := weblog.Init(tt.opts...)
			if (err != nil) != (tt.wantErr != nil) {
				t.Errorf("Init() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if errors.Is(err, tt.wantErr) == false {
				t.Errorf("Init() error = %v, wantErr %v", err, tt.wantErr)
			}

			// Cleanup, remove log file if it was created
			if err == nil {
				config := &weblog.Config{}
				for _, opt := range tt.opts {
					opt(config)
				}

				if config.Filename != "" {
					os.Remove(config.Filename) // ignore error
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
		// Add more test cases as needed
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := weblog.Level(tt.input)

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
	tests := []struct {
		name string
		want string
	}{
		{
			name: "Sorted Log Levels",
			want: "DEBUG,INFO,WARN,ERROR",
		},
		// Add more test cases as needed
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := weblog.Levels(); got != tt.want {
				t.Errorf("Levels() = %v, want %v", got, tt.want)
			}
		})
	}
}
