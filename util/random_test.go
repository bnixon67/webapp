// Copyright 2023 Bill Nixon. All rights reserved.
// Use of this source code is governed by the license found in the LICENSE file.

package util_test

import (
	"strings"
	"testing"
	"unicode/utf8"

	"github.com/bnixon67/webapp/util"
)

// isSubset checks if all characters in the 'str' are in the 'charset'.
func isSubset(str, charset string) bool {
	for _, r := range str {
		if !strings.ContainsRune(charset, r) {
			return false
		}
	}
	return true
}

func TestRandomStringFromCharset(t *testing.T) {
	tests := []struct {
		name        string
		length      int
		charset     string
		expectError bool
	}{
		{
			name:        "Standard ASCII charset",
			length:      5,
			charset:     "abcdefghijklmnopqrstuvwxyz",
			expectError: false,
		},
		{
			name:        "Empty charset",
			length:      5,
			charset:     "",
			expectError: true,
		},
		{
			name:        "Zero length",
			length:      0,
			charset:     "abcdefghijklmnopqrstuvwxyz",
			expectError: true,
		},
		{
			name:        "Negative length",
			length:      -1,
			charset:     "abcdefghijklmnopqrstuvwxyz",
			expectError: true,
		},
		{
			name:        "Unicode charset",
			length:      5,
			charset:     "日本語𠀋𡃁𠮷",
			expectError: false,
		},
		{
			name:        "Single character charset",
			length:      5,
			charset:     "a",
			expectError: false,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got, err := util.RandomStringFromCharset(tc.charset, tc.length)
			if (err != nil) != tc.expectError {
				t.Errorf("RandomStringFromCharset() error = %v, expectError %v", err, tc.expectError)
				return
			}

			if err == nil {
				if utf8.RuneCountInString(got) != tc.length {
					t.Errorf("RandomStringFromCharset() got = %v (%v), want length %v", utf8.RuneCountInString(got), got, tc.length)
				}
				if !isSubset(got, tc.charset) {
					t.Errorf("RandomStringFromCharset() got characters not in charset, got = %v, charset = %v", got, tc.charset)
				}
			}
		})
	}
}
