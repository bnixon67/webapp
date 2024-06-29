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
	CommonData
	Message string
}

// LoginGetHandler handles login GET requests.
func (app *AuthApp) LoginGetHandler(w http.ResponseWriter, r *http.Request) {
	logger := webhandler.NewRequestLoggerWithFuncName(r)

	if !webutil.IsMethodOrError(w, r, http.MethodGet) {
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

type LoginForm struct {
	Username string
	Password string
	Remember string
	Message  string
}

// ParseLoginForm extracts and validates the login form fields.
// Message is updated with any errors related to the validation.
func ParseLoginForm(r *http.Request) LoginForm {
	form := LoginForm{
		Username: strings.TrimSpace(r.PostFormValue("username")),
		Password: strings.TrimSpace(r.PostFormValue("password")),
		Remember: r.PostFormValue("remember"),
	}

	// Check for missing values.
	switch {
	case form.Username == "" && form.Password == "":
		form.Message = MsgMissingUsernameAndPassword
	case form.Username == "":
		form.Message = MsgMissingUsername
	case form.Password == "":
		form.Message = MsgMissingPassword
	}

	return form
}

// LoginPostHandler handles login POST requests.
func (app *AuthApp) LoginPostHandler(w http.ResponseWriter, r *http.Request) {
	logger := webhandler.NewRequestLoggerWithFuncName(r)

	if !webutil.IsMethodOrError(w, r, http.MethodPost) {
		logger.Error("invalid method")
		return
	}

	form := ParseLoginForm(r)

	logger = logger.With(
		slog.Group("form",
			slog.String("username", form.Username),
			slog.Bool("passwordNotEmpty", form.Password != ""),
			slog.String("remember", form.Remember),
		),
	)

	if form.Message != "" {
		logger.Error("missing form values",
			slog.String("message", form.Message))

		data := LoginPageData{Message: form.Message}
		app.RenderPage(w, logger, "login.html", &data)

		return
	}

	// Attempt to login the user.
	token, err := app.LoginUser(form.Username, form.Password)
	if err != nil {
		logger.Error("failed to login user", "err", err)
		app.DB.WriteEvent(EventLogin, false, form.Username, err.Error())

		data := LoginPageData{Message: MsgLoginFailed}
		app.RenderPage(w, logger, "login.html", &data)

		return
	}
	app.DB.WriteEvent(EventLogin, true, form.Username, "logged in user")

	// Create and set login cookie.
	session := form.Remember != "on"
	cookie := LoginCookie(token.Value, token.Expires, session)
	http.SetCookie(w, cookie)

	// Redirect to the specified "r" query parameter or default to root.
	redirect := r.URL.Query().Get("r")
	if redirect == "" || !webutil.IsLocalSafeURL(redirect) {
		redirect = "/"
	}
	http.Redirect(w, r, redirect, http.StatusSeeOther)

	logger.Info("done")
}

// LoginCookie creates and return a login cookie.
func LoginCookie(value string, expires time.Time, session bool) *http.Cookie {
	if session {
		// Set expires to zero time.Time value for session cookie.
		expires = time.Time{}
	}

	return &http.Cookie{
		Name:     LoginTokenCookieName,
		Value:    value,
		Expires:  expires,
		Secure:   true,
		HttpOnly: true,
		SameSite: http.SameSiteStrictMode,
	}
}
