// Copyright 2024 Bill Nixon. All rights reserved.
// Use of this source code is governed by the license found in the LICENSE file.

package webapp

import (
	"fmt"
	"log/slog"
	"net/http"

	"github.com/bnixon67/webapp/webhandler"
	"github.com/bnixon67/webapp/webutil"
)

// BuildDateTimeFormat defines the display format for build dates.
const BuildDateTimeFormat = "2006-01-02 15:04:05"

// BuildHandlerGet responds with the application's build date and time.
func (app *WebApp) BuildHandlerGet(w http.ResponseWriter, r *http.Request) {
	logger := webhandler.NewRequestLoggerWithFuncName(r)

	if !webutil.IsMethodOrError(w, r, http.MethodGet) {
		logger.Error("invalid method")
		return
	}

	webutil.SetNoCacheHeaders(w)
	webutil.SetContentTypeText(w)

	buildTime := app.BuildDateTime.Format(BuildDateTimeFormat)
	fmt.Fprintln(w, buildTime)

	logger.Info("done", slog.String("buildTime", buildTime))
}
