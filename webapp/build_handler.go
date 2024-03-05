// Copyright 2023 Bill Nixon. All rights reserved.
// Use of this source code is governed by the license found in the LICENSE file.

package webapp

import (
	"fmt"
	"log/slog"
	"net/http"

	"github.com/bnixon67/webapp/webhandler"
	"github.com/bnixon67/webapp/webutil"
)

// BuildDateTimeFormat can be used to format a time as "YYYY-MM-DD HH:MM:SS"
const BuildDateTimeFormat = "2006-01-02 15:04:05"

// BuildHandlerGet responds with the executable modification date and time.
func (app *WebApp) BuildHandlerGet(w http.ResponseWriter, r *http.Request) {
	logger := webhandler.RequestLoggerWithFunc(r)

	if !webutil.EnforceMethod(w, r, http.MethodGet) {
		logger.Error("invalid method")
		return
	}

	webutil.SetNoCacheHeaders(w)

	webutil.SetTextContentType(w)
	build := app.BuildDateTime.Format(BuildDateTimeFormat)
	fmt.Fprintln(w, build)

	logger.Info("done", slog.String("build", build))
}
