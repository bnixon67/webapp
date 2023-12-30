// Copyright 2023 Bill Nixon. All rights reserved.
// Use of this source code is governed by the license found in the LICENSE file.

package websse

import (
	"net/http"

	"github.com/bnixon67/webapp/webhandler"
	"github.com/bnixon67/webapp/webutil"
)

// SendMessageHandler will publish any message requests.
func (app *Server) SendMessageHandler(w http.ResponseWriter, r *http.Request) {
	// Get logger with request info and function name.
	logger := webhandler.GetRequestLoggerWithFunc(r)

	// Check if the HTTP method is valid.
	if !webutil.ValidMethod(w, r, http.MethodPost) {
		logger.Error("invalid method")
		return
	}

	event := r.URL.Query().Get("event")
	data := r.URL.Query().Get("data")
	id := r.URL.Query().Get("id")
	// TODO: retry := r.URL.Query().Get("retry")

	// Publish the event.
	app.Publish(Message{Event: event, Data: data, ID: id})
}
