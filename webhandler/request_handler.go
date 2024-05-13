// Copyright 2023 Bill Nixon. All rights reserved.
// Use of this source code is governed by the license found in the LICENSE file.

package webhandler

import (
	"fmt"
	"log/slog"
	"net/http"
	"net/http/httputil"

	"github.com/bnixon67/webapp/webutil"
)

// RequestGetHandler serves as an HTTP handler that responds by providing
// a detailed dump of the incoming HTTP request.
func RequestGetHandler(w http.ResponseWriter, r *http.Request) {
	logger := NewRequestLoggerWithFuncName(r)

	if !webutil.IsMethodOrError(w, r, http.MethodGet) {
		logger.Error("invalid method")
		return
	}

	// Attempt to dump the complete request for logging.
	b, err := httputil.DumpRequest(r, true)
	if err != nil {
		errMsg := fmt.Sprintf("error dumping request: %v", err)
		http.Error(w, errMsg, http.StatusInternalServerError)
		logger.Error("failed to dump request",
			slog.String("error", err.Error()))
		return
	}

	webutil.SetContentTypeText(w)
	webutil.SetNoCacheHeaders(w)
	fmt.Fprintln(w, string(b))

	logger.Info("handler done")
}
