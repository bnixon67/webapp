package websse

import (
	"testing"
)

func TestAddAndRemoveClient(t *testing.T) {
	server := NewServer()

	tests := []struct {
		name  string
		id    string
		event string
	}{
		{"AddClient1", "client1", "event1"},
		{"AddClient2", "client2", "event2"},
		{"AddClient3", "client3", "event1"},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			client := server.addClient(tc.id, tc.event)
			if client == nil {
				t.Errorf("addClient returned nil for id %s and event %s", tc.id, tc.event)
			}
			if client.id != tc.id {
				t.Errorf("Expected client ID %s, got %s", tc.id, client.id)
			}

			server.removeClient(tc.event, client)
			for _, c := range server.eventClients[tc.event] {
				if c == client {
					t.Errorf("removeClient did not remove client for event %s", tc.event)
				}
			}
		})
	}
}
