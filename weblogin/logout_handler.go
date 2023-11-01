/*
Copyright 2023 Bill Nixon

Licensed under the Apache License, Version 2.0 (the "License"); you may not use
this file except in compliance with the License.  You may obtain a copy of the
License at http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software distributed
under the License is distributed on an "AS IS" BASIS, WITHOUT WARRANTIES OR
CONDITIONS OF ANY KIND, either express or implied.  See the License for the
specific language governing permissions and limitations under the License.
*/
package weblogin

import (
	"log/slog"
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
	// Get logger with request info from request context and add calling function name.
	logger := webhandler.LoggerFromContext(r.Context()).With(slog.String("func", webhandler.FuncName()))

	// Check if the HTTP method is valid.
	if !webutil.ValidMethod(w, r, http.MethodGet) {
		logger.Error("invalid method")
		return
	}

	user, err := GetUserFromRequest(w, r, app.DB)
	if err != nil {
		logger.Error("failed to GetUser", "err", err)
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}

	// create an empty sessionToken cookie with negative MaxAge to delete
	http.SetCookie(w,
		&http.Cookie{
			Name: SessionTokenCookieName, Value: "", MaxAge: -1,
		})

	// get sessionToken to remove
	sessionTokenValue, err := GetCookieValue(r, SessionTokenCookieName)
	if err != nil {
		logger.Error("failed to GetCookieValue", "err", err)
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}

	// remove session from database
	// TODO: consider removing all sessions for user
	if sessionTokenValue != "" {
		err := RemoveToken(app.DB, "session", sessionTokenValue)
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
		LogoutPageData{Title: app.Cfg.Title})
	if err != nil {
		logger.Error("failed to RenderTemplate", "err", err)
		return
	}

	logger.Info("logged out", "user", user)
	WriteEvent(app.DB, Event{Name: EventLogout, Success: true, UserName: user.UserName, Message: "success"})
}
