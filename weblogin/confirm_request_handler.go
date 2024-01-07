// Copyright 2024 Bill Nixon. All rights reserved.
// Use of this source code is governed by the license found in the LICENSE file.

package weblogin

import (
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"strings"

	"github.com/bnixon67/webapp/webhandler"
	"github.com/bnixon67/webapp/webutil"
)

// Constants for error and informational messages displayed to the user.
const (
	ConfirmRequestTmpl     = "confirm_request.html"
	ConfirmSentRequestTmpl = "confirm_request_sent.html"
	ConfirmTokenSize       = 12   // Size of the confirm token; TODO: move to config
	ConfirmTokenExpires    = "5m" // Token expiry time; TODO: move to config
)

// ConfirmRequestPageData contains data to render the confirm request templates.
type ConfirmRequestPageData struct {
	Title     string // The application's title.
	Message   string // An message to display to the user.
	EmailFrom string // The email address that sends the confirm message.
}

// renderConfirmRequestPage renders the page.
func (app *LoginApp) renderConfirmRequestPage(w http.ResponseWriter, logger *slog.Logger, data ConfirmRequestPageData) {
	if data.Title == "" {
		data.Title = app.Cfg.App.Name
	}

	err := webutil.RenderTemplate(app.Tmpl, w, "confirm_request.html", data)
	if err != nil {
		logger.Error("unable to render template", "err", err)
		webutil.HttpError(w, http.StatusInternalServerError)
	}
}

// ConfirmRequestHandler handles HTTP request to request a email confirmation.
func (app *LoginApp) ConfirmRequestHandler(w http.ResponseWriter, r *http.Request) {
	// Get logger with request info and function name.
	logger := webhandler.GetRequestLoggerWithFunc(r)

	// Check if the HTTP method is valid.
	if !webutil.ValidMethod(w, r, http.MethodGet, http.MethodPost) {
		logger.Error("invalid method")
		return
	}

	// Route to the appropriate handler based on the HTTP method.
	switch r.Method {
	case http.MethodGet:
		app.confirmRequestGet(w, r)
	case http.MethodPost:
		app.confirmRequestPost(w, r)
	}
}

// confirmRequestGet serves the page to initiate a confirm email request.
func (app *LoginApp) confirmRequestGet(w http.ResponseWriter, r *http.Request) {
	// Get logger with request info and function name.
	logger := webhandler.GetRequestLoggerWithFunc(r)

	data := ConfirmRequestPageData{}
	app.renderConfirmRequestPage(w, logger, data)

	logger.Info("done")
}

// validateConfirmRequestForm ensures all required fields are present and valid.
// It returns an empty string if validate or a message if not.
func validateConfirmRequestForm(email, action string) string {
	if action == "" {
		return MsgMissingAction
	}

	if email == "" {
		return MsgMissingEmail
	}

	if action != "confirm_request" {
		return MsgInvalidAction
	}

	return ""
}

// confirmRequestPost processes a confirm email request.
func (app *LoginApp) confirmRequestPost(w http.ResponseWriter, r *http.Request) {
	// Extract form values.
	email := strings.TrimSpace(r.PostFormValue("email"))
	action := strings.TrimSpace(r.PostFormValue("action"))

	// Get logger with request info and function name and add form values.
	logger := webhandler.GetRequestLoggerWithFunc(r).With(
		slog.String("email", email), slog.String("action", action))

	// Validate form values.
	msg := validateConfirmRequestForm(email, action)
	if msg != "" {
		logger.Warn("invalid form data", "errMessage", msg)
		data := ConfirmRequestPageData{Message: msg}
		app.renderConfirmRequestPage(w, logger, data)
		return
	}

	// Get username for email provided on the form.
	username, err := app.DB.GetUserNameForEmail(email)
	if err != nil && !errors.Is(err, ErrUserNotFound) {
		logger.Error("failed to get username for email",
			"err", err, "email", email)
		webutil.HttpError(w, http.StatusInternalServerError)
		return
	}
	if username == "" {
		logger.Warn("did not find username for email",
			"err", err, "email", email)
		// continue to allow email not registered message
	}

	// Create and save a confirm email token.
	token, err := app.DB.createConfirmEmailToken(username)
	if err != nil {
		slog.Error("failed to create confirm email token",
			"err", err, "username", username)
		webutil.HttpError(w, http.StatusInternalServerError)
		return
	}

	err = sendEmailToConfirm(action, username, email, token, app.Cfg)
	if err != nil {
		logger.Error("unable to send email", "err", err)
		webutil.HttpError(w, http.StatusInternalServerError)
		return
	}

	err = webutil.RenderTemplate(app.Tmpl, w, ConfirmSentRequestTmpl,
		ConfirmRequestPageData{
			Title:     app.Cfg.App.Name,
			EmailFrom: app.Cfg.SMTP.User,
		})
	if err != nil {
		logger.Error("failed to render template", "err", err)
		return
	}

	logger.Info("done")
}

// Template for the emails sent during confirm email process.
const confirmRequestEmailTmpl = `
To confirm your email for {{.Title}}, please visit {{.BaseURL}}/confirm?ctoken={{.Token.Value}} by {{.Token.Expires.Format "January 2, 2006 3:04 PM MST"}}.

You can ignore this message if you did not request to confirm an email for {{.Title}}.
`

// sendEmailToConfirm sends an email to allow user to confirm their email.
func sendEmailToConfirm(action, username, email string, token Token, cfg Config) error {
	subj := fmt.Sprintf("%s confirm email", cfg.App.Name)

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
	case action == "confirm_request":
		body, err = emailBody(
			"password",
			confirmRequestEmailTmpl,
			emailData{
				Token:   token,
				Title:   cfg.App.Name,
				BaseURL: cfg.BaseURL,
			})

		if err != nil {
			return err
		}
	}

	err = SendEmail(cfg.SMTP.User, cfg.SMTP.Password, cfg.SMTP.Host, cfg.SMTP.Port,
		email, subj, body)
	if err != nil {
		return err
	}

	slog.Info("sent email", slog.Group("email",
		slog.String("to", email), slog.String("subject", subj)))

	return err
}

// createConfirmEmailToken generates a new token to confirm a user's email.
func (db *LoginDB) createConfirmEmailToken(username string) (Token, error) {
	// special case for empty username
	if username == "" {
		return Token{}, nil
	}

	return db.SaveNewToken("confirm", username, ConfirmTokenSize, ConfirmTokenExpires)
}
