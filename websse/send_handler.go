// Copyright 2023 Bill Nixon. All rights reserved.
// Use of this source code is governed by the license found in the LICENSE file.

package websse

import (
	"fmt"
	"net/http"
	"strconv"
	"time"

	"github.com/bnixon67/webapp/webhandler"
	"github.com/bnixon67/webapp/webutil"
)

// SendMessageHandler publishes any messages posted to handler.
//
// Accepts query parameters "event", "data", "id", and "retry" to create
// the message.
func (s *Server) SendMessageHandler(w http.ResponseWriter, r *http.Request) {
	// Get logger with request info and function name.
	logger := webhandler.RequestLoggerWithFunc(r)

	// Check if the HTTP method is valid.
	if !webutil.ValidMethod(w, r, http.MethodPost) {
		logger.Error("invalid method")
		return
	}

	var msg Message

	// Get query parameters.
	msg.Event = r.URL.Query().Get("event")
	msg.Data = r.URL.Query().Get("data")
	msg.ID = r.URL.Query().Get("id")
	retryStr := r.URL.Query().Get("retry")

	// Convert retry string to int
	if len(retryStr) > 0 {
		retry, err := strconv.Atoi(retryStr)
		if err != nil {
			logger.Error("unable to convert string to int",
				"retry", retryStr,
				"error", err,
			)
			webutil.HttpError(w, http.StatusUnprocessableEntity)
			return
		}

		msg.Retry = retry
	}

	// Publish the message with data prefixed with a timestamp.
	msg.Data = fmt.Sprintf("%s %s", time.Now().Format("15:04:05"), msg.Data)
	err := s.Publish(msg)
	if err != nil {
		logger.Error("unable to publish message",
			"err", err,
			"message", msg,
		)
		webutil.HttpError(w, http.StatusUnprocessableEntity)
		return
	}

	return
}
