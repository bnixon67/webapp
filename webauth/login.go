// Copyright 2024 Bill Nixon. All rights reserved.
// Use of this source code is governed by the license found in the LICENSE file.

package webauth

const LoginTokenSize = 32
const LoginTokenKind = "login"

// CreateLoginToken creates a login token for username.
func (app *AuthApp) CreateLoginToken(username string) (Token, error) {
	token, err := app.DB.CreateToken(LoginTokenKind, username, LoginTokenSize, app.Cfg.Auth.LoginExpires)
	if err != nil {
		return Token{}, err
	}

	return token, nil
}

// LoginUser returns a login token if the username and password are correct.
func (app *AuthApp) LoginUser(username, password string) (Token, error) {
	db := app.DB

	err := db.CheckPassword(username, password)
	if err != nil {
		db.WriteEvent(EventLogin, false, username, err.Error())
		return Token{}, err
	}

	token, err := app.CreateLoginToken(username)
	if err != nil {
		db.WriteEvent(EventLogin, false, username, err.Error())
		return Token{}, err
	}

	db.WriteEvent(EventLogin, true, username, "logged in user")
	return token, nil
}
