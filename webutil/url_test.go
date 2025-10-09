// Copyright 2025 Bill Nixon. All rights reserved.
// Use of this source code is governed by the license found in the LICENSE file.

package webutil_test

import (
	"strings"
	"testing"

	"github.com/bnixon67/webapp/webutil"
)

func TestIsLocalURL(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		wantSafe string
		wantOk   bool
	}{
		{"Relative path", "/local/path", "/local/path", true},
		{"Path not starting with /", "local/path", "", false},
		{"Backslashes", "\\foo", "", false},
		{"Absolute URL", "http://example.com/path", "", false},
		{"URL with host", "//example.com/path", "", false},
		{"Empty URL", "", "", false},
		{"URL with traversal", "/local/../path", "/path", true},
		{"Root URL", "/", "/", true},
		{"Invalid URL", "http://example.com/../", "", false},
		{"Only host", "example.com", "", false},
		{"Path with double dots", "/..", "/", true},
		{"Path starting with dot", "/./path", "/path", true},
		{"JS scheme", "javascript:alert(1)", "", false},
		{"Mailto scheme", "mailto:a@b", "", false},
		{"Scheme-relative", "//evil.com/x", "", false},

		{"Double slashes in path", "///a//b", "/a/b", true},
		{"Dot segment mid-path", "/a/./b", "/a/b", true},
		{"Encoded traversal mid", "/a/%2e%2E/b", "/b", true},
		{"Encoded traversal at root", "/%2e%2e", "/", true},

		{"Encoded backslash", "/a%5cb", "", false},
		{"Literal backslash later", "/a\\b", "", false},

		{"Encoded slash in path", "/a%2Fb", "/a/b", true},

		{"Keep query", "/search?q=go+path", "/search?q=go+path", true},
		{"Keep fragment", "/a#frag", "/a#frag", true},
		{"Query + fragment", "/a?x=1#f", "/a?x=1#f", true},
		{"Only query", "?x=1", "", false},

		{"Relative dotdot", "../x", "", false},
		{"Starts with single dot", "./x", "", false},

		{"Space encoded", "/a%20b", "/a b", true},
		{"Unicode in path", "/caf%C3%A9", "/café", true},

		{"Empty string", "", "", false},
		{"Only host text", "example.com", "", false},
		{"Absolute URL https", "https://ex.com/x", "", false},
		{"Absolute URL with port", "http://ex.com:8080/x", "", false},
		{"IPv6 host", "http://[::1]/x", "", false},

		{"Invalid percent-encoding", "/%zz", "", false},
		{"Truncated escape", "/foo%2", "", false},
		{"Null byte in path", "/\x00foo", "", false},
		{"Malformed scheme", "http ://example.com", "", false},
		{"Bare colon", "://missing", "", false},

		{"Invalid hex in escape", "/foo%2Zbar", "", false},
		{"Invalid hex", "/foo%G1", "", false},
		{"Unescaped space", "/foo bar", "", false},
		{"Encoded space", "/foo%20bar", "/foo bar", true},
		{"Tab in input", "/foo\tbar", "", false},

		{"Bare percent at end", "/foo%", "", false}, // invalid escape
		{"Invalid hex (G)", "/foo%G1", "", false},   // non-hex
		{"Invalid hex (zz)", "/foo%zz", "", false},  // non-hex
		{"Root bare percent", "/%", "", false},      // invalid escape at root
		{"Bad escape in path, ok query", "/x%2?a=%20", "", false},

		{"Tab character", "/foo%09bar", "", false},     // %09 = horizontal tab
		{"Newline character", "/foo%0Abar", "", false}, // %0A = line feed
		{"Carriage return", "/foo%0Dbar", "", false},   // %0D = CR
		{"DEL character", "/foo%7Fbar", "", false},     // 0x7F

		{"Fragment only", "#frag", "", false},
		{"Path with only dot", ".", "", false},
		{"Path with 'C:' segment", "/C:/windows", "/C:/windows", true},
		{"Mixed case traversal (decoded)", "/a/%2E./b", "/b", true},
		{"Mixed case single dot", "/a/%2E/b", "/a/b", true}, // %2E => "."
		{"Mixed case traversal", "/a/%2E./b", "/b", true},   // %2E. => ".."
		{"Encoded tab (control)", "/a%09b", "", false},
		{"Encoded NUL (control)", "/a%00b", "", false},
		{"Normalize then keep query", "/a//b/../c?x=1", "/a/c?x=1", true},
		{"Normalize many slashes+query", "////a///b/./../c?z=1", "/a/c?z=1", true},

		{"Collapse to root", "/a/..", "/", true},
		{"Collapse deep above root", "/a/../../b", "/b", true},

		{"Query preserved", "/a?x=1&y=%2B", "/a?x=1&y=%2B", true},
		{"Fragment preserved", "/a#sec-1", "/a#sec-1", true},

		{"Non-ASCII byte", "/foo%ba", "/foo\xba", true},

		{"Control in query (US)", "/a?\x1e=1", "", false}, // 0x1E
		{"DEL in query", "/a?x=\x7f", "", false},          // 0x7F
		{"Control in fragment", "/a#\x1e", "", false},     // 0x1E
		{"DEL in fragment", "/a#\x7f", "", false},         // 0x7F

	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotSafe, gotOk := webutil.ValidateLocalRedirect(tt.input)
			if (gotSafe != tt.wantSafe) || (gotOk != tt.wantOk) {
				t.Errorf("ValidateLocalRedirect(%q) = (%q, %v); want (%q, %v)",
					tt.input, gotSafe, gotOk, tt.wantSafe, tt.wantOk)
			}
		})
	}
}

func FuzzValidateLocalRedirect(f *testing.F) {
	seeds := []string{
		"/",
		"/a",
		"/a/b",
		"/a?x=1#y",
		"/a%2Fb",
		"/a%2e%2e/b",
		"/a%5cb",
		"/foo%0Abar",
		"/foo%",
		"/foo%2",
		"/foo%G1",
		"http://ex.com/x",
		"//ex.com/x",
		" ",
		"\\",
		"?x=1",
	}
	for _, s := range seeds {
		f.Add(s)
	}

	f.Fuzz(func(t *testing.T, s string) {
		safe, ok := webutil.ValidateLocalRedirect(s)
		if ok {
			if !strings.HasPrefix(safe, "/") {
				t.Fatalf("safe path must start with '/': %q", safe)
			}
			// No control chars in the cleaned path
			for _, r := range safe {
				if r < 0x20 || r == 0x7f {
					t.Fatalf("control char leaked in safe path: %q", safe)
				}
			}
			if strings.ContainsRune(safe, '\\') {
				t.Fatalf("backslash leaked in safe path: %q", safe)
			}
		}
	})
}
