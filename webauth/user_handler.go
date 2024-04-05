// Copyright 2024 Bill Nixon. All rights reserved.
// Use of this source code is governed by the license found in the LICENSE file.

package webauth

import (
	"net/http"

	"github.com/bnixon67/webapp/webhandler"
	"github.com/bnixon67/webapp/webutil"
)

// UserPageData contains data passed to the HTML template.
type UserPageData struct {
	Title   string
	Message string
	User    User
}

// UserGetHandler shows user information.
func (app *AuthApp) UserGetHandler(w http.ResponseWriter, r *http.Request) {
	// Get logger with request info and function name.
	logger := webhandler.RequestLoggerWithFunc(r)

	// Check if the HTTP method is valid.
	if r.Method != http.MethodGet {
		webutil.RespondWithError(w, http.StatusMethodNotAllowed)
		logger.Error("invalid method")
		return
	}

	// Attempt to get the user from the request.
	user, err := app.DB.UserFromRequest(w, r)
	if err != nil {
		webutil.RespondWithError(w, http.StatusInternalServerError)
		logger.Error("failed to get user from request", "err", err)
		return
	}

	// Render the template with the data.
	err = webutil.RenderTemplate(app.Tmpl, w, "user.html",
		UserPageData{Message: "", User: user, Title: app.Cfg.App.Name})
	if err != nil {
		logger.Error("failed to render template", "err", err)
		return
	}

	logger.Info("done", "user", user)
}
