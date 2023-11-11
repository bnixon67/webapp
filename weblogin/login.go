package weblogin

import (
	"fmt"
	"log/slog"
)

// LoginUser returns a session Token if userName and password is correct.
func (app *LoginApp) LoginUser(userName, password string) (Token, error) {
	err := app.DB.CompareUserPassword(userName, password)
	if err != nil {
		app.DB.WriteEvent(EventLogin, false, userName, err.Error())

		return Token{}, err
	}

	// create and save a new session token
	token, err := app.DB.SaveNewToken("session", userName, 32, app.Cfg.SessionExpires)
	if err != nil {
		app.DB.WriteEvent(EventSaveToken, false, userName, err.Error())
		slog.Error("unable to SaveNewToken", "err", err, "userName", userName)
		return Token{}, fmt.Errorf("unable to save token: %w", err)
	}

	app.DB.WriteEvent(EventLogin, true, userName, "success")

	return token, nil
}
