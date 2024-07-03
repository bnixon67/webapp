// Copyright 2024 Bill Nixon. All rights reserved.
// Use of this source code is governed by the license found in the LICENSE file.

package webauth

import (
	"net/http"

	"github.com/bnixon67/webapp/webhandler"
	"github.com/bnixon67/webapp/webutil"
)

const ConfirmRequestSentTmpl = "confirm_request_sent.html"

// ConfirmRequestSentData contains data to render the confirm template.
type ConfirmRequestSentData struct {
	CommonData
	EmailFrom string // The email address that sends the confirm message.
}

// ConfirmRequestSentHandlerGet handles GET requests for confirm request
// sent success page.
func (app *AuthApp) ConfirmRequestSentHandlerGet(w http.ResponseWriter, r *http.Request) {
	logger := webhandler.RequestLoggerWithFuncName(r)

	if !webutil.IsMethodOrError(w, r, http.MethodGet) {
		logger.Error("invalid method")
		return
	}

	data := ConfirmRequestSentData{EmailFrom: app.Cfg.SMTP.Username}
	app.RenderPage(w, logger, ConfirmRequestSentTmpl, &data)

	logger.Info("done")
}
