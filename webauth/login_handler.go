// Copyright 2024 Bill Nixon. All rights reserved.
// Use of this source code is governed by the license found in the LICENSE file.

package webauth

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
	CommonPageData
	Message string
}

// LoginGetHandler handles login GET requests.
func (app *AuthApp) LoginGetHandler(w http.ResponseWriter, r *http.Request) {
	// Get logger with request info and function name.
	logger := webhandler.RequestLoggerWithFunc(r)

	// Check if the HTTP method is valid.
	if r.Method != http.MethodGet {
		webutil.HttpError(w, http.StatusMethodNotAllowed)
		logger.Error("invalid method")
		return
	}

	app.RenderPage(w, logger, "login.html", &LoginPageData{})

	logger.Info("done")
}

const (
	MsgMissingUsernameAndPassword = "Missing username and password."
	MsgMissingUsername            = "Missing username."
	MsgMissingPassword            = "Missing password."
	MsgLoginFailed                = "Login failed."
)

// LoginPostHandler handles login POST requests.
func (app *AuthApp) LoginPostHandler(w http.ResponseWriter, r *http.Request) {
	// Get logger with request info and function name.
	logger := webhandler.RequestLoggerWithFunc(r)

	// Check if the HTTP method is valid.
	if r.Method != http.MethodPost {
		webutil.HttpError(w, http.StatusMethodNotAllowed)
		logger.Error("invalid method")
		return
	}

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

		data := LoginPageData{Message: msg}
		app.RenderPage(w, logger, "login.html", &data)

		return
	}

	// Attempt to login the user.
	token, err := app.LoginUser(username, password)
	if err != nil {
		logger.Error("failed to login user", "err", err)
		app.DB.WriteEvent(EventLogin, false, username, err.Error())

		data := LoginPageData{Message: MsgLoginFailed}
		app.RenderPage(w, logger, "login.html", &data)

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
