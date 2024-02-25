// Copyright 2023 Bill Nixon. All rights reserved.
// Use of this source code is governed by the license found in the LICENSE file.

package webapp_test

import (
	"errors"
	"math"
	"reflect"
	"strings"
	"testing"

	"github.com/bnixon67/webapp/webapp"
	"github.com/bnixon67/webapp/weblog"
	"github.com/bnixon67/webapp/webserver"
	"github.com/google/go-cmp/cmp"
)

// TestConfigFromJSONFile tests the ConfigFromJSONFile function.
func TestConfigFromJSONFile(t *testing.T) {
	var emptyConfig webapp.Config

	testCases := []struct {
		name           string
		configFileName string
		wantErr        error
		wantConfig     webapp.Config
	}{
		{
			name:           "emptyFileName",
			configFileName: "",
			wantErr:        webapp.ErrConfigOpen,
			wantConfig:     emptyConfig,
		},
		{
			name:           "emptyJSON",
			configFileName: "testdata/empty.json",
			wantErr:        nil,
			wantConfig:     emptyConfig,
		},
		{
			name:           "invalidJSON",
			configFileName: "testdata/invalid.json",
			wantErr:        webapp.ErrConfigDecode,
			wantConfig:     emptyConfig,
		},
		{
			name:           "validJSON",
			configFileName: "testdata/valid.json",
			wantErr:        nil,
			wantConfig: webapp.Config{
				App: webapp.ConfigApp{
					Name: "Test Name",
				},
			},
		},
		{
			name:           "allJSON",
			configFileName: "testdata/all.json",
			wantErr:        nil,
			wantConfig: webapp.Config{
				App: webapp.ConfigApp{
					Name:        "Test Name",
					AssetsDir:   "directory",
					TmplPattern: "*.html",
				},
				Server: webserver.Config{
					Host:     "localhost",
					Port:     "8080",
					CertFile: "cert.pem",
					KeyFile:  "key.pem",
				},
				Log: &weblog.Config{
					Filename:  "log.txt",
					Type:      "text",
					Level:     "debug",
					AddSource: true,
				},
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			config, err := webapp.ConfigFromJSONFile(tc.configFileName)

			if !errors.Is(err, tc.wantErr) {
				t.Fatalf("got err: %v, want err: %v", err, tc.wantErr)
			}

			if diff := cmp.Diff(tc.wantConfig, config); diff != "" {
				t.Errorf("config mismatch for %q (-want +got):\n%s", tc.configFileName, diff)
			}
		})
	}
}

// hasBit returns true if the bit at 'position' in 'n' is set.
func hasBit(n int, position uint) bool {
	// Perform a bitwise AND operation between n and a bit mask.
	// The bit mask is obtained by shifting 1 to the left 'pos' times.
	// This creates a number where only the bit at position 'pos' is set.
	val := n & (1 << position)

	// If bit at 'position' in 'n' is 1, then 'val' will be greater than 0,
	// since that is the only bit set in the bit mask.
	return (val > 0)
}

// TestConfigIsValid tests the IsValid function.
func TestConfigIsValid(t *testing.T) {
	type tc struct {
		config webapp.Config
		want   bool
	}

	var cases []tc

	// required fields
	required := []string{
		"App.Name",
	}

	// Iterate over all possible combinations of settings in 'required'.
	// The number of combinations is 2 raised to the power of the
	// number of items in 'required' since each item in 'required'
	// can either be included or not in each combination.
	for a := 0; a < int(math.Pow(2, float64(len(required)))); a++ {
		// Initialize a new instance of webapp.Config struct.
		config := webapp.Config{}

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
		cases = append(cases, tc{config, false})
	}
	// last case should be true since all required fields are present
	cases[len(cases)-1].want = true

	for _, testCase := range cases {
		got, _ := testCase.config.IsValid()
		if got != testCase.want {
			t.Errorf("got %v, want %v for c.IsValid(%+v)", got, testCase.want, testCase.config)
		}
	}
}
