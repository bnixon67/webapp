// Copyright 2024 Bill Nixon. All rights reserved.
// Use of this source code is governed by the license found in the LICENSE file.

package webapp_test

import (
	"net/http"
	"testing"

	"github.com/bnixon67/webapp/assets"
	"github.com/bnixon67/webapp/webhandler"
)

func TestGetHelloTextMessage(t *testing.T) {
	tests := []webhandler.TestCase{
		{
			Name:          "Valid GET Request",
			RequestMethod: http.MethodGet,
			WantStatus:    http.StatusOK,
			WantBody:      "hello from Test App\n",
		},
		{
			Name:          "Invalid POST Request",
			RequestMethod: http.MethodPost,
			WantStatus:    http.StatusMethodNotAllowed,
			WantBody:      "Error: Method Not Allowed\n",
		},
	}

	app := AppForTest(t)
	webhandler.HandlerTestWithCases(t, app.HelloTextHandlerGet, tests)
}

func TestGetHelloHTMLMessage(t *testing.T) {
	tests := []webhandler.TestCase{
		{
			Name:          "Valid GET Request",
			RequestMethod: http.MethodGet,
			WantStatus:    http.StatusOK,
			WantBody:      assets.HelloHTML,
		},
		{
			Name:          "Invalid POST Request",
			RequestMethod: http.MethodPost,
			WantStatus:    http.StatusMethodNotAllowed,
			WantBody:      "Error: Method Not Allowed\n",
		},
	}

	app := AppForTest(t)
	webhandler.HandlerTestWithCases(t, app.HelloHTMLHandlerGet, tests)
}
