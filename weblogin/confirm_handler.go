// Copyright 2024 Bill Nixon. All rights reserved.
// Use of this source code is governed by the license found in the LICENSE file.

package weblogin

import (
	"errors"
	"log/slog"
	"net/http"
	"strings"

	"github.com/bnixon67/webapp/webhandler"
	"github.com/bnixon67/webapp/webutil"
)

// ConfirmPageData contains data to render the confirm template.
type ConfirmPageData struct {
	Title        string
	Message      string
	ConfirmToken string
}

// renderConfirmPage renders the confirm page.
//
// If the page cannot be rendered, http.StatusInternalServerError is
// set and the caller should ensure no further writes are done to w.
func (app *LoginApp) renderConfirmPage(w http.ResponseWriter, logger *slog.Logger, data ConfirmPageData) {
	// Ensure title is set.
	if data.Title == "" {
		data.Title = app.Cfg.App.Name
	}

	err := webutil.RenderTemplate(app.Tmpl, w, "confirm.html", data)
	if err != nil {
		logger.Error("unable to render template", "err", err)
		webutil.HttpError(w, http.StatusInternalServerError)
	}
}

// ConfirmHandler handles request to confirm a user.
func (app *LoginApp) ConfirmHandler(w http.ResponseWriter, r *http.Request) {
	// Get logger with request info and function name.
	logger := webhandler.GetRequestLoggerWithFunc(r)

	// Check if the HTTP method is valid.
	if !webutil.ValidMethod(w, r, http.MethodGet, http.MethodPost) {
		logger.Error("invalid method")
		return
	}

	switch r.Method {
	case http.MethodGet:
		// Get confirm token.
		ctoken := r.URL.Query().Get("ctoken")

		data := ConfirmPageData{ConfirmToken: ctoken}
		app.renderConfirmPage(w, logger, data)

		logger.Info("done")
		return

	case http.MethodPost:
		app.confirmPost(w, r)
		return
	}
}

const (
	MsgMissingConfirmToken  = "Please provide a confirmation token."
	MsgInvalidConfirmToken  = "Please provide a valid confirmation token."
	MsgExpiredConfirmToken  = "The confirmation token has expired. Please request a new one."
	MsgUserAlreadyConfirmed = "The user has already been confirmed." // TODO: is there a security risk in providing this information?
)

// confirmPost is called for the POST method of the RegisterHandler.
func (app *LoginApp) confirmPost(w http.ResponseWriter, r *http.Request) {
	// Get logger with request info and function name.
	logger := webhandler.GetRequestLoggerWithFunc(r)

	// Get confirm token.
	ctoken := strings.TrimSpace(r.PostFormValue("ctoken"))

	// Check for missing values.
	if ctoken == "" {
		logger.Warn("missing ctoken")
		data := ConfirmPageData{Message: MsgMissingConfirmToken}
		app.renderConfirmPage(w, logger, data)
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

		data := ConfirmPageData{Message: msg, ConfirmToken: ctoken}
		app.renderConfirmPage(w, logger, data)
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
			app.renderConfirmPage(w, logger, data)
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
