package websse

import (
	"fmt"
	"sync"
	"testing"
)

func TestPublish(t *testing.T) {
	server := NewServer()
	server.Run()

	tests := []struct {
		name       string
		message    Message
		numClients int
	}{
		{"PublishEvent0", Message{Event: "event0", Data: "data1"}, 0},
		{"PublishEvent1", Message{Event: "event1", Data: "data1"}, 1},
		{"PublishEvent2", Message{Event: "event2", Data: "data2"}, 2},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			var wg sync.WaitGroup

			// Create multiple clients and add to server
			for n := 0; n < tc.numClients; n++ {

				client := server.addClient(
					fmt.Sprintf("testClient%d", n),
					tc.message.Event,
				)
				wg.Add(1)

				go func(c *Client) {
					defer wg.Done()
					msg := <-client.msgChan
					if msg != tc.message {
						t.Errorf("Expected message %v, got %v", tc.message, msg)
					}
				}(client)
			}

			server.Publish(tc.message)
			wg.Wait()
		})
	}
}
