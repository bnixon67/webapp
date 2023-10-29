// Copyright 2023 Bill Nixon. All rights reserved.
// Use of this source code is governed by the license found in the LICENSE file.

// Package webhandler provides handlers, middleware, and utilities for web applications.
// It simplifies common tasks, enhances request processing, and includes features like request logging, unique request IDs, and HTML template rendering.
package webhandler

import (
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/google/go-cmp/cmp"
)

type TestCase struct {
	Name           string
	RequestMethod  string
	RequestHeaders http.Header
	WantStatus     int
	WantBody       string
}

// HandlerTestWithCases is a utility function for testing a handler.
func HandlerTestWithCases(t *testing.T, handlerFunc http.HandlerFunc, testCases []TestCase) {
	for _, tt := range testCases {
		t.Run(tt.Name, func(t *testing.T) {
			req := httptest.NewRequest(tt.RequestMethod, "/test", nil)

			req.Header = tt.RequestHeaders

			w := httptest.NewRecorder()

			handlerFunc(w, req)

			resp := w.Result()

			if resp.StatusCode != tt.WantStatus {
				t.Errorf("Want status code %d, got %d", tt.WantStatus, resp.StatusCode)
			}

			body, _ := io.ReadAll(resp.Body)

			diff := cmp.Diff(tt.WantBody, string(body))
			if diff != "" {
				t.Errorf("Body mismatch (-want +got)\n:%s", diff)
			}
		})
	}
}
