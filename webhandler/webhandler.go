// Copyright 2023 Bill Nixon. All rights reserved.
// Use of this source code is governed by the license found in the LICENSE file.

// Package webhandler provides handlers, middleware, and utilities for web applications.
// It simplifies common tasks, enhances request processing, and includes features like request logging, unique request IDs, and HTML template rendering.
package webhandler

import (
	"io"
	"net/http"
	"net/http/httptest"
	"sort"
	"strings"
	"testing"

	"github.com/google/go-cmp/cmp"
)

type TestCase struct {
	Name               string
	Target             string
	RequestMethod      string
	RequestHeaders     http.Header
	RequestCookies     []http.Cookie
	RequestBody        string
	WantStatus         int
	WantBody           string
	WantCookies        []http.Cookie
	WantCookiesCmpOpts cmp.Options
}

// HandlerTestWithCases is a utility function for testing a handler.
func HandlerTestWithCases(t *testing.T, handlerFunc http.HandlerFunc, testCases []TestCase) {
	for _, tt := range testCases {
		t.Run(tt.Name, func(t *testing.T) {
			// Handle empty Target
			if tt.Target == "" {
				tt.Target = "/test"
			}

			req := httptest.NewRequest(tt.RequestMethod, tt.Target, strings.NewReader(tt.RequestBody))

			if len(tt.RequestHeaders) > 0 {
				req.Header = tt.RequestHeaders
			}

			for _, cookie := range tt.RequestCookies {
				req.AddCookie(&cookie)
			}

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

			if tt.WantCookies == nil {
				if len(resp.Cookies()) != 0 {
					t.Errorf("Want no cookies, got cookies\n%+v", resp.Cookies())
				}
			} else {
				// Convert wantCookies to []*http.Cookie
				wantPtrCookies := toPointerSlice(tt.WantCookies)

				gotCookies := resp.Cookies()

				// Sort both slices
				sortPointerCookies(wantPtrCookies)
				sortPointerCookies(gotCookies)

				diff = cmp.Diff(wantPtrCookies, gotCookies, tt.WantCookiesCmpOpts)
				if diff != "" {
					t.Errorf("Cookies mismatch (-want +got)\n:%s", diff)
				}
			}
		})
	}
}

// Sort slice of http.Cookie
func sortValueCookies(cookies []http.Cookie) {
	sort.Slice(cookies, func(i, j int) bool {
		return cookies[i].Name < cookies[j].Name
	})
}

// Sort slice of *http.Cookie
func sortPointerCookies(cookies []*http.Cookie) {
	sort.Slice(cookies, func(i, j int) bool {
		return cookies[i].Name < cookies[j].Name
	})
}

// Convert []http.Cookie to []*http.Cookie
func toPointerSlice(cookies []http.Cookie) []*http.Cookie {
	pointers := make([]*http.Cookie, len(cookies))
	for i := range cookies {
		pointers[i] = &cookies[i]
	}
	return pointers
}
