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
				BaseURL:          "test URL",
				ParseGlobPattern: "testParseGlobPattern",
				LoginExpires:     "42h",
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

// hasBit returns true if the bit at 'position' in 'n' is set (i.e., is 1).
func hasBit(n int, position uint) bool {
	// Perform a bitwise AND operation between n and a bit mask.
	// The bit mask is obtained by shifting 1 to the left 'pos' times.
	// This creates a number where only the bit at position 'pos' is set.
	val := n & (1 << position)

	// If bit at 'position' in 'n' is 1, then 'val' will be greater than 0,
	// since that is the only bit set in the bit mask.
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
		"BaseURL",
		"ParseGlobPattern",
		"LoginExpires",
		"SQL.DriverName",
		"SQL.DataSourceName",
		"SMTP.Host",
		"SMTP.Port",
		"SMTP.User",
		"SMTP.Password",
	}

	// Iterate over all possible combinations of settings in 'required'.
	// The number of combinations is 2 raised to the power of the
	// number of items in 'required' since each item in 'required'
	// can either be included or not in each combination.
	for a := 0; a < int(math.Pow(2, float64(len(required)))); a++ {
		// Initialize a new instance of weblogin.Config struct.
		config := weblogin.Config{}

		// Iterate over each item in 'required' in reverse order.
		for n := len(required) - 1; n >= 0; n-- {
			// Check if the bit at position 'n' in 'a' is set.
			// This determines whether to include the 'n'th item
			// of 'required' in this combination.
			if hasBit(a, uint(n)) {
				// Split the 'n'th required item by '.'
				// to handle nested fields.
				f := strings.Split(required[n], ".")

				// Depending on the number of parts after
				// splitting, set the corresponding field in
				// 'config' to a predefined value ('x').
				switch len(f) {
				case 1: // top-level field
					reflect.ValueOf(&config).Elem().FieldByName(required[n]).SetString("x")
				case 2: // nested field
					v := "x"
					reflect.ValueOf(&config).Elem().FieldByName(f[0]).FieldByName(f[1]).SetString(v)
				}
			}
		}

		// Add the modified 'config' to 'cases'.
		cases = append(cases, tcase{config, false})
	}
	// last case should be true since all required fields are present
	//cases[len(cases)-1].expected = true

	for _, testCase := range cases {
		got, _ := testCase.config.IsValid()
		if got != testCase.expected {
			t.Errorf("c.IsValid(%+v) = %v; expected %v", testCase.config, got, testCase.expected)
		}
	}
}

func TestConfigMarshalJSON(t *testing.T) {
	input := weblogin.Config{
		SQL: weblogin.ConfigSQL{
			DataSourceName: "user:password@localhost/db",
		},
		SMTP: weblogin.ConfigSMTP{
			Password: "supersecret",
		},
	}

	want := `{"App":{"Name":"","AssetsDir":""},"Server":{"Host":"","Port":"","CertFile":"","KeyFile":""},"Log":{"Filename":"","Type":"","Level":"","WithSource":false},"BaseURL":"","ParseGlobPattern":"","LoginExpires":"","SQL":{"DriverName":"","DataSourceName":"[REDACTED]"},"SMTP":{"Host":"","Port":"","User":"","Password":"[REDACTED]"}}`

	testCases := []struct {
		name  string
		input weblogin.Config
		want  string
	}{
		{
			name:  "test",
			input: input,
			want:  string(want),
		},
	}

	// Run tests
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			got, err := tc.input.MarshalJSON()
			if err != nil {
				t.Fatalf("Error during MarshalJSON: %v", err)
			}
			if diff := cmp.Diff(string(got), tc.want); diff != "" {
				t.Errorf("config mismatch for (-got +want):\n%s", diff)
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
				SQL: weblogin.ConfigSQL{
					DataSourceName: "user:password@localhost/db",
				},
				SMTP: weblogin.ConfigSMTP{
					Password: "supersecret",
				},
			},
			want: `{Config:{App:{Name: AssetsDir:} Server:{Host: Port: CertFile: KeyFile:} Log:{Filename: Type: Level: WithSource:false}} BaseURL: ParseGlobPattern: LoginExpires: SQL:{DriverName: DataSourceName:[REDACTED]} SMTP:{Host: Port: User: Password:[REDACTED]}}`,
		},
	}

	// Run tests
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			got := tc.input.String()
			if diff := cmp.Diff(got, tc.want); diff != "" {
				t.Errorf("config mismatch for (-got +want):\n%s", diff)
			}
		})
	}
}
