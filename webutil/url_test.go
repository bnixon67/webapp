// Copyright 2024 Bill Nixon. All rights reserved.
// Use of this source code is governed by the license found in the LICENSE file.

package webutil_test

import (
	"testing"

	"github.com/bnixon67/webapp/webutil"
)

func TestIsLocalURL(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  bool
	}{
		{"Relative path", "/local/path", true},
		{"Absolute URL", "http://example.com/path", false},
		{"URL with host", "//example.com/path", false},
		{"Empty URL", "", false},
		{"URL with traversal", "/local/../path", false},
		{"Root URL", "/", true},
		{"Invalid URL", "http://example.com/../", false},
		{"Only host", "example.com", false},
		{"Path with double dots", "/..", false},
		{"Path starting with dot", "/./path", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := webutil.IsLocalSafeURL(tt.input)
			if got != tt.want {
				t.Errorf("isLocalURL(%q) = %v; want %v", tt.input, got, tt.want)
			}
		})
	}
}
