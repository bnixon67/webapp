package weblogin

import (
	"fmt"
	"log/slog"
)

// LoginUser returns a session Token if userName and password is correct.
func (app *LoginApp) LoginUser(userName, password string) (Token, error) {
	err := CompareUserPassword(app.DB, userName, password)
	if err != nil {
		WriteEvent(app.DB,
			Event{Name: EventLogin, Success: false, UserName: userName, Message: err.Error()})

		return Token{}, err
	}

	// create and save a new session token
	token, err := SaveNewToken(app.DB, "session", userName, 32, app.Cfg.SessionExpires)
	if err != nil {
		WriteEvent(app.DB,
			Event{Name: EventSaveToken, Success: false, UserName: userName, Message: err.Error()})
		slog.Error("unable to SaveNewToken", "err", err, "userName", userName)
		return Token{}, fmt.Errorf("unable to save token: %w", err)
	}

	WriteEvent(app.DB, Event{Name: EventLogin, Success: true, UserName: userName, Message: "success"})

	return token, nil
}
