// Copyright 2023 Bill Nixon. All rights reserved.
// Use of this source code is governed by the license found in the LICENSE file.

package weblogin

import (
	"net/http"

	"github.com/bnixon67/webapp/webhandler"
	"github.com/bnixon67/webapp/webutil"
)

// HelloPageData contains data passed to the HTML template.
type HelloPageData struct {
	Title   string
	Message string
	User    User
}

// HelloHandler prints a simple hello and any user information.
func (app *LoginApp) HelloHandler(w http.ResponseWriter, r *http.Request) {
	// Get logger with request info and function name.
	logger := webhandler.GetRequestLoggerWithFunc(r)

	// Check if the HTTP method is valid.
	if !webutil.ValidMethod(w, r, http.MethodGet) {
		logger.Error("invalid method")
		return
	}

	// Attempt to get the user from the request.
	user, err := GetUserFromRequest(w, r, app.DB)
	if err != nil {
		logger.Error("failed to get user from request", "err", err)
		http.Error(w,
			http.StatusText(http.StatusInternalServerError),
			http.StatusInternalServerError)
		return
	}

	// Render the template with the data.
	err = webutil.RenderTemplate(app.Tmpl, w, "hello.html",
		HelloPageData{Message: "", User: user, Title: app.Cfg.Title})
	if err != nil {
		logger.Error("failed to render template", "err", err)
		return
	}

	logger.Info("success", "user", user)
}
