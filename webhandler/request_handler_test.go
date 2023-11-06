// Copyright 2023 Bill Nixon. All rights reserved.
// Use of this source code is governed by the license found in the LICENSE file.

package webhandler_test

import (
	"net/http"
	"testing"

	"github.com/bnixon67/webapp/webhandler"
)

func TestGetRequestInfo(t *testing.T) {
	tests := []webhandler.TestCase{
		{
			Name:          "Valid GET Request",
			Target:        "/request",
			RequestMethod: http.MethodGet,
			WantStatus:    http.StatusOK,
			WantBody:      "GET /request HTTP/1.1\r\nHost: example.com\r\n\r\n\n",
		},
		{
			Name:          "Valid GET Request with Header",
			Target:        "/request",
			RequestMethod: http.MethodGet,
			RequestHeaders: http.Header{
				"Foo": {"foo1"},
				"bar": {"bar1"},
			},
			WantStatus: http.StatusOK,
			WantBody:   "GET /request HTTP/1.1\r\nHost: example.com\r\nFoo: foo1\r\nbar: bar1\r\n\r\n\n",
		},
		{
			Name:          "Invalid POST Request",
			Target:        "/request",
			RequestMethod: http.MethodPost,
			WantStatus:    http.StatusMethodNotAllowed,
			WantBody:      "POST Method Not Allowed\n",
		},
	}

	// Test the handler using the utility function.
	webhandler.HandlerTestWithCases(t, webhandler.RequestHandler, tests)
}
