// Copyright 2024 Bill Nixon. All rights reserved.
// Use of this source code is governed by the license found in the LICENSE file.

package webhandler

import (
	"net/http"
)

// AddSecurityHeaders returns middleware that applies essential security
// headers to HTTP responses to enhance web application security.
//
// It sets the following headers:
//   - Content-Security-Policy: Restricts sources for default resource loading
//     to the same origin and explicitly allows inline styles, helping prevent
//     XSS attacks.
//   - X-Content-Type-Options: Disables MIME type sniffing and enforces
//     the MIME types specified in Content-Type headers to mitigate MIME type
//     confusion attacks.
//   - X-Frame-Options: Prohibits embedding the content in frames, safeguarding
//     against clickjacking.
//   - X-XSS-Protection: Enables browser-side XSS filters and configures
//     them to block detected XSS attacks.
func AddSecurityHeaders(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Security-Policy",
			"default-src 'self'; style-src 'self' 'unsafe-inline'")
		w.Header().Set("X-Content-Type-Options", "nosniff")
		w.Header().Set("X-Frame-Options", "DENY")
		w.Header().Set("X-XSS-Protection", "1; mode=block")

		next.ServeHTTP(w, r)
	})
}
