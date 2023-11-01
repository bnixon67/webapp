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
	"database/sql"
	"errors"
	"log/slog"
	"net/http"

	"github.com/bnixon67/webapp/webhandler"
	"github.com/bnixon67/webapp/webutil"
)

// UsersPageData contains data passed to the HTML template.
type UsersPageData struct {
	Title   string
	Message string
	User    User
	Users   []User
}

// UsersHandler prints a simple hello message.
func (app *LoginApp) UsersHandler(w http.ResponseWriter, r *http.Request) {
	// Get logger with request info from request context and add calling function name.
	logger := webhandler.LoggerFromContext(r.Context()).With(slog.String("func", webhandler.FuncName()))

	// Check if the HTTP method is valid.
	if !webutil.ValidMethod(w, r, http.MethodGet) {
		logger.Error("invalid method")
		return
	}

	currentUser, err := GetUserFromRequest(w, r, app.DB)
	if err != nil {
		logger.Error("failed GetUser", "err", err)
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}

	users, err := GetUsers(app.DB)
	if err != nil {
		logger.Error("failed GetUsers", "err", err)
	}

	// display page
	err = webutil.RenderTemplate(app.Tmpl, w, "users.html",
		UsersPageData{Message: "", User: currentUser, Users: users})
	if err != nil {
		logger.Error("failed to RenderTemplate", "err", err)
		return
	}
}

// GetUsers returns a list of all users.
func GetUsers(db *sql.DB) ([]User, error) {
	var users []User
	var err error

	if db == nil {
		slog.Error("db is nil")
		return users, errors.New("invalid db")
	}

	qry := `SELECT userName, fullName, email, admin, created FROM users`

	rows, err := db.Query(qry)
	if err != nil {
		slog.Error("query for users failed", "err", err)
		return users, err

	}
	defer rows.Close()

	for rows.Next() {
		var user User

		err = rows.Scan(&user.UserName, &user.FullName, &user.Email, &user.IsAdmin, &user.Created)
		if err != nil {
			slog.Error("failed rows.Scan", "err", err)
		}

		users = append(users, user)
	}
	err = rows.Err()
	if err != nil {
		slog.Error("failed rows.Err", "err", err)
	}

	return users, err
}
