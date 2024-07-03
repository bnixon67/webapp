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

// LoginPageName is the name of the login HTML template.
const LoginPageName = "login.html"

// LoginPageData contains data passed to the login HTML template.
type LoginPageData struct {
	CommonData
	Message string
}

// LoginGetHandler handles login GET requests.
func (app *AuthApp) LoginGetHandler(w http.ResponseWriter, r *http.Request) {
	logger := webhandler.RequestLoggerWithFuncName(r)

	if !webutil.IsMethodOrError(w, r, http.MethodGet) {
		logger.Error("invalid method")
		return
	}

	app.RenderPage(w, logger, LoginPageName, &LoginPageData{})
}

const (
	MsgMissingUsernameAndPassword = "Missing username and password."
	MsgMissingUsername            = "Missing username."
	MsgMissingPassword            = "Missing password."
	MsgLoginFailed                = "Login failed."
)

type loginForm struct {
	Username string
	Password string
	Remember string
	Message  string
}

// parseLoginForm extracts and validates the login form fields.
// loginForm.Message will contain any errors related to the validation.
func parseLoginForm(r *http.Request) loginForm {
	form := loginForm{
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

// LoginCookie creates and return a login cookie.
func LoginCookie(value string, expires time.Time, remember bool) *http.Cookie {
	if !remember {
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

// LoginPostHandler handles login POST requests.
func (app *AuthApp) LoginPostHandler(w http.ResponseWriter, r *http.Request) {
	logger := webhandler.RequestLoggerWithFuncName(r)

	if !webutil.IsMethodOrError(w, r, http.MethodPost) {
		logger.Error("invalid method")
		return
	}

	form := parseLoginForm(r)

	logger = logger.With(
		slog.Group("form",
			slog.String("username", form.Username),
			slog.Bool("passwordEmpty", form.Password == ""),
			slog.String("remember", form.Remember),
		),
	)

	if form.Message != "" {
		logger.Error("missing form values",
			slog.String("message", form.Message))

		data := LoginPageData{Message: form.Message}
		app.RenderPage(w, logger, LoginPageName, &data)

		return
	}

	token, err := app.LoginUser(form.Username, form.Password)
	if err != nil {
		logger.Error("failed to login user", "err", err)

		data := LoginPageData{Message: MsgLoginFailed}
		app.RenderPage(w, logger, LoginPageName, &data)

		return
	}

	cookie := LoginCookie(token.Value, token.Expires, form.Remember == "on")
	http.SetCookie(w, cookie)

	redirect := r.URL.Query().Get("r")
	if redirect == "" || !webutil.IsLocalSafeURL(redirect) {
		redirect = "/"
	}

	http.Redirect(w, r, redirect, http.StatusSeeOther)
}
