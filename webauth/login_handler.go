// Copyright 2024 Bill Nixon. All rights reserved.
// Use of this source code is governed by the license found in the LICENSE file.

package webauth

import (
	"errors"
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
	Alert Alert
}

// LoginGetHandler renders the login page.
func (app *AuthApp) LoginGetHandler(w http.ResponseWriter, r *http.Request) {
	logger := webhandler.RequestLoggerWithFuncName(r)

	if !webutil.IsMethodOrError(w, r, http.MethodGet) {
		logger.Error("invalid method")
		return
	}

	// TODO: add a CSRF token and valid on POST.
	app.RenderPage(w, logger, LoginPageName, &LoginPageData{})
}

type loginForm struct {
	Username string
	Password string
	Remember bool
}

// Validation errors.
var (
	ErrMissingUsername = errors.New("missing username")
	ErrMissingPassword = errors.New("missing password")
	ErrMissingBoth     = errors.New("missing username and password")
)

// parseLoginForm extracts and validates the login form fields.
func parseLoginForm(r *http.Request) (loginForm, error) {
	// Extract values.
	username := strings.TrimSpace(r.PostFormValue("username"))
	password := r.PostFormValue("password") // don't trim passwords
	remember := r.PostFormValue("remember") == "on"

	// Check for missing values.
	switch {
	case username == "" && password == "":
		return loginForm{}, ErrMissingBoth
	case username == "":
		return loginForm{}, ErrMissingUsername
	case password == "":
		return loginForm{}, ErrMissingPassword
	}

	return loginForm{
		Username: username,
		Password: password,
		Remember: remember,
	}, nil
}

const (
	MsgMissingBoth     = "Missing username and password."
	MsgMissingUsername = "Missing username."
	MsgMissingPassword = "Missing password."
	MsgLoginFailed     = "Login failed."
)

var validationAlerts = map[error]Alert{
	ErrMissingBoth:     {Type: AlertError, Text: MsgMissingBoth},
	ErrMissingUsername: {Type: AlertError, Text: MsgMissingUsername},
	ErrMissingPassword: {Type: AlertError, Text: MsgMissingPassword},
}

func mapValidationErr(err error) Alert {
	for e, alert := range validationAlerts {
		if errors.Is(err, e) {
			return alert
		}
	}
	return Alert{Type: AlertError, Text: "Invalid input."}
}

// LoginPostHandler processes login.
func (app *AuthApp) LoginPostHandler(w http.ResponseWriter, r *http.Request) {
	logger := webhandler.RequestLoggerWithFuncName(r)

	if !webutil.IsMethodOrError(w, r, http.MethodPost) {
		logger.Error("invalid method")
		return
	}

	form, err := parseLoginForm(r)
	if err != nil {
		logger.Warn("invalid form", "err", err)
		app.RenderPage(w, logger, LoginPageName, &LoginPageData{
			Alert: mapValidationErr(err),
		})
		return
	}

	token, err := app.LoginUser(form.Username, form.Password)
	if err != nil {
		logger.Error("failed to login user", "err", err, "username", form.Username)

		app.RenderPage(w, logger, LoginPageName,
			&LoginPageData{Alert: Alert{Type: AlertError, Text: MsgLoginFailed}})

		return
	}

	http.SetCookie(w, LoginCookie(token.Value, token.Expires, form.Remember))

	redirect := r.URL.Query().Get("r")
	if safe, ok := webutil.ValidateLocalRedirect(redirect); ok {
		redirect = safe
	} else {
		redirect = "/"
	}

	http.Redirect(w, r, redirect, http.StatusSeeOther)
}

// LoginCookie creates the login cookie. Session if remember is false.
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
