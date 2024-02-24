// Copyright 2024 Bill Nixon. All rights reserved.
// Use of this source code is governed by the license found in the LICENSE file.

package webauth

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
	ConfirmTokenSize    = 12   // Size of the confirm token
	ConfirmTokenExpires = "5m" // Token expiry time
	// TODO: move ConfirmTokenSize and ConfirmTokenExpires to config
)

// ConfirmRequestPageData contains data to render the confirm request template.
type ConfirmRequestPageData struct {
	CommonPageData
	Message   string // An message to display to the user.
	EmailFrom string // The email address that sends the confirm message.
}

// ConfirmRequestHandler handles a request to request a email confirmation.
func (app *AuthApp) ConfirmRequestHandler(w http.ResponseWriter, r *http.Request) {
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
		app.confirmRequestGet(w, r)
	case http.MethodPost:
		app.confirmRequestPost(w, r)
	}
}

// confirmRequestGet serves the page to initiate a confirm email request.
func (app *AuthApp) confirmRequestGet(w http.ResponseWriter, r *http.Request) {
	// Get logger with request info and function name.
	logger := webhandler.RequestLoggerWithFunc(r)

	data := ConfirmRequestPageData{}
	app.RenderPage(w, logger, "confirm_request.html", &data)

	logger.Info("done")
}

// confirmRequestPost processes a confirm email request.
func (app *AuthApp) confirmRequestPost(w http.ResponseWriter, r *http.Request) {
	// Extract form values.
	email := strings.TrimSpace(r.PostFormValue("email"))

	// Get logger with request info and function name and add form values.
	logger := webhandler.RequestLoggerWithFunc(r)
	logger = logger.With(
		slog.String("email", email),
	)

	// Validate form values.
	if email == "" {
		logger.Warn("email is empty")
		data := ConfirmRequestPageData{Message: MsgMissingEmail}
		app.RenderPage(w, logger, "confirm_request.html", &data)
		return
	}

	// Get username for email provided on the form.
	username, err := app.DB.UsernameForEmail(email)
	if err != nil && !errors.Is(err, ErrUserNotFound) {
		logger.Error("failed to get username for email",
			"err", err, "email", email)
		webutil.HttpError(w, http.StatusInternalServerError)
		return
	}
	if username == "" {
		logger.Warn("did not find username for email",
			"err", err, "email", email)
		// Don't return to allow sending an email indicating
		// that the provided email address is not registered.
	}

	// Create and save a confirm email token.
	token, err := app.DB.CreateConfirmEmailToken(username)
	if err != nil {
		slog.Error("failed to create confirm email token",
			"err", err, "username", username)
		webutil.HttpError(w, http.StatusInternalServerError)
		return
	}

	err = sendEmailToConfirm(username, email, token, app.Cfg)
	if err != nil {
		logger.Error("unable to send email", "err", err)
		webutil.HttpError(w, http.StatusInternalServerError)
		return
	}

	err = webutil.RenderTemplate(app.Tmpl, w, "confirm_request_sent.html",
		ConfirmRequestPageData{
			CommonPageData: CommonPageData{Title: app.Cfg.App.Name},
			EmailFrom:      app.Cfg.SMTP.User,
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
func sendEmailToConfirm(username, email string, token Token, cfg Config) error {
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
	default:
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

	err = SendEmail(cfg.SMTP,
		MailMessage{To: email, Subject: subj, Body: body})
	if err != nil {
		return err
	}

	slog.Info("sent email", slog.Group("email",
		slog.String("to", email), slog.String("subject", subj)))

	return err
}

// CreateConfirmEmailToken generates a new token to confirm a user's email.
func (db *AuthDB) CreateConfirmEmailToken(username string) (Token, error) {
	// special case for empty username
	if username == "" {
		return Token{}, nil
	}

	return db.CreateToken("confirm", username, ConfirmTokenSize, ConfirmTokenExpires)
}
