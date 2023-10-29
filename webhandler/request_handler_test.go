// Copyright 2023 Bill Nixon. All rights reserved.
// Use of this source code is governed by the license found in the LICENSE file.

package webhandler_test

import (
	"net/http"
	"testing"

	"github.com/bnixon67/webapp/webhandler"
)

func TestGetRequestInfo(t *testing.T) {
	tests := []TestCase{
		{
			name:          "Valid GET Request",
			requestMethod: http.MethodGet,
			wantStatus:    http.StatusOK,
			wantBody:      "GET /test HTTP/1.1\r\nHost: example.com\r\n\r\n\n",
		},
		{
			name:          "Valid GET Request with Header",
			requestMethod: http.MethodGet,
			requestHeaders: http.Header{
				"Foo": {"foo1"},
				"bar": {"bar1"},
			},
			wantStatus: http.StatusOK,
			wantBody:   "GET /test HTTP/1.1\r\nHost: example.com\r\nFoo: foo1\r\nbar: bar1\r\n\r\n\n",
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
	HandlerTestWithCases(t, handler.RequestHandler, tests)
}
