// Copyright 2025 Bill Nixon. All rights reserved.
// Use of this source code is governed by the license found in the LICENSE file.

package webutil

import (
	"net/url"
	"path"
	"strings"
	"unicode"
)

// hasCtrl reports whether the string contains ASCII control characters
// (U+0000–U+001F or U+007F). Used to block invisible or non-printable input.
func hasCtrl(s string) bool {
	for _, r := range s {
		if r < 0x20 || r == 0x7f {
			return true
		}
	}
	return false
}

// ValidateLocalRedirect checks whether redirectURL is a safe local redirect
// target.
//
// It verifies that the URL:
//   - Contains no scheme or host (so it can’t point off-site)
//   - Begins with “/” (absolute local path)
//   - Contains no whitespace, control characters, or backslashes
//   - Decodes cleanly from any percent-encoding
//   - Produces a normalized path via path.Clean
//   - Has query and fragment parts free of spaces, control chars, or backslashes
//
// On success it returns the cleaned URL (path plus original query/fragment)
// and true. On failure it returns an empty string and false.
func ValidateLocalRedirect(redirectURL string) (string, bool) {
	// Reject unescaped whitespace in raw input (spaces, tabs, newlines, etc.).
	if strings.IndexFunc(redirectURL, unicode.IsSpace) >= 0 {
		return "", false
	}

	// Reject backslashes in raw input to prevent ambiguous path parsing.
	if strings.ContainsRune(redirectURL, '\\') {
		return "", false
	}

	u, err := url.Parse(redirectURL)
	if err != nil {
		// Malformed URL (bad escapes, invalid syntax).
		return "", false
	}

	// Reject URLs with a scheme or host (absolute or protocol-relative).
	if u.Scheme != "" || u.Host != "" {
		return "", false
	}

	// Reject relative path like "../foo".
	if !strings.HasPrefix(u.Path, "/") {
		return "", false
	}

	// Prefer RawPath if present; fall back to Path.
	// RawPath may preserve original escapes that Path already decoded.
	raw := u.RawPath
	if raw == "" {
		raw = u.Path
	}

	// Decode percent-encoded sequences; reject if invalid.
	decoded, err := url.PathUnescape(raw)
	if err != nil {
		return "", false
	}

	// Reject any backslash that appears after decoding (e.g., %5C).
	if strings.ContainsRune(decoded, '\\') {
		return "", false
	}

	// Reject control characters in the decoded path.
	if hasCtrl(decoded) {
		return "", false
	}

	// Canonicalize the path to remove redundant elements like "/./" or "/a/../".
	clean := path.Clean(decoded)

	// Validate query and fragment: disallow control characters.
	if hasCtrl(u.RawQuery) || hasCtrl(u.Fragment) {
		return "", false
	}

	// Reconstruct a safe URL: cleaned path plus validated query/fragment.
	safe := clean
	if u.RawQuery != "" {
		safe += "?" + u.RawQuery
	}
	if u.Fragment != "" {
		safe += "#" + u.Fragment
	}
	return safe, true
}
