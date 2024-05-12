// Copyright 2023 Bill Nixon. All rights reserved.
// Use of this source code is governed by the license found in the LICENSE file.

package webhandler

import (
	"log/slog"
	"net/http"
)

// AddSecurityHeaders is middleware that adds headers to improve security.
func AddSecurityHeaders(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

		// Set security headers

		// Content-Security-Policy: This header helps prevent Cross-Site Scripting (XSS) and data injection attacks. The policy default-src 'self' means that the browser should only load content (scripts, stylesheets, images, etc.) from the same origin as the document.  'self' restricts resource loading to the same origin.  It can be fine-tuned to specify policies for scripts, styles, images, etc., separately. Note that inline styles are no allowed.
		//w.Header().Set("Content-Security-Policy", "default-src 'self'")
		w.Header().Set("Content-Security-Policy", "default-src 'self'; style-src 'self' 'unsafe-inline'")

		// X-Content-Type-Options: This header is used to protect against MIME type confusion attacks. The value nosniff tells the browser not to perform MIME type sniffing, and instead, to strictly follow the declared content type in the HTTP headers. This can prevent maliciously crafted files from being interpreted as a different MIME type.
		w.Header().Set("X-Content-Type-Options", "nosniff")

		// X-Frame-Options: This header controls whether a browser should be allowed to render a page in a <frame>, <iframe>, <embed>, or <object>. Setting it to DENY means the page cannot be displayed in a frame, regardless of the site attempting to do so. This is a defense against clickjacking attacks.
		w.Header().Set("X-Frame-Options", "DENY")

		// X-XSS-Protection: This header is a feature of Internet Explorer, Chrome and Safari that stops pages from loading when they detect reflected Cross-Site Scripting (XSS) attacks. The setting 1; mode=block enables the XSS filter built into most recent web browsers and tells it to block responses that contain detected attacks.
		w.Header().Set("X-XSS-Protection", "1; mode=block")

		logger := NewRequestLogger(r)
		logger.Debug("executed",
			slog.String("func", "AddSecurityHeaders"))

		next.ServeHTTP(w, r)
	})
}
