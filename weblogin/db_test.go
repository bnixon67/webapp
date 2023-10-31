// Copyright 2023 Bill Nixon. All rights reserved.
// Use of this source code is governed by the license found in the LICENSE file.

package weblogin_test

import (
	"database/sql"
	"database/sql/driver"
	"errors"
	"testing"

	"github.com/bnixon67/webapp/weblogin"
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
			wantErr:        weblogin.ErrDBOpen,
		},
		{
			name:           "invalid source",
			driverName:     "mock_driver",
			dataSourceName: "invalid_source",
			wantErr:        weblogin.ErrDBPing,
		},
	}

	// Create a map to mock the sql.Open function
	mockDriver := &CustomMockDriver{}
	sql.Register("mock_driver", mockDriver)

	// Run tests
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			_, err := weblogin.InitDB(tc.driverName, tc.dataSourceName)

			if !errors.Is(err, tc.wantErr) {
				t.Errorf("got err %q, want %q for InitDB(%q, %q)", err, tc.wantErr, tc.driverName, tc.dataSourceName)
			}
		})
	}
}
