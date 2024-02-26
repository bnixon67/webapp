// Copyright 2024 Bill Nixon. All rights reserved.
// Use of this source code is governed by the license found in the LICENSE file.

package webauth

import (
	"net/http"

	"github.com/bnixon67/webapp/webhandler"
	"github.com/bnixon67/webapp/webutil"
)

// LogoutPageData contains data passed to the HTML template.
type LogoutPageData struct {
	CommonData
	Message string
}

// LogoutHandler handles /logout requests.
func (app *AuthApp) LogoutHandler(w http.ResponseWriter, r *http.Request) {
	// Get logger with request info and function name.
	logger := webhandler.RequestLoggerWithFunc(r)

	// Check if the HTTP method is valid.
	if !webutil.ValidMethod(w, r, http.MethodGet) {
		logger.Error("invalid method")
		return
	}

	user, err := app.DB.UserFromRequest(w, r)
	if err != nil {
		logger.Error("failed to GetUser", "err", err)
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}

	// Create an empty loginToken cookie with negative MaxAge to delete.
	http.SetCookie(w,
		&http.Cookie{
			Name: LoginTokenCookieName, Value: "", MaxAge: -1,
		})

	// Get loginToken to remove.
	loginTokenValue, err := CookieValue(r, LoginTokenCookieName)
	if err != nil {
		logger.Error("failed to GetCookieValue", "err", err)
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}

	// Remove login token from database.
	// TODO: consider removing all logins for user
	if loginTokenValue != "" {
		err := app.DB.RemoveToken(LoginTokenKind, loginTokenValue)
		if err != nil {
			logger.Error("failed to RemoveToken",
				"loginTokenValue", loginTokenValue,
				"err", err)
			// TODO: display error or just continue?
			http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
			return
		}
	}

	// Render page.
	app.RenderPage(w, logger, "logout.html", &LogoutPageData{})

	logger.Info("logged out", "user", user)
	app.DB.WriteEvent(EventLogout, true, user.Username, "logged out user")
}
