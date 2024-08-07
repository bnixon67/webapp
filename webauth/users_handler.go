// Copyright 2024 Bill Nixon. All rights reserved.
// Use of this source code is governed by the license found in the LICENSE file.

package webauth

import (
	"errors"
	"log/slog"
	"net/http"

	"github.com/bnixon67/webapp/csv"
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

// UsersHandler shows a list of the current users.
func (app *AuthApp) UsersHandler(w http.ResponseWriter, r *http.Request) {
	logger := webhandler.RequestLoggerWithFuncName(r)

	// Check if the HTTP method is valid.
	if !webutil.CheckAllowedMethods(w, r, http.MethodGet) {
		logger.Error("invalid method")
		return
	}

	currentUser, err := app.DB.UserFromRequest(w, r)
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
	err = webutil.RenderTemplateOrError(app.Tmpl, w, "users.html",
		UsersPageData{
			Title:   app.Cfg.App.Name,
			Message: "",
			User:    currentUser,
			Users:   users,
		})
	if err != nil {
		logger.Error("failed to RenderTemplate", "err", err)
		return
	}

	logger.Info("done")
}

// UsersCSVHandler provides list of the current users as a CSV file.
func (app *AuthApp) UsersCSVHandler(w http.ResponseWriter, r *http.Request) {
	logger := webhandler.RequestLoggerWithFuncName(r)

	// Check if the HTTP method is valid.
	if !webutil.CheckAllowedMethods(w, r, http.MethodGet) {
		logger.Error("invalid method")
		return
	}

	user, err := app.DB.UserFromRequest(w, r)
	if err != nil {
		logger.Error("failed GetUser", "err", err)
		webutil.RespondWithError(w, http.StatusInternalServerError)
		return
	}

	if !user.IsAdmin {
		logger.Error("user not authorized", "user", user)
		webutil.RespondWithError(w, http.StatusUnauthorized)
		return
	}

	users, err := GetUsers(app.DB)
	if err != nil {
		logger.Error("failed GetUsers", "err", err)
		webutil.RespondWithError(w, http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "text/csv")
	w.Header().Set("Content-Disposition", "attachment;filename=users.csv")

	err = csv.SliceOfStructsToCSV(w, users)
	if err != nil {
		logger.Error("failed to convert struct to CSV",
			"err", err, "users", users)
		webutil.RespondWithError(w, http.StatusInternalServerError)
		return
	}
}

// GetUsers returns a list of all users.
func GetUsers(db *AuthDB) ([]User, error) {
	var users []User
	var err error

	if db == nil {
		slog.Error("db is nil")
		return users, errors.New("invalid db")
	}

	qry := `SELECT username, fullName, email, admin, created FROM users`

	rows, err := db.Query(qry)
	if err != nil {
		slog.Error("query for users failed", "err", err)
		return users, err

	}
	defer rows.Close()

	for rows.Next() {
		var user User

		err = rows.Scan(&user.Username, &user.FullName, &user.Email, &user.IsAdmin, &user.Created)
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
