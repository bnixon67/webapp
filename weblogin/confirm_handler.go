// Copyright 2023 Bill Nixon. All rights reserved.
// Use of this source code is governed by the license found in the LICENSE file.

package weblogin

import (
	"log/slog"
	"net/http"
	"strings"

	"github.com/bnixon67/webapp/webhandler"
	"github.com/bnixon67/webapp/webutil"
)

// ConfirmPageData contains data passed to the HTML template.
type ConfirmPageData struct {
	Title        string
	Message      string
	ConfirmToken string
}

// ConfirmHandler handles /reset requests.
func (app *LoginApp) ConfirmHandler(w http.ResponseWriter, r *http.Request) {
	// Get logger with request info from request context and add calling function name.
	logger := webhandler.LoggerFromContext(r.Context()).With(slog.String("func", webhandler.FuncName()))

	// Check if the HTTP method is valid.
	if !webutil.ValidMethod(w, r, http.MethodGet, http.MethodPost) {
		logger.Error("invalid method")
		return
	}

	switch r.Method {
	case http.MethodGet:
		err := webutil.RenderTemplate(app.Tmpl, w, "confirm.html",
			ConfirmPageData{
				Title:        app.Cfg.App.Name,
				ConfirmToken: r.URL.Query().Get("rtoken"),
			})
		if err != nil {
			logger.Error("unable to RenderTemplate", "err", err)
			return
		}
		logger.Info("ConfirmHandler")

	case http.MethodPost:
		app.confirmPost(w, r, "confirm.html")
	}
}

// confirmPost is called for the POST method of the RegisterHandler.
func (app *LoginApp) confirmPost(w http.ResponseWriter, r *http.Request, tmplFileName string) {
	// Get logger with request info from request context and add calling function name.
	logger := webhandler.LoggerFromContext(r.Context()).With(slog.String("func", webhandler.FuncName()))

	// get form values
	token := strings.TrimSpace(r.PostFormValue("ctoken"))

	// check for missing values
	// redundant given client side required fields, but good practice
	if token == "" {
		msg := MsgMissingRequired
		logger.Warn("missing field(s)",
			slog.Group("form",
				"rtoken empty", token == "",
			),
		)
		err := webutil.RenderTemplate(app.Tmpl, w, tmplFileName,
			ResetPageData{
				Title:      app.Cfg.App.Name,
				Message:    msg,
				ResetToken: r.URL.Query().Get("rtoken"),
			})
		if err != nil {
			logger.Error("unable to RenderTemplate", "err", err)
			return
		}
		return
	}

	userName, err := app.DB.GetUserNameForConfirmToken(token)
	if err != nil {
		logger.Error("failed GetUserNameForConfirmToken",
			"token", token,
			"err", err)
		msg := "Please provide a valid token"
		if err == ErrConfirmTokenExpired {
			msg = "Confirm token request expired. Please request again."
		}
		err := webutil.RenderTemplate(app.Tmpl, w, tmplFileName,
			ResetPageData{
				Title:      app.Cfg.App.Name,
				Message:    msg,
				ResetToken: r.URL.Query().Get("rtoken"),
			})
		if err != nil {
			logger.Error("failed to RenderTemplate", "err", err)
			return
		}
		return
	}

	// store the user and hashed password
	_, err = app.DB.Exec("UPDATE users SET confirmed = ? WHERE username = ?", true, userName)
	if err != nil {
		logger.Error("update confirmed failed",
			"userName", userName, "err", err)
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}

	// TODO: don't allow reuse of the reset token if successful

	// register successful
	logger.Info("email confirmed", "userName", userName)
	app.DB.WriteEvent(EventConfirmed, true, userName, "success")
	http.Redirect(w, r, "/login", http.StatusSeeOther)
}
