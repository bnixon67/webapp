// Copyright 2024 Bill Nixon. All rights reserved.
// Use of this source code is governed by the license found in the LICENSE file.

package csv_test

import (
	"bytes"
	"errors"
	"testing"

	"github.com/bnixon67/webapp/csv"
	"github.com/google/go-cmp/cmp"
)

// Custom writer that fails on write
type failWriter struct{}

func (fw *failWriter) Write(p []byte) (n int, err error) {
	return 0, errors.New("write failed")
}

// TestSliceOfStructsToCSV tests the SliceOfStructsToCSV function.
func TestSliceOfStructsToCSV(t *testing.T) {
	type User struct {
		ID    int
		Name  string `csv:"Full Name"`
		Email string
	}

	testCases := []struct {
		name    string
		input   interface{}
		want    string
		wantErr error
	}{
		{
			name: "Valid slice of structs",
			input: []User{
				{1, "Alice", "alice@example.com"},
				{2, "Bob", "bob@example.com"},
			},
			want:    "ID,Full Name,Email\n1,Alice,alice@example.com\n2,Bob,bob@example.com\n",
			wantErr: nil,
		},
		{
			name:    "Not a slice",
			input:   User{1, "Alice", "alice@example.com"},
			want:    "",
			wantErr: csv.ErrCSVNotSlice,
		},
		{
			name:    "Slice with non-struct elements",
			input:   []int{1, 2, 3},
			want:    "",
			wantErr: csv.ErrCSVNotSliceOfStructs,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			var buf bytes.Buffer
			err := csv.SliceOfStructsToCSV(&buf, tc.input)

			if !errors.Is(err, tc.wantErr) {
				t.Errorf("SliceOfStructsToCSV() error = %v, wantErr %v", err, tc.wantErr)
				return
			}
			got := buf.String()
			if diff := cmp.Diff(tc.want, got); diff != "" {
				t.Errorf("(-want +got)\n%s", diff)
			}
		})
	}

	t.Run("Error writing CSV", func(t *testing.T) {
		err := csv.SliceOfStructsToCSV(&failWriter{}, []User{{1, "Alice", "alice@example.com"}})
		if !errors.Is(err, csv.ErrCSVWriteFailed) {
			t.Errorf("SliceOfStructsToCSV() error = %v, wantErr %v", err, csv.ErrCSVWriteFailed)
		}
	})
}
