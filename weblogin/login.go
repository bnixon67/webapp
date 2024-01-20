package weblogin

import (
	"errors"
	"fmt"
	"log/slog"
)

var ErrAppNil = errors.New("app is nil")

// LoginUser returns a session Token if username and password is correct.
func (app *LoginApp) LoginUser(username, password string) (Token, error) {
	if app == nil {
		return Token{}, ErrAppNil
	}

	err := app.DB.CompareUserPassword(username, password)
	if err != nil {
		app.DB.WriteEvent(EventLogin, false, username, err.Error())

		return Token{}, err
	}

	// create and save a new session token
	token, err := app.DB.SaveNewToken("session", username, 32, app.Cfg.SessionExpires)
	if err != nil {
		app.DB.WriteEvent(EventSaveToken, false, username, err.Error())
		slog.Error("unable to SaveNewToken", "err", err, "username", username)
		return Token{}, fmt.Errorf("unable to save token: %w", err)
	}

	app.DB.WriteEvent(EventLogin, true, username, "user login")

	return token, nil
}
