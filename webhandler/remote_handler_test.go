// Copyright 2023 Bill Nixon. All rights reserved.
// Use of this source code is governed by the license found in the LICENSE file.

package webhandler_test

import (
	"net/http"
	"testing"

	"github.com/bnixon67/webapp/webhandler"
)

func TestGetRemoteAddr(t *testing.T) {
	tests := []TestCase{
		{
			name:          "Valid GET Request with no headers",
			requestMethod: http.MethodGet,
			wantStatus:    http.StatusOK,
			wantBody:      "RemoteAddr: 192.0.2.1:1234\n",
		},
		{
			name:           "Valid GET Request with headers",
			requestMethod:  http.MethodGet,
			requestHeaders: http.Header{"X-Real-Ip": {"192.0.2.1:5678"}},
			wantStatus:     http.StatusOK,
			wantBody:       "RemoteAddr: 192.0.2.1:1234\nX-Real-Ip: 192.0.2.1:5678\n",
		},
		{
			name:          "Invalid Request Method",
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
	HandlerTestWithCases(t, handler.RemoteHandler, tests)
}
