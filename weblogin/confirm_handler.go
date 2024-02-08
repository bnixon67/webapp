// Copyright 2024 Bill Nixon. All rights reserved.
// Use of this source code is governed by the license found in the LICENSE file.

package weblogin

import (
	"errors"
	"html/template"
	"net/http"
	"strings"

	"github.com/bnixon67/webapp/webhandler"
	"github.com/bnixon67/webapp/webutil"
)

// ConfirmPageData contains data to render the confirm template.
type ConfirmPageData struct {
	CommonPageData
	Message      template.HTML
	ConfirmToken string
}

// ConfirmHandler handles request to confirm a user.
func (app *LoginApp) ConfirmHandler(w http.ResponseWriter, r *http.Request) {
	// Get logger with request info and function name.
	logger := webhandler.RequestLoggerWithFunc(r)

	// Check if the HTTP method is valid.
	if !webutil.ValidMethod(w, r, http.MethodGet, http.MethodPost) {
		logger.Error("invalid method")
		return
	}

	switch r.Method {
	case http.MethodGet:
		app.ConfirmGetHandler(w, r)
	case http.MethodPost:
		app.ConfirmPostHandler(w, r)
		return
	}
}

func (app *LoginApp) ConfirmGetHandler(w http.ResponseWriter, r *http.Request) {
	// Get logger with request info and function name.
	logger := webhandler.RequestLoggerWithFunc(r)

	// Get confirm token.
	ctoken := r.URL.Query().Get("ctoken")

	data := ConfirmPageData{ConfirmToken: ctoken}
	app.RenderPage(w, logger, "confirm.html", &data)

	logger.Info("done")
	return
}

const (
	MsgMissingConfirmToken  = "Please provide a confirmation token."
	MsgInvalidConfirmToken  = "Invalid confirmation token. Please <a href=\"/confirm_request\">request</a> a new one."
	MsgExpiredConfirmToken  = "The confirmation token has expired. Please <a href=\"/confirm_request\">request</a> a new one."
	MsgUserAlreadyConfirmed = "The user has already been confirmed." // TODO: is there a security risk in providing this information?
)

// ConfirmPostHandler is called for the POST method of the RegisterHandler.
func (app *LoginApp) ConfirmPostHandler(w http.ResponseWriter, r *http.Request) {
	// Get logger with request info and function name.
	logger := webhandler.RequestLoggerWithFunc(r)

	// Get confirm token.
	ctoken := strings.TrimSpace(r.PostFormValue("ctoken"))

	// Check for missing values.
	if ctoken == "" {
		logger.Warn("missing ctoken")
		data := ConfirmPageData{
			Message: template.HTML(MsgMissingConfirmToken),
		}
		app.RenderPage(w, logger, "confirm.html", &data)
		return
	}

	// Get username for the confirm token.
	username, err := app.DB.UsernameForConfirmToken(ctoken)
	if err != nil {
		logger.Error("failed to get username for confirm token",
			"ctoken", ctoken,
			"err", err)

		msg := MsgInvalidConfirmToken
		if err == ErrConfirmTokenExpired {
			msg = MsgExpiredConfirmToken
		}

		data := ConfirmPageData{
			Message:      template.HTML(msg),
			ConfirmToken: ctoken,
		}
		app.RenderPage(w, logger, "confirm.html", &data)
		return
	}

	// Confirm the user.
	err = app.DB.ConfirmUser(username)
	if err != nil {
		logger.Error("failed to confirm user",
			"username", username, "err", err)

		// Special case if user already confirmed.
		if errors.Is(err, ErrUserAlreadyConfirmed) {
			data := ConfirmPageData{
				Message: MsgUserAlreadyConfirmed,
			}
			app.RenderPage(w, logger, "confirm.html", &data)
			return
		}

		webutil.HttpError(w, http.StatusInternalServerError)
		return
	}

	// Remove token to prevent reuse.
	err = app.DB.RemoveToken("confirm", ctoken)
	if err != nil {
		logger.Error("failed to remove token",
			"ctoken", ctoken, "err", err)
	}

	// Confirmation was successful.
	logger.Info("user confirmed", "username", username)
	app.DB.WriteEvent(EventConfirmed, true, username, "success")

	// Redirect to login page.
	// TODO: allow a path for redirect instead of just login.
	http.Redirect(w, r, "/login", http.StatusSeeOther)
}
