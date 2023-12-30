// Copyright 2023 Bill Nixon. All rights reserved.
// Use of this source code is governed by the license found in the LICENSE file.

package websse

import (
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"time"

	"github.com/bnixon67/webapp/webhandler"
	"github.com/bnixon67/webapp/webutil"
)

// EventStreamHandler handles event stream client connections.
func (s *Server) EventStreamHandler(w http.ResponseWriter, r *http.Request) {
	// Get logger with request info and function name.
	logger := webhandler.GetRequestLoggerWithFunc(r)

	// Check if the HTTP method is valid.
	if !webutil.ValidMethod(w, r, http.MethodGet) {
		logger.Error("invalid method")
		return
	}

	// Get event from query parameters.
	event := r.URL.Query().Get("event")

	// Add to list of receivers for this event.
	id := webhandler.RequestID(r.Context())
	client := s.addClient(id, event)

	// Write necessary HTTP headers for SSE.
	writeHeaders(w)

	// Process messages and handle client disconnects.
	s.processMessages(event, client, w, r, logger)

	logger.Info("client done", "client.id", client.id)
}

// writeHeaders writes the necessary SSE headers.
func writeHeaders(w http.ResponseWriter) {
	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("Connection", "keep-alive")
	w.Header().Set("Access-Control-Allow-Origin", "*") // TODO: optional
}

// processMessages waits for and sends messages and handles client disconnects.
func (s *Server) processMessages(event string, client *Client, w http.ResponseWriter, r *http.Request, logger *slog.Logger) {
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

// writeMessage writes a message to the event stream of the client.
func (s *Server) writeMessage(w http.ResponseWriter, msg Message) error {
	if len(msg.Event) > 0 {
		_, err := fmt.Fprintf(w, "event: %s\n", msg.Event)
		if err != nil {
			return err
		}
	}

	if len(msg.Data) > 0 {
		// TODO: remove timestamp
		_, err := fmt.Fprintf(w, "data: %s %s\n",
			time.Now().Format("15:04:05"), msg.Data)
		if err != nil {
			return err
		}
	}

	if len(msg.ID) > 0 {
		_, err := fmt.Fprintf(w, "id: %s\n", msg.Data)
		if err != nil {
			return err
		}
	}

	// TODO: add Retry

	// Newline required per standard.
	_, err := fmt.Fprint(w, "\n")
	if err != nil {
		return err
	}

	// Flush to avoid any buffering
	f, ok := w.(http.Flusher)
	if !ok {
		return errors.New("streaming not supported")
	}

	f.Flush()

	return nil
}
