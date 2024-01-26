// Copyright 2023 Bill Nixon. All rights reserved.
// Use of this source code is governed by the license found in the LICENSE file.

package weblogin

import (
	"log/slog"
	"net/http"

	"github.com/bnixon67/webapp/webhandler"
	"github.com/bnixon67/webapp/webutil"
)

// EventsPageData contains data passed to the HTML template.
type EventsPageData struct {
	Title  string
	User   User
	Events []Event
}

// renderEventsPage renders the events page.  If the page cannot be
// rendered, http.StatusInternalServerError is set and the caller should
// ensure no further writes are done to w.
func (app *LoginApp) renderEventsPage(w http.ResponseWriter, logger *slog.Logger, data EventsPageData) {
	// Ensure title is set.
	if data.Title == "" {
		data.Title = app.Cfg.App.Name
	}

	err := webutil.RenderTemplate(app.Tmpl, w, "events.html", data)
	if err != nil {
		logger.Error("unable to render template", "err", err)
		webutil.HttpError(w, http.StatusInternalServerError)
	}
}

// EventsHandler displays a list of events.
func (app *LoginApp) EventsHandler(w http.ResponseWriter, r *http.Request) {
	// Get logger with request info and function name.
	logger := webhandler.RequestLoggerWithFunc(r)

	// Check if the HTTP method is valid.
	if !webutil.ValidMethod(w, r, http.MethodGet) {
		logger.Error("invalid method")
		return
	}

	user, err := app.DB.UserFromRequest(w, r)
	if err != nil {
		logger.Error("failed to get user", "err", err)
		webutil.HttpError(w, http.StatusInternalServerError)
		return
	}

	events, err := app.DB.GetEvents()
	if err != nil {
		logger.Error("failed to get events", "err", err)
		webutil.HttpError(w, http.StatusInternalServerError)
		return
	}

	app.renderEventsPage(w, logger,
		EventsPageData{
			Title:  app.Cfg.App.Name,
			User:   user,
			Events: events,
		})

	logger.Info("done")
}

// EventsCSVHandler provides list of events as a CSV file.
func (app *LoginApp) EventsCSVHandler(w http.ResponseWriter, r *http.Request) {
	// Get logger with request info and function name.
	logger := webhandler.RequestLoggerWithFunc(r)

	// Check if the HTTP method is valid.
	if !webutil.ValidMethod(w, r, http.MethodGet) {
		logger.Error("invalid method")
		return
	}

	user, err := app.DB.UserFromRequest(w, r)
	if err != nil {
		logger.Error("failed to get user", "err", err)
		webutil.HttpError(w, http.StatusInternalServerError)
		return
	}

	if !user.IsAdmin {
		logger.Error("user not authorized", "user", user)
		webutil.HttpError(w, http.StatusUnauthorized)
		return
	}

	events, err := app.DB.GetEvents()
	if err != nil {
		logger.Error("failed to get events", "err", err)
		webutil.HttpError(w, http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "text/csv")
	w.Header().Set("Content-Disposition", "attachment;filename=events.csv")

	err = webutil.SliceOfStructsToCSV(w, events)
	if err != nil {
		logger.Error("failed to convert struct to CSV",
			"err", err, "events", events)
		webutil.HttpError(w, http.StatusInternalServerError)
		return
	}
}
