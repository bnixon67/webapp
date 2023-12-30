// Copyright 2023 Bill Nixon. All rights reserved.
// Use of this source code is governed by the license found in the LICENSE file.

/*
Package main provides an implementation of a Server-Sent Events (SSE) server.

It defines a Server which encapsulates all the logic required to handle
SSE connections, register clients for messages, and broadcast messages
to registered clients. The server supports multiple event types and
ensures thread-safe operations through the use of mutexes.

The package is designed to be integrated into a larger web application,
allowing for real-time broadcasting of messages to web clients via an
event stream. This is facilitated via EventStreamHandler used by clients
to listen to an event stream.

The main components include:
- Client: Represents an event stream client.
- Message: Represents a message in the event stream.
- Server: Manages broadcasting messages to clients.
*/
package websse

import (
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"sync"
	"time"

	"github.com/bnixon67/webapp/webhandler"
	"github.com/bnixon67/webapp/webutil"
)

// Client represents event stream clients.
type Client struct {
	id      string       // id is a unique id for the client.
	msgChan chan Message // msgChan is used to send messages to clients.
}

// Message represents a message in the event stream.
type Message struct {
	Event string // Event identifies the type (name) of event.
	Data  string // Data is the data field for the message.
	ID    string // ID is the event ID in the client.
	Retry int    // Retry is the reconnection time in milliseconds.
}

// Server encapsulates the server-sent event logic.
type Server struct {
	// mu is an mutex to access to eventClients.
	mu sync.Mutex

	// eventClients contains all clients registered for an event.
	eventClients map[string][]*Client

	// broadcast is the channel to send an event.
	broadcast chan Message
}

// writeHeaders writes the necessary SSE headers.
func writeHeaders(w http.ResponseWriter) {
	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("Connection", "keep-alive")
	w.Header().Set("Access-Control-Allow-Origin", "*") // TODO: optional
}

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

// addClient creates a new client and adds it to the event client list.
func (s *Server) addClient(id, event string) *Client {
	client := &Client{
		id:      id,
		msgChan: make(chan Message, 10), // buffered channel
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	s.eventClients[event] = append(s.eventClients[event], client)

	slog.Info("added client", "client.id", client.id, "event", event)

	return client
}

// removeClient removes a client from the event client list.
func (s *Server) removeClient(event string, client *Client) {
	s.mu.Lock()
	defer s.mu.Unlock()

	clients := s.eventClients[event]
	for i, c := range clients { // find the client in the list
		if c == client {
			// avoid memory leak
			clients[i] = nil

			// remove element
			s.eventClients[event] = append(clients[:i],
				clients[i+1:]...)
			break
		}
	}

	slog.Info("removed client", "client.id", client.id, "event", event)
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

// SendHandler will publish any message requests.
func (app *Server) SendHandler(w http.ResponseWriter, r *http.Request) {
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

// listenAndBroadcast listens for messages to broadcast to clients.
// This should be started as a go routine to handle message to the channels.
func (s *Server) listenAndBroadcast() {
	// Receive events and broadcast to clients.
	for event := range s.broadcast {
		s.broadcastToClients(event)
	}
}

// broadcastToClients sends a message to registered clients.
func (s *Server) broadcastToClients(msg Message) {
	s.mu.Lock()
	defer s.mu.Unlock()

	clients := s.eventClients[msg.Event]

	slog.Info("start broadcast", "event", msg, "clients", len(clients))

	for _, client := range clients {
		slog.Info("sending", "client.id", client.id, "message", msg)
		client.msgChan <- msg
	}

	slog.Info("end broadcast", "event", msg, "clients", len(clients))
}

// Publish sends a message to the broadcast channel.
func (s *Server) Publish(msg Message) {
	s.broadcast <- msg
}

// NewServer returns a new server to process server-side events.
func NewServer() *Server {
	s := &Server{
		eventClients: make(map[string][]*Client),
		broadcast:    make(chan Message),
	}

	return s
}

// Run runs the server in a goroutine.
func (s *Server) Run() {
	go s.listenAndBroadcast()
}
