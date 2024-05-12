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

// ConfirmHandlerGet processes GET requests for the email confirmation page.
//
// This function retrieves the 'ctoken' (confirmation token) from the query
// parameters of the request URL. If a 'ctoken' is present, it is displayed to
// the user on the confirmation page. Users may also manually enter a 'ctoken'.
//
// Submission of the token is via a POST request, which is handled by
// ConfirmHandlerPost, which completes the email confirmation process.
func (app *AuthApp) ConfirmHandlerGet(w http.ResponseWriter, r *http.Request) {
	logger := webhandler.NewRequestLoggerWithFuncName(r)

	if !webutil.IsMethodOrError(w, r, http.MethodGet) {
		logger.Error("invalid method")
		return
	}

	ctoken := r.URL.Query().Get("ctoken")

	data := ConfirmData{ConfirmToken: ctoken}
	app.RenderPage(w, logger, ConfirmTmpl, &data)

	logger.Info("done")
}

const (
	MsgMissingConfirmToken = "Please provide a token."
	MsgInvalidConfirmToken = "Token is invalid. Request a new token."
	MsgExpiredConfirmToken = "Token is expired. Request a new token."
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

// ConfirmHandlerPost processes POST requests for user email confirmation.
//
// It extracts the 'ctoken' (confirmation token) from the form data to verify
// and confirm the associated user. If the token is valid, the user's status
// is updated to confirmed.
func (app *AuthApp) ConfirmHandlerPost(w http.ResponseWriter, r *http.Request) {
	logger := webhandler.NewRequestLoggerWithFuncName(r)

	if !webutil.IsMethodOrError(w, r, http.MethodPost) {
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

	http.Redirect(w, r, "/confirmed", http.StatusSeeOther)
}
