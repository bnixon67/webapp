// Copyright 2023 Bill Nixon. All rights reserved.
// Use of this source code is governed by the license found in the LICENSE file.

package weblogin

import (
	"log/slog"
	"net/http"
	"strings"

	"github.com/bnixon67/webapp/webhandler"
	"github.com/bnixon67/webapp/webutil"
	"golang.org/x/crypto/bcrypt"
)

// ResetPageData contains data passed to the HTML template.
type ResetPageData struct {
	Title      string
	Message    string
	ResetToken string
}

// ResetHandler handles /reset requests.
func (app *LoginApp) ResetHandler(w http.ResponseWriter, r *http.Request) {
	// Get logger with request info from request context and add calling function name.
	logger := webhandler.LoggerFromContext(r.Context()).With(slog.String("func", webhandler.FuncName()))

	// Check if the HTTP method is valid.
	if !webutil.ValidMethod(w, r, http.MethodGet, http.MethodPost) {
		logger.Error("invalid method")
		return
	}

	switch r.Method {
	case http.MethodGet:
		err := webutil.RenderTemplate(app.Tmpl, w, "reset.html",
			ResetPageData{
				Title:      app.Cfg.Title,
				ResetToken: r.URL.Query().Get("rtoken"),
			})
		if err != nil {
			logger.Error("unable to RenderTemplate", "err", err)
			return
		}
		logger.Info("ResetHandler")

	case http.MethodPost:
		app.resetPost(w, r, "reset.html")
	}
}

// resetPost is called for the POST method of the RegisterHandler.
func (app *LoginApp) resetPost(w http.ResponseWriter, r *http.Request, tmplFileName string) {
	// Get logger with request info from request context and add calling function name.
	logger := webhandler.LoggerFromContext(r.Context()).With(slog.String("func", webhandler.FuncName()))

	// get form values
	resetToken := strings.TrimSpace(r.PostFormValue("rtoken"))
	password1 := strings.TrimSpace(r.PostFormValue("password1"))
	password2 := strings.TrimSpace(r.PostFormValue("password2"))

	// check for missing values
	// redundant given client side required fields, but good practice
	if resetToken == "" || password1 == "" || password2 == "" {
		msg := MsgMissingRequired
		logger.Warn("missing field(s)",
			slog.Group("form",
				"rtoken empty", resetToken == "",
				"password1 empty", password1 == "",
				"password2 empty", password2 == "",
			),
		)
		err := webutil.RenderTemplate(app.Tmpl, w, tmplFileName,
			ResetPageData{
				Title:      app.Cfg.Title,
				Message:    msg,
				ResetToken: r.URL.Query().Get("rtoken"),
			})
		if err != nil {
			logger.Error("unable to RenderTemplate", "err", err)
			return
		}
		return
	}

	// check that password fields match
	// may be redundant if done client side, but good practice
	if password1 != password2 {
		msg := MsgPasswordsDifferent
		logger.Warn("passwords don't match")
		err := webutil.RenderTemplate(app.Tmpl, w, tmplFileName,
			ResetPageData{
				Title:      app.Cfg.Title,
				Message:    msg,
				ResetToken: r.URL.Query().Get("rtoken"),
			})
		if err != nil {
			logger.Error("unable to RenderTemplate", "err", err)
			return
		}
		return
	}

	userName, err := app.DB.GetUserNameForResetToken(resetToken)
	if err != nil {
		logger.Error("failed GetUserNameForResetToken",
			"resetToken", resetToken,
			"err", err)
		msg := "Please provide a valid Reset Token"
		if err == ErrResetPasswordTokenExpired {
			msg = "Request password request expired. Please request again."
		}
		err := webutil.RenderTemplate(app.Tmpl, w, tmplFileName,
			ResetPageData{
				Title:      app.Cfg.Title,
				Message:    msg,
				ResetToken: r.URL.Query().Get("rtoken"),
			})
		if err != nil {
			logger.Error("failed to RenderTemplate", "err", err)
			return
		}
		return
	}

	// hash the password
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password1), bcrypt.DefaultCost)
	if err != nil {
		msg := "Cannot hash password"
		logger.Error("failed bcrypt.GenerateFromPassword",
			"userName", userName, "err", err)
		err := webutil.RenderTemplate(app.Tmpl, w, tmplFileName,
			ResetPageData{Title: app.Cfg.Title, Message: msg})
		if err != nil {
			logger.Error("unable to RenderTemplate", "err", err)
			return
		}
		return
	}

	// store the user and hashed password
	_, err = app.DB.Exec("UPDATE users SET hashedPassword = ? WHERE username = ?", string(hashedPassword), userName)
	if err != nil {
		logger.Error("update password failed",
			"userName", userName, "err", err)
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}

	// TODO: don't allow reuse of the reset token if successful

	// register successful
	logger.Info("successful password reset", "userName", userName)
	app.DB.WriteEvent(EventResetPass, true, userName, "success")
	http.Redirect(w, r, "/login", http.StatusSeeOther)
}
