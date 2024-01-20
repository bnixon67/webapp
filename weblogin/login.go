// Copyright 2024 Bill Nixon. All rights reserved.
// Use of this source code is governed by the license found in the LICENSE file.

package weblogin

import (
	"errors"
)

var ErrAppNil = errors.New("app is nil")

const SessionTokenSize = 32
const SessionTokenKind = "session"

// CreateSessionToken creates a session token for username.
func (app *LoginApp) CreateSessionToken(username string) (Token, error) {
	token, err := app.DB.CreateToken(SessionTokenKind, username, SessionTokenSize, app.Cfg.SessionExpires)
	if err != nil {
		return Token{}, err
	}

	return token, nil
}

// LoginUser returns a session token if the username and password are correct.
func (app *LoginApp) LoginUser(username, password string) (Token, error) {
	if app == nil {
		return Token{}, ErrAppNil
	}

	err := app.DB.AuthenticateUser(username, password)
	if err != nil {
		return Token{}, err
	}

	token, err := app.CreateSessionToken(username)
	if err != nil {
		return Token{}, err
	}

	return token, nil
}
