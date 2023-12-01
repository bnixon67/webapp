// Copyright 2023 Bill Nixon. All rights reserved.
// Use of this source code is governed by the license found in the LICENSE file.

package weblogin

import (
	"errors"
	"log/slog"
	"net/http"

	"github.com/bnixon67/webapp/webhandler"
	"github.com/bnixon67/webapp/webutil"
)

// EventsPageData contains data passed to the HTML template.
type EventsPageData struct {
	Title   string
	Message string
	User    User
	Events  []Event
}

// EventsHandler displays a list of events.
func (app *LoginApp) EventsHandler(w http.ResponseWriter, r *http.Request) {
	// Get logger with request info from request context and add calling function name.
	logger := webhandler.LoggerFromContext(r.Context()).With(slog.String("func", webhandler.FuncName()))

	// Check if the HTTP method is valid.
	if !webutil.ValidMethod(w, r, http.MethodGet) {
		logger.Error("invalid method")
		return
	}

	currentUser, err := app.DB.GetUserFromRequest(w, r)
	if err != nil {
		logger.Error("failed GetUser", "err", err)
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}

	events, err := GetEvents(app.DB)
	if err != nil {
		logger.Error("failed GetEvents", "err", err)
	}

	// display page
	err = webutil.RenderTemplate(app.Tmpl, w, "events.html",
		EventsPageData{
			Title:   app.Cfg.Name,
			Message: "",
			User:    currentUser,
			Events:  events,
		})
	if err != nil {
		logger.Error("failed to RenderTemplate", "err", err)
		return
	}
}

// GetEvents returns a list of all events.
func GetEvents(db *LoginDB) ([]Event, error) {
	var events []Event
	var err error

	if db == nil {
		slog.Error("db is nil")
		return events, errors.New("invalid db")
	}

	qry := `SELECT name, success, userName, message, created FROM events`

	rows, err := db.Query(qry)
	if err != nil {
		slog.Error("query for events failed", "err", err)
		return events, err

	}
	defer rows.Close()

	for rows.Next() {
		var event Event

		err = rows.Scan(&event.Name, &event.Success, &event.UserName, &event.Message, &event.Created)
		if err != nil {
			slog.Error("failed rows.Scan", "err", err)
		}

		events = append(events, event)
	}
	err = rows.Err()
	if err != nil {
		slog.Error("failed rows.Err", "err", err)
	}

	return events, err
}
