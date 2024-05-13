// Copyright 2024 Bill Nixon. All rights reserved.
// Use of this source code is governed by the license found in the LICENSE file.

package webauth

import (
	"fmt"
	"log/slog"
	"net/http"
	"strings"

	"github.com/bnixon67/webapp/webhandler"
	"github.com/bnixon67/webapp/webutil"
)

const (
	MsgMissingRequired    = "Please provide required values"
	MsgUsernameExists     = "Username already exists."
	MsgEmailExists        = "Email already registered."
	MsgPasswordsDifferent = "Passwords do not match."
	MsgRegisterFailed     = "Unable to register user."
)

// RegisterPageData contains data passed to the HTML template.
type RegisterPageData struct {
	CommonData
	Message string
}

// RegisterHandler handles requests to register a user.
func (app *AuthApp) RegisterHandler(w http.ResponseWriter, r *http.Request) {
	// Get logger with request info and function name.
	logger := webhandler.NewRequestLoggerWithFuncName(r)

	// Check if the HTTP method is valid.
	if !webutil.CheckAllowedMethods(w, r, http.MethodGet, http.MethodPost) {
		logger.Error("invalid method")
		return
	}

	switch r.Method {
	case http.MethodGet:
		app.RenderPage(w, logger, "register.html", &RegisterPageData{})
		logger.Info("done")

	case http.MethodPost:
		app.registerPost(w, r)
	}
}

// registerPost handles POST of the registration form.
func (app *AuthApp) registerPost(w http.ResponseWriter, r *http.Request) {
	// Get logger with request info and function name.
	logger := webhandler.NewRequestLoggerWithFuncName(r)

	// Get form values and remove leading and trailing white space.
	username := strings.TrimSpace(r.PostFormValue("username"))
	fullName := strings.TrimSpace(r.PostFormValue("fullName"))
	email := strings.TrimSpace(r.PostFormValue("email"))
	password1 := strings.TrimSpace(r.PostFormValue("password1"))
	password2 := strings.TrimSpace(r.PostFormValue("password2"))

	logger = logger.With(
		slog.Group("form",
			"username", username,
			"fullName", fullName,
			"email", email,
			// Don't log password values.
			"password1 empty", password1 == "",
			"password2 empty", password2 == "",
		),
	)

	// Check for missing values.
	if IsEmpty(username, fullName, email, password1, password2) {
		logger.Warn("missing values")
		app.RenderPage(w, logger, "register.html",
			&RegisterPageData{Message: MsgMissingRequired})
		return
	}

	// Check that password match.
	if password1 != password2 {
		logger.Warn("passwords do not match")
		app.RenderPage(w, logger, "register.html",
			&RegisterPageData{Message: MsgPasswordsDifferent})
		return
	}

	// Check that username doesn't already exist.
	userExists, err := app.DB.UserExists(username)
	if err != nil {
		logger.Error("UserExists failed", "err", err)
		webutil.RespondWithError(w, http.StatusInternalServerError)
		return
	}
	if userExists {
		logger.Warn("user name already exists")
		app.DB.WriteEvent(EventRegister, false, username, "user name already exists")
		app.RenderPage(w, logger, "register.html",
			&RegisterPageData{Message: MsgUsernameExists})
		return
	}

	// Check that email doesn't already exist.
	emailExists, err := app.DB.EmailExists(email)
	if err != nil {
		logger.Error("EmailExists failed")
		webutil.RespondWithError(w, http.StatusInternalServerError)
		return
	}
	if emailExists {
		logger.Warn("email already exists")
		app.DB.WriteEvent(EventRegister, false, username, "email already exists: "+email)
		app.RenderPage(w, logger, "register.html",
			&RegisterPageData{Message: MsgEmailExists})
		return
	}

	// Register user.
	err = app.DB.RegisterUser(username, fullName, email, password1)
	if err != nil {
		logger.Error("RegisterUser failed", "err", err)
		app.DB.WriteEvent(EventRegister, false, username, err.Error())
		app.RenderPage(w, logger, "register.html",
			&RegisterPageData{Message: MsgRegisterFailed})
		return
	}

	// Registration successful
	logger.Info("registered user")
	app.DB.WriteEvent(EventRegister, true, username, "registered user")

	err = app.sendRegistrationEmail(username, fullName, email)
	if err != nil {
		logger.Error("unable to send registration email", "err", err)
	}

	http.Redirect(w, r, "/login", http.StatusSeeOther)
}

const registrationEmailTmpl = `
{{.FullName}},

Thank you for registering for {{.Title}}. Your username is {{.Username}}.

Please visit {{.BaseURL}}/confirm?ctoken={{.Token.Value}} by {{.Token.Expires.Format "January 2, 2006 3:04 PM MST"}} to confirm your account.

You can ignore this message if you did not register for an account.
`

type registrationData struct {
	FullName string
	Username string
	Title    string
	BaseURL  string
	Token    Token
}

func (app *AuthApp) sendRegistrationEmail(username, fullName, email string) error {
	// Create and save a confirm email token.
	token, err := app.DB.CreateConfirmEmailToken(username)
	if err != nil {
		slog.Error("failed to create confirm email token",
			"err", err, "username", username)
		return err
	}

	subj := fmt.Sprintf("%s registration", app.Cfg.App.Name)

	data := registrationData{
		FullName: fullName,
		Username: username,
		Title:    app.Cfg.App.Name,
		BaseURL:  app.Cfg.Auth.BaseURL,
		Token:    token,
	}
	body, err := emailBody("register", registrationEmailTmpl, data)
	if err != nil {
		return err
	}

	err = app.Cfg.SMTP.SendMessage(app.Cfg.SMTP.Username, []string{email}, subj, body)
	if err != nil {
		return err
	}

	slog.Info("sent email", slog.Group("email",
		slog.String("to", email), slog.String("subject", subj)))

	return nil
}
