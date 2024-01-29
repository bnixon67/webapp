// Copyright 2024 Bill Nixon. All rights reserved.
// Use of this source code is governed by the license found in the LICENSE file.

package weblogin

import (
	"log/slog"
	"net/http"
	"strings"
	"time"

	"github.com/bnixon67/webapp/webhandler"
	"github.com/bnixon67/webapp/webutil"
)

// LoginPageData contains data passed to the HTML template.
type LoginPageData struct {
	Title   string
	Message string
}

// renderLoginPage renders the login page.
//
// If the page cannot be rendered, http.StatusInternalServerError is
// set and the caller should ensure no further writes are done to w.
func (app *LoginApp) renderLoginPage(w http.ResponseWriter, logger *slog.Logger, data LoginPageData) {
	// Ensure title is set.
	if data.Title == "" {
		data.Title = app.Cfg.App.Name
	}

	err := webutil.RenderTemplate(app.Tmpl, w, "login.html", data)
	if err != nil {
		logger.Error("unable to render template", "err", err)
		return
	}

	return
}

// LoginHandler handles login requests.
func (app *LoginApp) LoginHandler(w http.ResponseWriter, r *http.Request) {
	// Get logger with request info and function name.
	logger := webhandler.RequestLoggerWithFunc(r)

	// Check if the HTTP method is valid.
	if !webutil.ValidMethod(w, r, http.MethodGet, http.MethodPost) {
		logger.Error("invalid method")
		return
	}

	switch r.Method {
	case http.MethodGet:
		app.LoginGetHandler(w, r)
	case http.MethodPost:
		app.LoginPostHandler(w, r)
	}
}

// LoginGetHandler handles login GET requests.
func (app *LoginApp) LoginGetHandler(w http.ResponseWriter, r *http.Request) {
	// Get logger with request info and function name.
	logger := webhandler.RequestLoggerWithFunc(r)

	app.renderLoginPage(w, logger, LoginPageData{})

	logger.Info("done")
}

const (
	MsgMissingUsernameAndPassword = "Missing username and password."
	MsgMissingUsername            = "Missing username."
	MsgMissingPassword            = "Missing password."
	MsgLoginFailed                = "Login failed."
)

// LoginPostHandler handles login POST requests.
func (app *LoginApp) LoginPostHandler(w http.ResponseWriter, r *http.Request) {
	// Get logger with request info and function name.
	logger := webhandler.RequestLoggerWithFunc(r)

	// Get form values.
	username := strings.TrimSpace(r.PostFormValue("username"))
	password := strings.TrimSpace(r.PostFormValue("password"))
	remember := r.PostFormValue("remember")

	logger = slog.With(slog.Group("form",
		slog.String("username", username),
		slog.Bool("password", password == ""),
		slog.String("remember", remember)))

	// Check for missing values.
	var msg string
	switch {
	case username == "" && password == "":
		msg = MsgMissingUsernameAndPassword
	case username == "":
		msg = MsgMissingUsername
	case password == "":
		msg = MsgMissingPassword
	}
	if msg != "" {
		logger.Error("missing form values", slog.String("message", msg))

		app.renderLoginPage(w, logger, LoginPageData{Message: msg})
		return
	}

	// Attempt to login the user.
	token, err := app.LoginUser(username, password)
	if err != nil {
		logger.Error("failed to login user", "err", err)
		app.DB.WriteEvent(EventLogin, false, username, err.Error())

		app.renderLoginPage(w, logger, LoginPageData{Message: MsgLoginFailed})
		return
	}

	// Create cookie for login token.
	app.DB.WriteEvent(EventLogin, true, username, "user logged in")
	var expires time.Time
	if remember == "on" {
		expires = token.Expires
	}
	cookie := LoginCookie(token.Value, expires)
	http.SetCookie(w, cookie)

	// Redirect to the specified "r" query parameter or default to root.
	redirect := r.URL.Query().Get("r")
	if redirect == "" {
		redirect = "/"
	}
	http.Redirect(w, r, redirect, http.StatusSeeOther)

	logger.Info("login successful")
}

func LoginCookie(value string, expires time.Time) *http.Cookie {
	return &http.Cookie{
		Name:     LoginTokenCookieName,
		Value:    value,
		Expires:  expires,
		Secure:   true,
		HttpOnly: true,
		SameSite: http.SameSiteStrictMode,
	}
}
