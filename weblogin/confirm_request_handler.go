// Copyright 2023 Bill Nixon. All rights reserved.
// Use of this source code is governed by the license found in the LICENSE file.

package weblogin

import (
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

	err := webutil.RenderTemplate(app.Tmpl, w, ConfirmRequestTmpl,
		ConfirmRequestPageData{Title: app.Cfg.App.Name})
	if err != nil {
		logger.Error("unable to render confirm template", "err", err)
		return
	}

	logger.Info("success")
}

// validateConfirmRequestForm ensures all required fields are present and valid.
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
	errMessage := validateConfirmRequestForm(email, action)
	if errMessage != "" {
		logger.Warn("invalid form data", "errMessage", errMessage)
		err := webutil.RenderTemplate(app.Tmpl, w, "confirm.html",
			ConfirmRequestPageData{
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
	username, err := app.DB.GetUserNameForEmail(email)
	if err != nil || username == "" {
		// Don't use logger since log entry doesn't need to contain the request info.
		slog.Warn("failed to get username from email", "err", err, "email", email)
	}

	// create and save a confirm email token
	token, err := app.DB.createConfirmEmailToken(username)
	if err != nil {
		slog.Error("failed to create confirm email token", "err", err, "username", username)
	}

	err = sendEmailToConfirm(action, username, email, token, app.Cfg)
	if err != nil {
		logger.Error("unable to send email", "err", err)
		http.Error(w,
			http.StatusText(http.StatusInternalServerError),
			http.StatusInternalServerError)
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

	logger.Info("success")
}

// Template for the emails sent during confirm email process.
const confirmRequestEmailTmpl = `
To confirm your email for {{.Title}}, please visit {{.BaseURL}}/confirm?rtoken={{.Token.Value}} by {{.Token.Expires.Format "January 2, 2006 3:04 PM MST"}}.

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
