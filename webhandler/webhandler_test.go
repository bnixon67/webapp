// Copyright 2023 Bill Nixon. All rights reserved.
// Use of this source code is governed by the license found in the LICENSE file.

package webhandler_test

import (
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/bnixon67/webapp/webhandler"
	"github.com/google/go-cmp/cmp"
)

func TestNew(t *testing.T) {
	tests := []struct {
		name     string
		opts     []webhandler.Option
		wantName string
		wantErr  bool
	}{
		{
			name:     "With AppName",
			opts:     []webhandler.Option{webhandler.WithAppName("TestApp")},
			wantName: "TestApp",
			wantErr:  false,
		},
		{
			name:     "Without AppName",
			opts:     []webhandler.Option{},
			wantName: "",
			wantErr:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			h, err := webhandler.New(tt.opts...)
			if (err != nil) != tt.wantErr {
				t.Errorf("New() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr && h.AppName != tt.wantName {
				t.Errorf("New() AppName = %v, want %v", h.AppName, tt.wantName)
			}
		})
	}
}

type TestCase struct {
	name           string
	requestMethod  string
	requestHeaders http.Header
	wantStatus     int
	wantBody       string
}

// HandlerTestWithCases is a utility function for testing a handler.
func HandlerTestWithCases(t *testing.T, handlerFunc http.HandlerFunc, testCases []TestCase) {
	for _, tt := range testCases {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest(tt.requestMethod, "/test", nil)

			req.Header = tt.requestHeaders

			w := httptest.NewRecorder()

			handlerFunc(w, req)

			resp := w.Result()

			if resp.StatusCode != tt.wantStatus {
				t.Errorf("Want status code %d, got %d", tt.wantStatus, resp.StatusCode)
			}

			body, _ := io.ReadAll(resp.Body)

			diff := cmp.Diff(tt.wantBody, string(body))
			if diff != "" {
				t.Errorf("Body mismatch (-want +got)\n:%s", diff)
			}
		})
	}
}
