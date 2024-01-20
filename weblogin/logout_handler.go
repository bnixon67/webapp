// Copyright 2023 Bill Nixon. All rights reserved.
// Use of this source code is governed by the license found in the LICENSE file.

package weblogin

import (
	"net/http"

	"github.com/bnixon67/webapp/webhandler"
	"github.com/bnixon67/webapp/webutil"
)

// LogoutPageData contains data passed to the HTML template.
type LogoutPageData struct {
	Title   string
	Message string
}

// LogoutHandler handles /logout requests.
func (app *LoginApp) LogoutHandler(w http.ResponseWriter, r *http.Request) {
	// Get logger with request info and function name.
	logger := webhandler.GetRequestLoggerWithFunc(r)

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

	// Create an empty sessionToken cookie with negative MaxAge to delete.
	http.SetCookie(w,
		&http.Cookie{
			Name: SessionTokenCookieName, Value: "", MaxAge: -1,
		})

	// Get sessionToken to remove.
	sessionTokenValue, err := CookieValue(r, SessionTokenCookieName)
	if err != nil {
		logger.Error("failed to GetCookieValue", "err", err)
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}

	// Remove session from database.
	// TODO: consider removing all sessions for user
	if sessionTokenValue != "" {
		err := app.DB.RemoveToken("session", sessionTokenValue)
		if err != nil {
			logger.Error("filed to RemoveToken",
				"sessionTokenValue", sessionTokenValue,
				"err", err)
			// TODO: display error or just continue?
			http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
			return
		}
	}

	// display page
	err = webutil.RenderTemplate(app.Tmpl, w, "logout.html",
		LogoutPageData{Title: app.Cfg.App.Name})
	if err != nil {
		logger.Error("failed to RenderTemplate", "err", err)
		return
	}

	logger.Info("logged out", "user", user)
	app.DB.WriteEvent(EventLogout, true, user.Username, "logged out user")
}
