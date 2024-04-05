// Copyright 2023 Bill Nixon. All rights reserved.
// Use of this source code is governed by the license found in the LICENSE file.

package webutil_test

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/bnixon67/webapp/webutil"
)

func TestIsMethodValid(t *testing.T) {
	tests := []struct {
		name          string
		requestMethod string
		validMethod   string
		want          bool
	}{
		{"ValidMethod", "GET", "GET", true},
		{"InvalidMethod", "POST", "GET", false},
		{"ValidMethodCaseSensitive", "get", "GET", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest(tt.requestMethod, "https://example.com", nil)
			w := httptest.NewRecorder()

			got := webutil.IsMethodValid(w, req, tt.validMethod)

			if got != tt.want {
				t.Errorf("IsMethodValid() = %v, want %v", got, tt.want)
			}

			if !tt.want {
				resp := w.Result()
				if resp.StatusCode != http.StatusMethodNotAllowed {
					t.Errorf("Expected status code %d, got %d", http.StatusMethodNotAllowed, resp.StatusCode)
				}
			}
		})
	}
}

func TestCheckAllowedMethods(t *testing.T) {
	tests := []struct {
		name             string
		method           string
		allowedMethods   []string
		expectedResponse string
		expectedStatus   int
		expectValid      bool
	}{
		{
			name:             "GET is Allowed",
			method:           http.MethodGet,
			allowedMethods:   []string{http.MethodGet},
			expectedResponse: "",
			expectedStatus:   0,
			expectValid:      true,
		},
		{
			name:             "POST is Not Allowed",
			method:           http.MethodPost,
			allowedMethods:   []string{http.MethodGet},
			expectedResponse: "POST Method Not Allowed",
			expectedStatus:   http.StatusMethodNotAllowed,
			expectValid:      false,
		},
		{
			name:             "OPTIONS Method",
			method:           http.MethodOptions,
			allowedMethods:   []string{http.MethodGet},
			expectedResponse: "",
			expectedStatus:   http.StatusNoContent,
			expectValid:      false,
		},
		// Add more test cases as needed
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req, err := http.NewRequest(tt.method, "https://example.com/foo", http.NoBody)
			if err != nil {
				t.Fatalf("Could not create request: %v", err)
			}

			rec := httptest.NewRecorder()
			isValid := webutil.CheckAllowedMethods(rec, req, tt.allowedMethods...)

			if isValid != tt.expectValid {
				t.Errorf("Expected valid: %v, got: %v", tt.expectValid, isValid)
			}

			if tt.expectedStatus != 0 && rec.Code != tt.expectedStatus {
				t.Errorf("Expected status %v, got %v", tt.expectedStatus, rec.Code)
			}

			if tt.expectedResponse != "" && rec.Body.String() != tt.expectedResponse+"\n" {
				t.Errorf("Expected response %q, got %q", tt.expectedResponse, rec.Body.String())
			}
		})
	}
}
