// Copyright 2023 Bill Nixon. All rights reserved.
// Use of this source code is governed by the license found in the LICENSE file.

package webhandler_test

import (
	"net/http"
	"testing"

	"github.com/bnixon67/webapp/assets"
	"github.com/bnixon67/webapp/webhandler"
)

func TestGetHelloTextMessage(t *testing.T) {
	tests := []TestCase{
		{
			name:          "Valid GET Request",
			requestMethod: http.MethodGet,
			wantStatus:    http.StatusOK,
			wantBody:      "hello from Test App\n",
		},
		{
			name:          "Invalid POST Request",
			requestMethod: http.MethodPost,
			wantStatus:    http.StatusMethodNotAllowed,
			wantBody:      "POST Method Not Allowed\n",
		},
	}

	// Create a web handler instance for testing.
	handler, err := webhandler.New(webhandler.WithAppName("Test App"))
	if err != nil {
		t.Fatalf("could not create web handler: %v", err)
	}

	// Test the handler using the utility function.
	HandlerTestWithCases(t, handler.HelloTextHandler, tests)
}

func TestGetHelloHTMLMessage(t *testing.T) {
	tests := []TestCase{
		{
			name:          "Valid GET Request",
			requestMethod: http.MethodGet,
			wantStatus:    http.StatusOK,
			wantBody:      assets.HelloHTML,
		},
		{
			name:          "Invalid POST Request",
			requestMethod: http.MethodPost,
			wantStatus:    http.StatusMethodNotAllowed,
			wantBody:      "POST Method Not Allowed\n",
		},
	}

	// Create a web handler instance for testing.
	handler, err := webhandler.New(webhandler.WithAppName("Test App"))
	if err != nil {
		t.Fatalf("could not create web handler: %v", err)
	}

	// Test the handler using the utility function.
	HandlerTestWithCases(t, handler.HelloHTMLHandler, tests)
}
