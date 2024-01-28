// Copyright 2023 Bill Nixon. All rights reserved.
// Use of this source code is governed by the license found in the LICENSE file.

package weblogin

import (
	"bytes"
	"fmt"
	"log/slog"
	"net/http"
	"slices"
	"strings"
	"text/template"

	"github.com/bnixon67/webapp/webhandler"
	"github.com/bnixon67/webapp/webutil"
)

// Constants for error and informational messages displayed to the user.
const (
	MsgMissingEmail    = "Please provide your email."
	MsgMissingAction   = "Please provide an action."
	MsgInvalidAction   = "Please provide a valid action."
	TemplateForgot     = "forgot.html"
	TemplateForgotSent = "forgot_sent.html"
	ResetTokenSize     = 12   // Size of the password reset token; TODO: move to config
	ResetTokenExpires  = "5m" // Token expiry time; TODO: move to config
)

// ForgotPageData contains data required to render the forgot templates.
type ForgotPageData struct {
	Title     string // The application's title.
	Message   string // An informational or error message to display to the user.
	EmailFrom string // The email address from which the reset or reminder email will be sent.
}

// ForgotHandler handles HTTP requests for forgot user or password.
func (app *LoginApp) ForgotHandler(w http.ResponseWriter, r *http.Request) {
	// Get logger with request info and function name.
	logger := webhandler.RequestLoggerWithFunc(r)

	// Check if the HTTP method is valid.
	if !webutil.ValidMethod(w, r, http.MethodGet, http.MethodPost) {
		logger.Error("invalid method")
		return
	}

	// Route to the appropriate handler based on the HTTP method.
	switch r.Method {
	case http.MethodGet:
		app.forgotGet(w, r)
	case http.MethodPost:
		app.forgotPost(w, r)
	}
}

// forgotGet serves the page to initiate a password reset request.
func (app *LoginApp) forgotGet(w http.ResponseWriter, r *http.Request) {
	// Get logger with request info and function name.
	logger := webhandler.RequestLoggerWithFunc(r)

	err := webutil.RenderTemplate(app.Tmpl, w, TemplateForgot,
		ForgotPageData{Title: app.Cfg.App.Name})
	if err != nil {
		logger.Error("unable to render forgot template", "err", err)
		return
	}

	logger.Info("success")
}

// validateForgotPostForm ensures all required form fields are present and valid.
func validateForgotPostForm(email, action string) string {
	if action == "" {
		return MsgMissingAction
	}

	if email == "" {
		return MsgMissingEmail
	}

	if !isValidAction(action) {
		return MsgInvalidAction
	}

	return ""
}

// isValidAction checks if the action is among the allowed ones.
func isValidAction(action string) bool {
	allowedActions := []string{"user", "password"}
	return slices.Contains(allowedActions, action)
}

// forgotPost processes a forgot user or password request.
func (app *LoginApp) forgotPost(w http.ResponseWriter, r *http.Request) {
	// Extract form values.
	email := strings.TrimSpace(r.PostFormValue("email"))
	action := strings.TrimSpace(r.PostFormValue("action"))

	// Get logger with request info and function name and add form values.
	logger := webhandler.RequestLoggerWithFunc(r).With(
		slog.String("email", email), slog.String("action", action))

	// Validate form values.
	errMessage := validateForgotPostForm(email, action)
	if errMessage != "" {
		logger.Warn("invalid form data", "errMessage", errMessage)
		err := webutil.RenderTemplate(app.Tmpl, w, "forgot.html",
			ForgotPageData{
				Title:   app.Cfg.App.Name,
				Message: errMessage,
			})
		if err != nil {
			logger.Error("unable to RenderTemplate", "err", err)
			return
		}
		return
	}

	// Get username for email provided on the form.
	username, err := app.DB.UsernameForEmail(email)
	if err != nil || username == "" {
		// Don't use logger since log entry doesn't need to contain the request info.
		slog.Warn("failed to get username from email", "err", err, "email", email)
	}

	// create and save a password reset token
	token, err := app.DB.createPasswordResetToken(username)
	if err != nil {
		slog.Error("failed to create password reset token", "err", err, "username", username)
	}

	err = sendEmailForAction(action, username, email, token, app.Cfg)
	if err != nil {
		logger.Error("unable to send email", "err", err)
		http.Error(w,
			http.StatusText(http.StatusInternalServerError),
			http.StatusInternalServerError)
		return
	}

	err = webutil.RenderTemplate(app.Tmpl, w, TemplateForgotSent,
		ForgotPageData{
			Title:     app.Cfg.App.Name,
			EmailFrom: app.Cfg.SMTP.User,
		})
	if err != nil {
		logger.Error("failed to render template", "err", err)
		return
	}

	logger.Info("success")
}

// emailData contain the data required to populate the email templates.
type emailData struct {
	Email    string
	Title    string
	BaseURL  string
	Token    Token
	Username string
}

// Templates for the emails sent during the 'forgot password' or 'forgot username' processes.
const (
	emailNotRegisteredTemplate = `
The email address {{.Email}} is not registered for {{.Title}}.

If you would like to register for {{.Title}}, please visit {{.BaseURL}}/register.
`
	emailForgotPasswordTemplate = `
To reset your password for {{.Title}}, please visit {{.BaseURL}}/reset?rtoken={{.Token.Value}} by {{.Token.Expires.Format "January 2, 2006 3:04 PM MST"}}.

You can ignore this message if you did not request a reset password for {{.Title}}.
`
	emailForgotUserTemplate = `
Your user name for {{.Title}} is {{.Username}}.
`
)

// emailBody constructs the body of an email using a given template and data.
func emailBody(name, text string, data any) (string, error) {
	// Create a template.
	tmpl, err := template.New(name).Parse(text)
	if err != nil {
		return "", err
	}

	// Execute the template with the data and capture the output.
	var buf bytes.Buffer
	if err := tmpl.Execute(&buf, data); err != nil {
		return "", err
	}

	// Return the result as a string.
	return buf.String(), nil
}

// sendEmailForAction sends an email corresponding to a user's reques action.
func sendEmailForAction(action, username, email string, token Token, cfg Config) error {
	subj := fmt.Sprintf("%s forgot %s request", cfg.App.Name, action)

	var body string
	var err error

	switch {
	case username == "":
		body, err = emailBody(
			"notregistered",
			emailNotRegisteredTemplate,
			emailData{
				Email:   email,
				Title:   cfg.App.Name,
				BaseURL: cfg.BaseURL,
			})
	case action == "password":
		body, err = emailBody(
			"password",
			emailForgotPasswordTemplate,
			emailData{
				Token:   token,
				Title:   cfg.App.Name,
				BaseURL: cfg.BaseURL,
			})
	case action == "user":
		body, err = emailBody(
			"user",
			emailForgotUserTemplate,
			emailData{
				Username: username,
				Title:    cfg.App.Name,
				BaseURL:  cfg.BaseURL,
			})
	}

	if err != nil {
		return err
	}

	err = SendEmail(cfg.SMTP,
		MailMessage{To: email, Subject: subj, Body: body})
	if err != nil {
		return err
	}

	slog.Info("sent email", slog.Group("email",
		slog.String("to", email), slog.String("subject", subj)))

	return nil
}

// createPasswordResetToken generates a new token for resetting a user's password.
func (db *LoginDB) createPasswordResetToken(username string) (Token, error) {
	// special case for empty username
	if username == "" {
		return Token{}, nil
	}

	return db.CreateToken("reset", username, ResetTokenSize, ResetTokenExpires)
}
