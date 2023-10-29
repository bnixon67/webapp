// Copyright 2023 Bill Nixon. All rights reserved.
// Use of this source code is governed by the license found in the LICENSE file.

package webserver_test

import (
	"context"
	"errors"
	"net/http"
	"testing"
	"time"

	"github.com/bnixon67/webapp/webserver"
)

func TestWebServer(t *testing.T) {
	tests := []struct {
		name         string
		options      []webserver.Option
		expectedErr  error
		expectedAddr string
	}{
		{
			name:        "Empty Configuration",
			options:     []webserver.Option{},
			expectedErr: nil,
		},
		{
			name: "Default Configuration",
			options: []webserver.Option{
				webserver.WithAddr(":9080"),
			},
			expectedErr: nil,
		},
		{
			name: "Default Configuration With Handler",
			options: []webserver.Option{
				webserver.WithAddr(":9081"),
				webserver.WithHandler(http.DefaultServeMux),
			},
			expectedErr: nil,
		},
		{
			name: "TLS Configuration",
			options: []webserver.Option{
				webserver.WithAddr(":9443"),
				webserver.WithTLS(
					"testdata/cert.pem", "testdata/key.pem"),
			},
			expectedErr: nil,
		},
		{
			name: "Invalid Address",
			options: []webserver.Option{
				webserver.WithAddr("invalid address"),
			},
			expectedErr: webserver.ErrServerStart,
		},
		{
			name: "Invalid TLS Configuration",
			options: []webserver.Option{
				webserver.WithAddr(":9443"),
				webserver.WithTLS(
					"cert.none", "key.none"),
			},
			expectedErr: webserver.ErrServerStart,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			server, err := webserver.New(tt.options...)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			ctx := context.Background()
			ctx, cancel := context.WithTimeout(ctx, 2*time.Second)
			defer cancel()

			errChan := make(chan error)

			go func() {
				errChan <- server.Start(ctx)
			}()

			select {
			case err := <-errChan:
				if !errors.Is(err, context.DeadlineExceeded) && !errors.Is(err, tt.expectedErr) {
					t.Errorf("expected context deadline exceeded error, got %v", err)
				}
			case <-time.After(3 * time.Second):
				t.Error("test timed out")
			}
		})
	}
}

/*
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
		errChan <- server.Start(ctx)
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
*/
