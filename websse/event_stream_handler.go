// Copyright 2023 Bill Nixon. All rights reserved.
// Use of this source code is governed by the license found in the LICENSE file.

package websse

import (
	"errors"
	"fmt"
	"log/slog"
	"net/http"

	"github.com/bnixon67/webapp/webhandler"
	"github.com/bnixon67/webapp/webutil"
)

// EventStreamHandler handles client connections.
//
// The handler waits for messages and sends them to the client. Multiple
// clients are supported.
//
// The endpoint that accepts an optional "event" query parameter.
// If the "event" query paramater is not provided, a general message
// event is assumed per the SSE standard.
func (s *Server) EventStreamHandler(w http.ResponseWriter, r *http.Request) {
	// Get logger with request info and function name.
	logger := webhandler.RequestLoggerWithFunc(r)

	// Check if the HTTP method is valid.
	if !webutil.ValidMethod(w, r, http.MethodGet) {
		logger.Error("invalid method")
		return
	}

	// Get event from query parameters.
	event := r.URL.Query().Get("event")

	// Only listen for registered events.
	if !s.EventExists(event) {
		slog.Error("event does not exist", "event", event)
		webutil.HttpError(w, http.StatusBadRequest)
		return
	}

	// Add client to listeners for this event.
	id := webhandler.RequestID(r.Context())
	client := s.addClient(id, event)

	// Write necessary HTTP headers for SSE.
	writeHeaders(w)

	// Process messages and handle client disconnects.
	s.process(event, client, w, r, logger)

	logger.Info("client done", "client.id", client.id)
}

// writeHeaders writes the necessary SSE headers.
func writeHeaders(w http.ResponseWriter) {
	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("Connection", "keep-alive")
	w.Header().Set("Access-Control-Allow-Origin", "*") // TODO: optional
}

// process waits for and sends messages and handles client disconnects.
func (s *Server) process(event string, client *Client, w http.ResponseWriter, r *http.Request, logger *slog.Logger) {
	for {
		select {

		case msg, ok := <-client.msgChan: // Message received.
			if !ok {
				logger.Error("message channel closed",
					"client.id", client.id,
					"event", event,
				)
				return
			}

			err := s.writeMessage(w, msg)
			if err != nil {
				slog.Error("failed to write message",
					"err", err,
					"message", msg,
				)
				return
			}

		case <-r.Context().Done(): // Client disconnected.
			logger.Info("client disconnected",
				"client.id", client.id,
				"event", event,
			)
			s.removeClient(event, client)
			return

		}
	}
}

var ErrStreamingNotSupported = errors.New("streaming not supported")

// writeMessage writes a message to the event stream of the client.
func (s *Server) writeMessage(w http.ResponseWriter, msg Message) error {
	// ignore empty Events
	if len(msg.Event) > 0 {
		_, err := fmt.Fprintf(w, "event: %s\n", msg.Event)
		if err != nil {
			return err
		}
	}

	// ignore empty Data
	if len(msg.Data) > 0 {
		_, err := fmt.Fprintf(w, "data: %s\n", msg.Data)
		if err != nil {
			return err
		}
	}

	// ignore empty ID
	if len(msg.ID) > 0 {
		_, err := fmt.Fprintf(w, "id: %s\n", msg.ID)
		if err != nil {
			return err
		}
	}

	if msg.Retry != 0 {
		_, err := fmt.Fprintf(w, "retry: %d\n", msg.Retry)
		if err != nil {
			return err
		}
	}

	// Newline required per standard.
	_, err := fmt.Fprint(w, "\n")
	if err != nil {
		return err
	}

	// Flush to avoid any buffering
	f, ok := w.(http.Flusher)
	if !ok {
		return ErrStreamingNotSupported
	}

	f.Flush()

	return nil
}
