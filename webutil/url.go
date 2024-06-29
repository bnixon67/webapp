// Copyright 2024 Bill Nixon. All rights reserved.
// Use of this source code is governed by the license found in the LICENSE file.

package webutil

import (
	"net/url"
	"strings"
)

// IsLocalSafeURL checks if a given URL is local and safe to use for
// redirection.  It ensures the URL has no scheme or host and does not
// attempt to traverse directories. This helps to avoid cross-site or redirect
// attacks by validating that the URL is confined to the local domain.
func IsLocalSafeURL(redirectURL string) bool {
	// Ensure the path does not contain any ".."
	if strings.Contains(redirectURL, "..") {
		return false
	}

	u, err := url.Parse(redirectURL)
	if err != nil {
		return false
	}

	// Ensure the URL has no scheme and no host component
	if u.Scheme != "" || u.Host != "" {
		return false
	}

	// Ensure the URL path starts with "/"
	if u.Path == "" || u.Path[0] != '/' {
		return false
	}

	return true
}
