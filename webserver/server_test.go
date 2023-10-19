package webserver_test

import (
	"context"
	"errors"
	"net/http"
	"testing"
	"time"

	"github.com/bnixon67/webapp/webserver"
)

const addr = ":9080"

func TestNew(t *testing.T) {
	tests := []struct {
		name        string
		options     []webserver.Option
		expectedErr error
	}{
		{
			name: "With Address and Handler",
			options: []webserver.Option{
				webserver.WithAddr(addr),
				webserver.WithHandler(http.DefaultServeMux),
			},
			expectedErr: nil,
		},
		// More test cases can be added as needed
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := webserver.New(tt.options...)
			if err != tt.expectedErr {
				t.Errorf("expected error %v, got %v", tt.expectedErr, err)
			}
		})
	}
}

func TestRun(t *testing.T) {
	server, err := webserver.New(
		webserver.WithAddr(addr),
		webserver.WithHandler(http.DefaultServeMux),
	)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	errChan := make(chan error)

	go func() {
		errChan <- webserver.Run(ctx, server)
	}()

	select {
	case err := <-errChan:
		if !errors.Is(err, context.DeadlineExceeded) {
			t.Errorf("expected context deadline exceeded error, got %v", err)
		}
	case <-time.After(3 * time.Second):
		t.Error("test timed out")
	}
}
