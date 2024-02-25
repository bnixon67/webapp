// Copyright 2024 Bill Nixon. All rights reserved.
// Use of this source code is governed by the license found in the LICENSE file.

package webauth_test

import (
	"errors"
	"math"
	"reflect"
	"strings"
	"testing"

	"github.com/bnixon67/webapp/webauth"
	"github.com/google/go-cmp/cmp"
)

func TestConfigFromJSONFile(t *testing.T) {
	testCases := []struct {
		name           string
		configFileName string
		wantErr        error
		wantConfig     webauth.Config
	}{
		{
			name:           "emptyFileName",
			configFileName: "",
			wantErr:        webauth.ErrConfigRead,
			wantConfig:     webauth.Config{},
		},
		{
			name:           "emptyJSON",
			configFileName: "testdata/empty.json",
			wantErr:        nil,
			wantConfig:     webauth.Config{},
		},
		{
			name:           "invalidJSON",
			configFileName: "testdata/invalid.json",
			wantErr:        webauth.ErrConfigUnmarshal,
			wantConfig:     webauth.Config{},
		},
		{
			name:           "validJSON",
			configFileName: "testdata/valid.json",
			wantErr:        nil,
			wantConfig: webauth.Config{
				BaseURL:      "test URL",
				LoginExpires: "42h",
				SQL: webauth.ConfigSQL{
					DriverName:     "testSQLDriverName",
					DataSourceName: "testSQLDataSourceName",
				},
				SMTP: webauth.ConfigSMTP{
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
			config, err := webauth.ConfigFromJSONFile(tc.configFileName)

			if tc.wantErr != nil && !errors.Is(err, tc.wantErr) || err != nil && tc.wantErr == nil {
				t.Fatalf("want error: %v, got: %v", tc.wantErr, err)
			}

			if diff := cmp.Diff(config, tc.wantConfig); diff != "" {
				t.Errorf("config mismatch for %q (-got +want):\n%s", tc.configFileName, diff)
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
		config   webauth.Config
		expected bool
	}

	var cases []tcase

	required := []string{
		"App.Name",
		"BaseURL",
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
		// Initialize a new instance of webauth.Config struct.
		config := webauth.Config{}

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

	for _, testCase := range cases {
		got, _ := testCase.config.IsValid()
		if got != testCase.expected {
			t.Errorf("c.IsValid(%+v) = %v; expected %v", testCase.config, got, testCase.expected)
		}
	}
}

func TestConfigMarshalJSON(t *testing.T) {
	input := webauth.Config{
		SQL: webauth.ConfigSQL{
			DataSourceName: "user:password@localhost/db",
		},
		SMTP: webauth.ConfigSMTP{
			Password: "supersecret",
		},
	}

	want := `{"App":{"Name":"","AssetsDir":"","TmplPattern":""},"Server":{"Host":"","Port":"","CertFile":"","KeyFile":""},"Log":{"Filename":"","Type":"","Level":"","AddSource":false},"BaseURL":"","LoginExpires":"","SQL":{"DriverName":"","DataSourceName":"[REDACTED]"},"SMTP":{"Host":"","Port":"","User":"","Password":"[REDACTED]"}}`

	testCases := []struct {
		name  string
		input webauth.Config
		want  string
	}{
		{
			name:  "test",
			input: input,
			want:  string(want),
		},
	}

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
		input webauth.Config
		want  string
	}{
		{
			name: "test",
			input: webauth.Config{
				SQL: webauth.ConfigSQL{
					DataSourceName: "user:password@localhost/db",
				},
				SMTP: webauth.ConfigSMTP{
					Password: "supersecret",
				},
			},
			want: `{Config:{App:{Name: AssetsDir: TmplPattern:} Server:{Host: Port: CertFile: KeyFile:} Log:{Filename: Type: Level: AddSource:false}} BaseURL: LoginExpires: SQL:{DriverName: DataSourceName:[REDACTED]} SMTP:{Host: Port: User: Password:[REDACTED]}}`,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			got := tc.input.String()
			if diff := cmp.Diff(got, tc.want); diff != "" {
				t.Errorf("config mismatch for (-got +want):\n%s", diff)
			}
		})
	}
}
