// Copyright 2023 Bill Nixon. All rights reserved.
// Use of this source code is governed by the license found in the LICENSE file.

package webhandler_test

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/bnixon67/webapp/webhandler"
)

func TestLogRequest(t *testing.T) {
	tests := []struct {
		name           string // Name of the test case
		method         string // HTTP method for the request
		url            string // URL for the request
		expectedStatus int    // Expected HTTP status code of the response
	}{
		{
			name:           "GET Request",
			method:         http.MethodGet,
			url:            "/test",
			expectedStatus: http.StatusOK,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create a request with the specified HTTP method and URL.
			req, err := http.NewRequest(tt.method, tt.url, http.NoBody)
			if err != nil {
				t.Fatalf("Could not create request: %v", err)
			}

			// Create a response recorder to record the response.
			rec := httptest.NewRecorder()

			// Create a next handler that just writes a 200 OK status.
			next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(http.StatusOK)
			})

			// Wrap and call next handler with LogRequest middleware.
			middleware := webhandler.LogRequest(next)
			middleware.ServeHTTP(rec, req)

			// Check if the status code is what we expect.
			if status := rec.Code; status != tt.expectedStatus {
				t.Errorf("Handler returned wrong status code: got %v want %v",
					status, tt.expectedStatus)
			}
		})
	}
}
