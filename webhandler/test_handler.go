// Copyright 2024 Bill Nixon. All rights reserved.
// Use of this source code is governed by the license found in the LICENSE file.

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

// TestCase defines a structure for parameters and expected results for
// handler tests.
type TestCase struct {
	Name               string        // Test case name.
	Target             string        // Request target URL.
	RequestMethod      string        // HTTP method.
	RequestHeaders     http.Header   // Headers to include in request.
	RequestCookies     []http.Cookie // Cookies to include in request.
	RequestBody        string        // Request body content.
	WantStatus         int           // Expected HTTP status code.
	WantBody           string        // Expected response body.
	WantCookies        []http.Cookie // Expected cookies in response.
	WantCookiesCmpOpts cmp.Options   // Comparison options for cookies.
}

// TestHandler tests an HTTP handler with a slice of test cases.
// It automates sending requests and comparing expected to actual outcomes.
func TestHandler(t *testing.T, handlerFunc http.HandlerFunc, testCases []TestCase) {
	for _, tc := range testCases {
		t.Run(tc.Name, func(t *testing.T) {
			// Set default request target if not specified.
			if tc.Target == "" {
				tc.Target = "/test"
			}

			r := httptest.NewRequest(tc.RequestMethod, tc.Target, strings.NewReader(tc.RequestBody))

			if len(tc.RequestHeaders) > 0 {
				r.Header = tc.RequestHeaders
			}

			for _, cookie := range tc.RequestCookies {
				r.AddCookie(&cookie)
			}

			w := httptest.NewRecorder()
			handlerFunc(w, r)
			result := w.Result()

			if result.StatusCode != tc.WantStatus {
				t.Errorf("Got status code %d, want %d",
					result.StatusCode, tc.WantStatus)
			}

			body, _ := io.ReadAll(result.Body)
			got := string(body)
			if diff := cmp.Diff(got, tc.WantBody); diff != "" {
				t.Errorf("Body mismatch (-got +want)\n:%s", diff)
			}

			if len(tc.WantCookies) > 0 {
				compareCookies(t, result.Cookies(), tc.WantCookies, tc.WantCookiesCmpOpts)
			} else if len(result.Cookies()) > 0 {
				t.Errorf("Expected no cookies, got %+v", result.Cookies())
			}
		})
	}
}

// compareCookies compares expected and actual slices of cookies.
func compareCookies(t *testing.T, got []*http.Cookie, want []http.Cookie, cmpOpts cmp.Options) {
	wantPtrs := toPointerSlice(want)
	sortCookies(got)
	sortCookies(wantPtrs)
	if diff := cmp.Diff(got, wantPtrs, cmpOpts); diff != "" {
		t.Errorf("Cookies mismatch (-got +want):\n%s", diff)
	}
}

// sortCookies sorts a slice of pointers to http.Cookie by their Name attribute.
func sortCookies(cookies []*http.Cookie) {
	sort.Slice(cookies, func(i, j int) bool {
		return cookies[i].Name < cookies[j].Name
	})
}

// toPointerSlice converts a slice of http.Cookie to a slice of *http.Cookie.
func toPointerSlice(cookies []http.Cookie) []*http.Cookie {
	pointers := make([]*http.Cookie, len(cookies))
	for i := range cookies {
		pointers[i] = &cookies[i]
	}
	return pointers
}
