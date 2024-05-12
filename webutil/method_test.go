// Copyright 2024 Bill Nixon. All rights reserved.
// Use of this source code is governed by the license found in the LICENSE file.

package webutil_test

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/bnixon67/webapp/webutil"
)

func TestIsMethodOrError(t *testing.T) {
	tests := []struct {
		name          string
		testMethod    string
		requestMethod string
		want          bool
	}{
		{
			name:          "ValidMethod",
			testMethod:    http.MethodGet,
			requestMethod: http.MethodGet,
			want:          true,
		},
		{
			name:          "InvalidMethod",
			testMethod:    http.MethodGet,
			requestMethod: http.MethodPost,
			want:          false,
		},
		{
			name:          "ValidMethodCaseSensitive",
			testMethod:    "get",
			requestMethod: http.MethodGet,
			want:          false,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			r := httptest.NewRequest(tc.requestMethod, "/", nil)
			w := httptest.NewRecorder()

			got := webutil.IsMethodOrError(w, r, tc.testMethod)

			if got != tc.want {
				t.Errorf("Got %v, want %v for test method %v and request method %v", got, tc.want, tc.testMethod, tc.requestMethod)
			}

			if tc.want == false {
				result := w.Result()
				if result.StatusCode != http.StatusMethodNotAllowed {
					t.Errorf("Got status code %d, want %d for test method %v and request method %v", result.StatusCode, http.StatusMethodNotAllowed, tc.testMethod, tc.requestMethod)
				}
			}
		})
	}
}

func TestCheckAllowedMethods(t *testing.T) {
	tests := []struct {
		name           string
		requestMethod  string
		allowed        []string
		want           bool
		wantStatusCode int
		wantBody       string
	}{
		{
			name:           "TestAllowedGET",
			requestMethod:  http.MethodGet,
			allowed:        []string{http.MethodGet},
			want:           true,
			wantStatusCode: http.StatusOK,
			wantBody:       "",
		},
		{
			name:           "TestAllowedGETMultiple",
			requestMethod:  http.MethodGet,
			allowed:        []string{http.MethodPut, http.MethodGet},
			want:           true,
			wantStatusCode: http.StatusOK,
			wantBody:       "",
		},
		{
			name:           "TestNotAllowedPOST",
			requestMethod:  http.MethodPost,
			allowed:        []string{http.MethodGet},
			want:           false,
			wantStatusCode: http.StatusMethodNotAllowed,
			wantBody:       "POST Method Not Allowed\n",
		},
		{
			name:           "TestOPTIONS",
			requestMethod:  http.MethodOptions,
			allowed:        []string{http.MethodGet},
			wantBody:       "",
			wantStatusCode: http.StatusNoContent,
			want:           false,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			r := httptest.NewRequest(tc.requestMethod, "/", http.NoBody)
			w := httptest.NewRecorder()

			got := webutil.CheckAllowedMethods(w, r, tc.allowed...)

			if got != tc.want {
				t.Errorf("Want %v, got %v for allowed methods %v and request method %v", tc.want, got, tc.allowed, tc.requestMethod)
			}

			if w.Code != tc.wantStatusCode {
				t.Errorf("Want status %v, got %v for allowed methods %v and request method %v", tc.wantStatusCode, w.Code, tc.allowed, tc.requestMethod)
			}

			gotBody := w.Body.String()
			if gotBody != tc.wantBody {
				t.Errorf("Want body %q, got %q for allowed methods %v and request method %v", gotBody, tc.wantBody, tc.allowed, tc.requestMethod)
			}
		})
	}
}
