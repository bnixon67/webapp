// Copyright 2023 Bill Nixon. All rights reserved.
// Use of this source code is governed by the license found in the LICENSE file.

package webhandler_test

import (
	"net/http"
	"testing"

	"github.com/bnixon67/webapp/webhandler"
)

func TestGetRemoteAddr(t *testing.T) {
	tests := []webhandler.TestCase{
		{
			Name:          "GetWithoutHeaders",
			Target:        "/remote",
			RequestMethod: http.MethodGet,
			WantStatus:    http.StatusOK,
			WantBody:      "RemoteAddr: 192.0.2.1:1234\n",
		},
		{
			Name:           "GetWithHeaders",
			Target:         "/remote",
			RequestMethod:  http.MethodGet,
			RequestHeaders: http.Header{"X-Real-Ip": {"192.0.2.1:5678"}},
			WantStatus:     http.StatusOK,
			WantBody:       "RemoteAddr: 192.0.2.1:1234\nX-Real-Ip: 192.0.2.1:5678\n",
		},
		{
			Name:          "InvalidMethod",
			Target:        "/remote",
			RequestMethod: http.MethodPost,
			WantStatus:    http.StatusMethodNotAllowed,
			WantBody:      "Error: Method Not Allowed\n",
		},
	}

	// Test the handler using the utility function.
	webhandler.HandlerTestWithCases(t, webhandler.RemoteGetHandler, tests)
}
