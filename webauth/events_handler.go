// Copyright 2024 Bill Nixon. All rights reserved.
// Use of this source code is governed by the license found in the LICENSE file.

package webauth

import (
	"net/http"

	"github.com/bnixon67/webapp/webhandler"
	"github.com/bnixon67/webapp/webutil"
)

// EventsPageData contains data passed to the HTML template.
type EventsPageData struct {
	CommonData
	User   User
	Events []Event
}

// EventsHandler displays a list of events.
func (app *AuthApp) EventsHandler(w http.ResponseWriter, r *http.Request) {
	// Get logger with request info and function name.
	logger := webhandler.NewRequestLoggerWithFuncName(r)

	// Check if the HTTP method is valid.
	if !webutil.CheckAllowedMethods(w, r, http.MethodGet) {
		logger.Error("invalid method")
		return
	}

	user, err := app.DB.UserFromRequest(w, r)
	if err != nil {
		logger.Error("failed to get user", "err", err)
		webutil.RespondWithError(w, http.StatusInternalServerError)
		return
	}

	events, err := app.DB.GetEvents()
	if err != nil {
		logger.Error("failed to get events", "err", err)
		webutil.RespondWithError(w, http.StatusInternalServerError)
		return
	}

	app.RenderPage(w, logger, "events.html",
		&EventsPageData{
			CommonData: CommonData{Title: app.Cfg.App.Name},
			User:       user,
			Events:     events,
		})

	logger.Info("done")
}

// EventsCSVHandler provides list of events as a CSV file.
func (app *AuthApp) EventsCSVHandler(w http.ResponseWriter, r *http.Request) {
	// Get logger with request info and function name.
	logger := webhandler.NewRequestLoggerWithFuncName(r)

	// Check if the HTTP method is valid.
	if !webutil.CheckAllowedMethods(w, r, http.MethodGet) {
		logger.Error("invalid method")
		return
	}

	user, err := app.DB.UserFromRequest(w, r)
	if err != nil {
		logger.Error("failed to get user", "err", err)
		webutil.RespondWithError(w, http.StatusInternalServerError)
		return
	}

	if !user.IsAdmin {
		logger.Error("user not authorized", "user", user)
		webutil.RespondWithError(w, http.StatusUnauthorized)
		return
	}

	events, err := app.DB.GetEvents()
	if err != nil {
		logger.Error("failed to get events", "err", err)
		webutil.RespondWithError(w, http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "text/csv")
	w.Header().Set("Content-Disposition", "attachment;filename=events.csv")

	err = webutil.SliceOfStructsToCSV(w, events)
	if err != nil {
		logger.Error("failed to convert struct to CSV",
			"err", err, "events", events)
		webutil.RespondWithError(w, http.StatusInternalServerError)
		return
	}
}
