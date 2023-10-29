// Copyright 2023 Bill Nixon. All rights reserved.
// Use of this source code is governed by the license found in the LICENSE file.

package webhandler_test

import (
	"strings"
	"testing"
	"unicode/utf8"

	"github.com/bnixon67/webapp/webhandler"
)

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

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := webhandler.RandomStringFromCharset(tt.charset, tt.length)
			if (err != nil) != tt.expectError {
				t.Errorf("RandomStringFromCharset() error = %v, expectError %v", err, tt.expectError)
				return
			}

			if err == nil {
				if utf8.RuneCountInString(got) != tt.length {
					t.Errorf("RandomStringFromCharset() got = %v (%v), want length %v", utf8.RuneCountInString(got), got, tt.length)
				}
				if !isSubset(got, tt.charset) {
					t.Errorf("RandomStringFromCharset() got characters not in charset, got = %v, charset = %v", got, tt.charset)
				}
			}
		})
	}
}

// isSubset checks if all characters in the 'str' are in the 'charset'.
func isSubset(str, charset string) bool {
	for _, r := range str {
		if !strings.ContainsRune(charset, r) {
			return false
		}
	}
	return true
}
