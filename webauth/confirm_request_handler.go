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
	CommonData
	Message string // An message to display to the user.
}

// ConfirmRequestHandlerGet processes GET requests for the confirmation request
// page. It allows a user to enter their email to request a confirmation
// token. Submission of the request with an email is via a POST request,
// which is handled by ConfirmRequestHandlerPost.
func (app *AuthApp) ConfirmRequestHandlerGet(w http.ResponseWriter, r *http.Request) {
	logger := webhandler.RequestLoggerWithFuncName(r)

	if !webutil.IsMethodOrError(w, r, http.MethodGet) {
		logger.Error("invalid method")
		return
	}

	data := ConfirmRequestPageData{}
	app.RenderPage(w, logger, "confirm_request.html", &data)

	logger.Info("done")
}

// ConfirmRequestHandlerPost processes POST rquests that emails a user a
// confirmation token. It extracts the 'email" from the form data to create
// the token and then email to the user.
func (app *AuthApp) ConfirmRequestHandlerPost(w http.ResponseWriter, r *http.Request) {
	logger := webhandler.RequestLoggerWithFuncName(r)

	if !webutil.IsMethodOrError(w, r, http.MethodPost) {
		logger.Error("invalid method")
		return
	}

	email := strings.TrimSpace(r.PostFormValue("email"))
	logger = logger.With(slog.String("email", email))

	if email == "" {
		logger.Warn("email is empty")
		data := ConfirmRequestPageData{Message: MsgMissingEmail}
		app.RenderPage(w, logger, "confirm_request.html", &data)
		return
	}

	username, err := app.DB.UsernameForEmail(email)
	if err != nil && !errors.Is(err, ErrUserNotFound) {
		logger.Error("failed to get username for email",
			"err", err, "email", email)
		webutil.RespondWithError(w, http.StatusInternalServerError)
		return
	}
	if username == "" {
		logger.Warn("did not find username for email",
			"err", err, "email", email)
		// Don't return to allow sending an email indicating
		// that the provided email address is not registered.
	}

	token, err := app.DB.CreateConfirmEmailToken(username)
	if err != nil {
		slog.Error("failed to create confirm email token",
			"err", err, "username", username)
		webutil.RespondWithError(w, http.StatusInternalServerError)
		return
	}

	err = sendEmailToConfirm(username, email, token, app.Cfg)
	if err != nil {
		logger.Error("unable to send email", "err", err)
		webutil.RespondWithError(w, http.StatusInternalServerError)
		return
	}

	http.Redirect(w, r, "/confirm_request_sent", http.StatusSeeOther)

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
				BaseURL: cfg.Auth.BaseURL,
			})
	default:
		body, err = emailBody(
			"password",
			confirmRequestEmailTmpl,
			emailData{
				Token:   token,
				Title:   cfg.App.Name,
				BaseURL: cfg.Auth.BaseURL,
			})

		if err != nil {
			return err
		}
	}

	err = cfg.SMTP.SendMessage(cfg.SMTP.User, []string{email}, subj, body)
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
