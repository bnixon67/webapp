// Copyright 2024 Bill Nixon. All rights reserved.
// Use of this source code is governed by the license found in the LICENSE file.

package webauth

import (
	"log/slog"
	"net/http"
	"strings"

	"github.com/bnixon67/webapp/webhandler"
	"github.com/bnixon67/webapp/webutil"
)

const ConfirmTmpl = "confirm.html"

// ConfirmData contains data to render the confirm template.
type ConfirmData struct {
	CommonData
	Message      string
	ConfirmToken string
}

// ConfirmHandlerGet handles confirm GET requests.
func (app *AuthApp) ConfirmHandlerGet(w http.ResponseWriter, r *http.Request) {
	logger := webhandler.RequestLoggerWithFuncName(r)

	if !webutil.IsMethod(w, r, http.MethodGet) {
		logger.Error("invalid method")
		return
	}

	// Get confirm token.
	ctoken := r.URL.Query().Get("ctoken")

	data := ConfirmData{ConfirmToken: ctoken}
	app.RenderPage(w, logger, ConfirmTmpl, &data)

	logger.Info("done")
}

const (
	MsgMissingConfirmToken = "Please provide a token."
	MsgInvalidConfirmToken = "Token is invalid. Request a new token below."
	MsgExpiredConfirmToken = "Token is expired. Request a new token below."
)

var tokenErrToMsg = map[error]string{
	ErrMissingConfirmToken: MsgMissingConfirmToken,
	ErrTokenNotFound:       MsgInvalidConfirmToken,
	ErrConfirmTokenExpired: MsgExpiredConfirmToken,
}

func (app *AuthApp) respondWithError(w http.ResponseWriter, logger *slog.Logger, err error, ctoken string) {
	logger.Error("failed to confirm user", "err", err)

	msg, ok := tokenErrToMsg[err]
	if !ok {
		webutil.RespondWithError(w, http.StatusInternalServerError)
		return
	}

	app.RenderPage(w, logger, ConfirmTmpl, &ConfirmData{Message: msg})
}

// ConfirmHandlerPost handles confirm POST requests.
func (app *AuthApp) ConfirmHandlerPost(w http.ResponseWriter, r *http.Request) {
	logger := webhandler.RequestLoggerWithFuncName(r)

	if !webutil.IsMethod(w, r, http.MethodPost) {
		logger.Error("invalid method")
		return
	}

	ctoken := strings.TrimSpace(r.PostFormValue("ctoken"))

	username, err := app.DB.UsernameForConfirmToken(ctoken)
	if err != nil {
		app.respondWithError(w, logger, err, ctoken)
		return
	}

	err = app.DB.ConfirmUser(username, ctoken)
	if err != nil {
		app.respondWithError(w, logger, err, ctoken)
		return
	}

	logger.Info("user confirmed", "username", username)
	app.DB.WriteEvent(EventConfirmed, true, username, "success")

	// Redirect to login page.
	// TODO: allow a path for redirect instead of just login.
	// TODO: show a confirmation page before the redirect.
	http.Redirect(w, r, "/login", http.StatusSeeOther)
}
