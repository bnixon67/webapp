// Copyright 2023 Bill Nixon. All rights reserved.
// Use of this source code is governed by the license found in the LICENSE file.

package webhandler_test

import (
	"net/http"
	"testing"

	"github.com/bnixon67/webapp/webhandler"
)

func TestGetBuildDateTime(t *testing.T) {
	// Retrieve the executable's modification time.
	dt, err := webhandler.ExecutableModTime()
	if err != nil {
		t.Fatalf("failed to get executable modification time: %v", err)
	}

	// Format the time as a string.
	build := dt.Format(webhandler.BuildDateTimeFormat)

	tests := []TestCase{
		{
			name:          "Valid GET Request",
			requestMethod: http.MethodGet,
			wantStatus:    http.StatusOK,
			wantBody:      build + "\n",
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
	HandlerTestWithCases(t, handler.BuildHandler, tests)
}
