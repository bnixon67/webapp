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

// HelloPageData contains data passed to the HTML template.
type HelloPageData struct {
	Title   string
	Message string
	User    User
}

// HelloHandler prints a simple hello and any user information.
func (app *LoginApp) HelloHandler(w http.ResponseWriter, r *http.Request) {
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

	// Render the template with the data.
	err = webutil.RenderTemplate(app.Tmpl, w, "hello.html",
		HelloPageData{Message: "", User: user, Title: app.Cfg.Title})
	if err != nil {
		logger.Error("failed to RenderTemplate", "err", err)
		return
	}

	logger.Info("hello", "user", user)
}
