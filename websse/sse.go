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
	"log/slog"
	"sync"
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
