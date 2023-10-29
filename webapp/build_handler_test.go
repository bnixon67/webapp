// Copyright 2023 Bill Nixon. All rights reserved.
// Use of this source code is governed by the license found in the LICENSE file.

package webapp_test

import (
	"net/http"
	"testing"

	"github.com/bnixon67/webapp/webapp"
	"github.com/bnixon67/webapp/webhandler"
)

func TestGetBuildDateTime(t *testing.T) {
	// Retrieve the executable's modification time.
	dt, err := webapp.ExecutableModTime()
	if err != nil {
		t.Fatalf("failed to get executable modification time: %v", err)
	}

	// Format the time as a string.
	build := dt.Format(webapp.BuildDateTimeFormat)

	tests := []webhandler.TestCase{
		{
			Name:          "Valid GET Request",
			RequestMethod: http.MethodGet,
			WantStatus:    http.StatusOK,
			WantBody:      build + "\n",
		},
		{
			Name:          "Invalid Request Method",
			RequestMethod: http.MethodPost,
			WantStatus:    http.StatusMethodNotAllowed,
			WantBody:      "POST Method Not Allowed\n",
		},
	}

	// Create a web app instance for testing.
	app, err := webapp.New(webapp.WithAppName("Test App"))
	if err != nil {
		t.Fatalf("could not create web handler: %v", err)
	}

	// Test the handler using the utility function.
	webhandler.HandlerTestWithCases(t, app.BuildHandler, tests)
}
