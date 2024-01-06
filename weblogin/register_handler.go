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

const (
	MsgMissingRequired    = "Please provide required values"
	MsgUserNameExists     = "User Name already exists."
	MsgEmailExists        = "Email Address already registered."
	MsgPasswordsDifferent = "Password values do not match."
	MsgRegisterFailed     = "Unable to register user."
)

// RegisterPageData contains data passed to the HTML template.
type RegisterPageData struct {
	Title   string
	Message string
}

// renderRegisterPage renders the page.
func (app *LoginApp) renderRegisterPage(w http.ResponseWriter, logger *slog.Logger, message string) {
	data := RegisterPageData{Title: app.Cfg.App.Name, Message: message}

	err := webutil.RenderTemplate(app.Tmpl, w, "register.html", data)
	if err != nil {
		logger.Error("unable to render template", "err", err)
		webutil.HttpError(w, http.StatusInternalServerError)
	}
}

// RegisterHandler handles requests to register a user.
func (app *LoginApp) RegisterHandler(w http.ResponseWriter, r *http.Request) {
	// Get logger with request info and function name.
	logger := webhandler.GetRequestLoggerWithFunc(r)

	// Check if the HTTP method is valid.
	if !webutil.ValidMethod(w, r, http.MethodGet, http.MethodPost) {
		logger.Error("invalid method")
		return
	}

	switch r.Method {
	case http.MethodGet:
		app.renderRegisterPage(w, logger, "")
		logger.Info("success")

	case http.MethodPost:
		app.registerPost(w, r)
	}
}

// registerPost handles POST of the registration form.
func (app *LoginApp) registerPost(w http.ResponseWriter, r *http.Request) {
	// Get logger with request info and function name.
	logger := webhandler.GetRequestLoggerWithFunc(r)

	// Get form values and remove leading and trailing white space.
	userName := strings.TrimSpace(r.PostFormValue("userName"))
	fullName := strings.TrimSpace(r.PostFormValue("fullName"))
	email := strings.TrimSpace(r.PostFormValue("email"))
	password1 := strings.TrimSpace(r.PostFormValue("password1"))
	password2 := strings.TrimSpace(r.PostFormValue("password2"))

	logger = slog.With(
		slog.Group("form",
			"userName", userName,
			"fullName", fullName,
			"email", email,
			// Don't log password values.
			"password1 empty", password1 == "",
			"password2 empty", password2 == "",
		),
	)

	// Check for missing values.
	if IsEmpty(userName, fullName, email, password1, password2) {
		logger.Warn("missing values")
		app.renderRegisterPage(w, logger, MsgMissingRequired)
		return
	}

	// Check that password match.
	if password1 != password2 {
		logger.Warn("passwords do not match")
		app.renderRegisterPage(w, logger, MsgPasswordsDifferent)
		return
	}

	// Check that userName doesn't already exist.
	userExists, err := app.DB.UserExists(userName)
	if err != nil {
		logger.Error("UserExists failed", "err", err)
		webutil.HttpError(w, http.StatusInternalServerError)
		return
	}
	if userExists {
		logger.Warn("user name already exists")
		app.DB.WriteEvent(EventRegister, false, userName, "user name already exists")
		app.renderRegisterPage(w, logger, MsgUserNameExists)
		return
	}

	// Check that email doesn't already exist.
	emailExists, err := app.DB.EmailExists(email)
	if err != nil {
		logger.Error("EmailExists failed")
		webutil.HttpError(w, http.StatusInternalServerError)
		return
	}
	if emailExists {
		logger.Warn("email already exists")
		app.DB.WriteEvent(EventRegister, false, userName, "email already exists: "+email)
		app.renderRegisterPage(w, logger, MsgEmailExists)
		return
	}

	// Register user.
	err = app.DB.RegisterUser(userName, fullName, email, password1)
	if err != nil {
		logger.Error("RegisterUser failed", "err", err)
		app.DB.WriteEvent(EventRegister, false, userName, err.Error())
		app.renderRegisterPage(w, logger, MsgRegisterFailed)
		return
	}

	// Registration successful
	logger.Info("registered user")
	app.DB.WriteEvent(EventRegister, true, userName, "registered user")
	http.Redirect(w, r, "/login", http.StatusSeeOther)
}
