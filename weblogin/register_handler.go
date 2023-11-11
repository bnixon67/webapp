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
	MsgMissingRequired    = "Please provide all the required values"
	MsgUserNameExists     = "User Name already exists."
	MsgEmailExists        = "Email Address already registered."
	MsgPasswordsDifferent = "Password values do not match."
)

// RegisterPageData contains data passed to the HTML template.
type RegisterPageData struct {
	Title   string
	Message string
}

// RegisterHandler handles /register requests.
func (app *LoginApp) RegisterHandler(w http.ResponseWriter, r *http.Request) {
	// Get logger with request info from request context and add calling function name.
	logger := webhandler.LoggerFromContext(r.Context()).With(slog.String("func", webhandler.FuncName()))

	// Check if the HTTP method is valid.
	if !webutil.ValidMethod(w, r, http.MethodGet, http.MethodPost) {
		logger.Error("invalid method")
		return
	}

	switch r.Method {
	case http.MethodGet:
		err := webutil.RenderTemplate(app.Tmpl, w, "register.html",
			RegisterPageData{Title: app.Cfg.Title})
		if err != nil {
			logger.Error("unable to parse template", "err", err)
			return
		}
		logger.Info("RegisterHandler")

	case http.MethodPost:
		app.registerPost(w, r)
	}
}

// registerPost is called for the POST method of the RegisterHandler.
func (app *LoginApp) registerPost(w http.ResponseWriter, r *http.Request) {
	// get form values
	userName := strings.TrimSpace(r.PostFormValue("userName"))
	fullName := strings.TrimSpace(r.PostFormValue("fullName"))
	email := strings.TrimSpace(r.PostFormValue("email"))
	password1 := strings.TrimSpace(r.PostFormValue("password1"))
	password2 := strings.TrimSpace(r.PostFormValue("password2"))

	logger := slog.With(
		slog.Group("form",
			"userName", userName,
			"fullName", fullName,
			"email", email,
			"password1 empty", password1 == "",
			"password2 empty", password2 == "",
		),
	)

	// check for missing values
	if IsEmpty(userName, fullName, email, password1, password2) {
		msg := MsgMissingRequired
		logger.Warn("missing values")
		err := webutil.RenderTemplate(app.Tmpl, w, "register.html",
			RegisterPageData{
				Title: app.Cfg.Title, Message: msg,
			})
		if err != nil {
			logger.Error("unable to execute template", "err", err)
			return
		}
		return
	}

	// check that password fields match
	if password1 != password2 {
		msg := MsgPasswordsDifferent
		logger.Warn("passwords do not match")
		err := webutil.RenderTemplate(app.Tmpl, w, "register.html",
			RegisterPageData{
				Title: app.Cfg.Title, Message: msg,
			})
		if err != nil {
			logger.Error("unable to execute template", "err", err)
			return
		}
		return
	}

	// check that userName doesn't already exist
	userExists, err := UserExists(app.DB, userName)
	if err != nil {
		logger.Error("UserExists failed", "err", err)
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}
	if userExists {
		logger.Warn("user already exists")
		WriteEvent(app.DB, Event{Name: EventRegister, Success: false, UserName: userName, Message: "user already exists"})
		err := webutil.RenderTemplate(app.Tmpl, w, "register.html",
			RegisterPageData{
				Title:   app.Cfg.Title,
				Message: MsgUserNameExists,
			})
		if err != nil {
			logger.Error("unable to execute template", "err", err)
			return
		}
		return
	}

	// check that email doesn't already exist
	emailExists, err := EmailExists(app.DB, email)
	if err != nil {
		logger.Error("EmailExists failed")
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}
	if emailExists {
		logger.Warn("email already exists")
		WriteEvent(app.DB, Event{Name: EventRegister, Success: false, UserName: userName, Message: "email already exists"})
		err := webutil.RenderTemplate(app.Tmpl, w, "register.html",
			RegisterPageData{
				Title:   app.Cfg.Title,
				Message: MsgEmailExists,
			})
		if err != nil {
			logger.Error("unable to execute template", "err", err)
			return
		}
		return
	}

	// Register User
	err = RegisterUser(app.DB, userName, fullName, email, password1)
	if err != nil {
		logger.Error("RegisterUser failed", "err", err)
		WriteEvent(app.DB, Event{Name: EventRegister, Success: false, UserName: userName, Message: err.Error()})
		err := webutil.RenderTemplate(app.Tmpl, w, "register.html",
			RegisterPageData{
				Title:   app.Cfg.Title,
				Message: "Unable to Register User",
			})
		if err != nil {
			logger.Error("unable to execute template", "err", err)
			return
		}
		return
	}

	// registration successful
	logger.Info("registered user")
	WriteEvent(app.DB, Event{Name: EventRegister, Success: true, UserName: userName, Message: "success"})
	http.Redirect(w, r, "/login", http.StatusSeeOther)
}
