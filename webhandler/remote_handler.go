// Copyright 2023 Bill Nixon. All rights reserved.
// Use of this source code is governed by the license found in the LICENSE file.

package webhandler

import (
	"fmt"
	"net/http"

	"github.com/bnixon67/webapp/webutil"
)

// RemoteGetHandler responds with the requesting client's RemoteAddr and
// potentially real IP addresses from common headers used by proxies or
// load balancers.
//
// This handler ensures that it only responds to HTTP GET requests and
// includes headers to prevent response caching.
func RemoteGetHandler(w http.ResponseWriter, r *http.Request) {
	logger := RequestLoggerWithFuncName(r)

	if !webutil.IsMethodOrError(w, r, http.MethodGet) {
		logger.Error("invalid method")
		return
	}

	webutil.SetContentTypeText(w)
	webutil.SetNoCacheHeaders(w)

	fmt.Fprintf(w, "RemoteAddr: %v\n", r.RemoteAddr)

	// Headers that may have actual IP if behind a proxy or load balancer.
	headers := []string{
		"X-Forwarded-For",
		"X-Real-Ip",
		"X-Client-Ip",
		"Cf-Connecting-Ip",
		"True-Client-Ip",
		"Forwarded",
		"X-ProxyUser-IP",
	}

	// Check and write any relevant headers that contain IP information.
	for _, header := range headers {
		val := r.Header.Get(header)
		if val != "" {
			fmt.Fprintf(w, "%s: %v\n", header, val)
		}
	}

	logger.Info("handler done")
}
