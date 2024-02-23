// Copyright 2023 Bill Nixon. All rights reserved.
// Use of this source code is governed by the license found in the LICENSE file.

package webauth_test

import (
	"database/sql"
	"database/sql/driver"
	"errors"
	"testing"

	"github.com/bnixon67/webapp/webauth"
)

// CustomMockDriver is a custom driver that only returns a connection error.
type CustomMockDriver struct{}

func (d CustomMockDriver) Open(name string) (driver.Conn, error) {
	if name == "valid_source" {
		return nil, nil
	}
	return nil, errors.New("mock connection error")
}

func TestInitDB(t *testing.T) {
	// Define test cases
	testCases := []struct {
		name           string
		driverName     string
		dataSourceName string
		wantErr        error
	}{
		{
			name:           "valid",
			driverName:     "mock_driver",
			dataSourceName: "valid_source",
			wantErr:        nil,
		},
		{
			name:           "invalid driver",
			driverName:     "invalid_driver",
			dataSourceName: "valid_source",
			wantErr:        webauth.ErrInitDBOpen,
		},
		{
			name:           "invalid source",
			driverName:     "mock_driver",
			dataSourceName: "invalid_source",
			wantErr:        webauth.ErrInitDBPing,
		},
	}

	// Create a map to mock the sql.Open function
	mockDriver := &CustomMockDriver{}
	sql.Register("mock_driver", mockDriver)

	// Run tests
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			_, err := webauth.InitDB(tc.driverName, tc.dataSourceName)

			if !errors.Is(err, tc.wantErr) {
				t.Errorf("got err %q, want %q for InitDB(%q, %q)", err, tc.wantErr, tc.driverName, tc.dataSourceName)
			}
		})
	}
}

func TestRowExists(t *testing.T) {
	a := AppForTest(t)
	db := a.DB

	// Define test cases
	tests := []struct {
		name    string
		query   string
		args    []interface{}
		want    bool
		wantErr bool
	}{
		{
			name:    "RowExists",
			query:   "SELECT 1 FROM users WHERE username = ?",
			args:    []interface{}{"test"},
			want:    true,
			wantErr: false,
		},
		{
			name:    "RowDoesNotExist",
			query:   "SELECT 1 FROM users WHERE username = ?",
			args:    []interface{}{"nosuchuser"},
			want:    false,
			wantErr: false,
		},
		{
			name:    "BadQuery",
			query:   "SELECT 1 FROM nosuchtable WHERE username = ?",
			args:    []interface{}{"nosuchuser"},
			want:    false,
			wantErr: true,
		},
		// Add more test cases as needed
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := db.RowExists(tt.query, tt.args...)
			if (err != nil) != tt.wantErr {
				t.Errorf("RowExists() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("RowExists() = %v, want %v", got, tt.want)
			}
		})
	}
}
