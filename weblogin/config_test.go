// Copyright 2023 Bill Nixon. All rights reserved.
// Use of this source code is governed by the license found in the LICENSE file.

package weblogin_test

import (
	"errors"
	"math"
	"reflect"
	"strings"
	"testing"

	"github.com/bnixon67/webapp/weblogin"
	"github.com/google/go-cmp/cmp"
)

func TestNewConfigFromFile(t *testing.T) {
	testCases := []struct {
		name           string
		configFileName string
		wantErr        error
		wantConfig     weblogin.Config
	}{
		{
			name:           "emptyFileName",
			configFileName: "",
			wantErr:        weblogin.ErrConfigOpen,
			wantConfig:     weblogin.Config{},
		},
		{
			name:           "emptyJSON",
			configFileName: "testdata/empty.json",
			wantErr:        nil,
			wantConfig:     weblogin.Config{},
		},
		{
			name:           "invalidJSON",
			configFileName: "testdata/invalid.json",
			wantErr:        weblogin.ErrConfigDecode,
			wantConfig:     weblogin.Config{},
		},
		{
			name:           "validJSON",
			configFileName: "testdata/valid.json",
			wantErr:        nil,
			wantConfig: weblogin.Config{
				Title:               "Test Title",
				BaseURL:             "test URL",
				ParseGlobPattern:    "testParseGlobPattern",
				SessionExpiresHours: 42,
				Server: weblogin.ConfigServer{
					Host: "test host",
					Port: "test port",
				},
				SQL: weblogin.ConfigSQL{
					DriverName:     "testSQLDriverName",
					DataSourceName: "testSQLDataSourceName",
				},
				SMTP: weblogin.ConfigSMTP{
					Host:     "test SMTP host",
					Port:     "test SMTP port",
					User:     "test SMTP user",
					Password: "test SMTP password",
				},
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			config, err := weblogin.GetConfigFromFile(tc.configFileName)

			if !errors.Is(err, tc.wantErr) {
				t.Fatalf("want err: %v, got err: %v", tc.wantErr, err)
			}

			if diff := cmp.Diff(config, tc.wantConfig); diff != "" {
				t.Errorf("config mismatch for %q (-got +want):\n%s",
					tc.configFileName, diff)
			}
		})
	}
}

func hasBit(n int, pos uint) bool {
	val := n & (1 << pos)
	return (val > 0)
}

func TestConfigIsValid(t *testing.T) {
	type tcase struct {
		config   weblogin.Config
		expected bool
	}

	var cases []tcase

	// required fields
	required := []string{
		"Title",
		"BaseURL",
		"ParseGlobPattern",
		"Server.Host",
		"Server.Port",
		"SQL.DriverName",
		"SQL.DataSourceName",
		"SMTP.Host",
		"SMTP.Port",
		"SMTP.User",
		"SMTP.Password",
	}

	// generate test cases based on required fields by looping thru all the possibilities and using bit logic to set fields
	for a := 0; a < int(math.Pow(2, float64(len(required)))); a++ {
		config := weblogin.Config{}

		for n := len(required) - 1; n >= 0; n-- {
			if hasBit(a, uint(n)) {
				f := strings.Split(required[n], ".")

				switch len(f) {
				case 1:
					reflect.ValueOf(&config).Elem().FieldByName(required[n]).SetString("x")
				case 2:
					v := "x"
					reflect.ValueOf(&config).Elem().FieldByName(f[0]).FieldByName(f[1]).SetString(v)
				}
			}
		}

		cases = append(cases, tcase{config, false})
	}
	// last case should be true since all required fields are present
	cases[len(cases)-1].expected = true

	for _, testCase := range cases {
		got, _ := testCase.config.IsValid()
		if got != testCase.expected {
			t.Errorf("c.IsValid(%+v) = %v; expected %v", testCase.config, got, testCase.expected)
		}
	}
}

func TestConfigMarshalJSON(t *testing.T) {
	testCases := []struct {
		name  string
		input weblogin.Config
		want  string
	}{
		{
			name: "test",
			input: weblogin.Config{
				Title: "AppConfig",
				SQL: weblogin.ConfigSQL{
					DataSourceName: "user:password@localhost/db",
				},
				SMTP: weblogin.ConfigSMTP{
					Password: "supersecret",
				},
			},
			want: `{"Title":"AppConfig","BaseURL":"","ParseGlobPattern":"","SessionExpiresHours":0,"Server":{"Host":"","Port":""},"SQL":{"DriverName":"","DataSourceName":"[REDACTED]"},"SMTP":{"Host":"","Port":"","User":"","Password":"[REDACTED]"}}`,
		},
	}

	// Run tests
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			got, err := tc.input.MarshalJSON()
			if err != nil {
				t.Fatalf("Error during MarshalJSON: %v", err)
			}
			if string(got) != tc.want {
				t.Errorf("got\n%s\n, want\n%s\n", got, tc.want)
			}
		})
	}
}

func TestConfigString(t *testing.T) {
	testCases := []struct {
		name  string
		input weblogin.Config
		want  string
	}{
		{
			name: "test",
			input: weblogin.Config{
				Title: "AppConfig",
				SQL: weblogin.ConfigSQL{
					DataSourceName: "user:password@localhost/db",
				},
				SMTP: weblogin.ConfigSMTP{
					Password: "supersecret",
				},
			},
			want: `{Title:AppConfig BaseURL: ParseGlobPattern: SessionExpiresHours:0 Server:{Host: Port:} SQL:{DriverName: DataSourceName:[REDACTED]} SMTP:{Host: Port: User: Password:[REDACTED]}}`,
		},
	}

	// Run tests
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			got := tc.input.String()
			if got != tc.want {
				t.Errorf("got\n%s\n, want\n%s\n", got, tc.want)
			}
		})
	}
}
