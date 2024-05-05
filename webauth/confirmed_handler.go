// Copyright 2024 Bill Nixon. All rights reserved.
// Use of this source code is governed by the license found in the LICENSE file.

package webauth

import (
	"net/http"

	"github.com/bnixon67/webapp/webhandler"
	"github.com/bnixon67/webapp/webutil"
)

const ConfirmedTmpl = "confirmed.html"

// ConfirmedData contains data to render the confirm template.
type ConfirmedData struct {
	CommonData
}

// ConfirmedHandlerGet handles GET requests for the email confirmation
// success page.
func (app *AuthApp) ConfirmedHandlerGet(w http.ResponseWriter, r *http.Request) {
	logger := webhandler.RequestLoggerWithFuncName(r)

	if !webutil.IsMethodOrError(w, r, http.MethodGet) {
		logger.Error("invalid method")
		return
	}

	data := ConfirmedData{}
	app.RenderPage(w, logger, ConfirmedTmpl, &data)

	logger.Info("done")
}
