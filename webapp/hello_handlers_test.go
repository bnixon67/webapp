// Copyright 2023 Bill Nixon. All rights reserved.
// Use of this source code is governed by the license found in the LICENSE file.

package webapp_test

import (
	"net/http"
	"testing"

	"github.com/bnixon67/webapp/assets"
	"github.com/bnixon67/webapp/webapp"
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
			WantBody:      "POST Method Not Allowed\n",
		},
	}

	// Create a web app instance for testing.
	app, err := webapp.New(webapp.WithName("Test App"))
	if err != nil {
		t.Fatalf("could not create web handler: %v", err)
	}

	// Test the handler using the utility function.
	webhandler.HandlerTestWithCases(t, app.HelloTextHandler, tests)
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
			WantBody:      "POST Method Not Allowed\n",
		},
	}

	// Create a web app instance for testing.
	app, err := webapp.New(webapp.WithName("Test App"))
	if err != nil {
		t.Fatalf("could not create web handler: %v", err)
	}

	// Test the handler using the utility function.
	webhandler.HandlerTestWithCases(t, app.HelloHTMLHandler, tests)
}
