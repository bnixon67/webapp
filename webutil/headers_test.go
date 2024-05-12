// Copyright 2024 Bill Nixon. All rights reserved.
// Use of this source code is governed by the license found in the LICENSE file.

package webutil_test

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/bnixon67/webapp/webutil"
	"github.com/google/go-cmp/cmp"
)

func TestSetNoCacheHeaders(t *testing.T) {
	w := httptest.NewRecorder()

	webutil.SetNoCacheHeaders(w)
	headers := w.Header()

	expectedHeaders := http.Header{
		"Cache-Control": []string{"no-cache, no-store, must-revalidate"},
		"Expires":       []string{"0"},
		"Pragma":        []string{"no-cache"},
	}

	if diff := cmp.Diff(expectedHeaders, headers); diff != "" {
		t.Errorf("Headers mismatch (-want +got)\n%s", diff)

	}
}

func testContentTypeSetting(t *testing.T, contentTypeFunc func(w http.ResponseWriter), expectedContentType string) {
	w := httptest.NewRecorder()

	// Call the function that sets the content type
	contentTypeFunc(w)
	headers := w.Header()

	// Define expected headers
	expectedHeaders := http.Header{
		"X-Content-Type-Options": []string{"nosniff"},
		"Content-Type":           []string{expectedContentType},
	}

	// Compare the actual headers against the expected headers
	if diff := cmp.Diff(expectedHeaders, headers); diff != "" {
		t.Errorf("Headers mismatch (-want +got)\n%s", diff)
	}
}

func TestSetContentType(t *testing.T) {
	contentType := "foo"

	testContentTypeSetting(t, func(w http.ResponseWriter) {
		webutil.SetContentType(w, contentType)
	}, contentType)
}

func TestSetContentTypeText(t *testing.T) {
	testContentTypeSetting(t, webutil.SetContentTypeText, "text/plain;charset=utf-8")
}

func TestSetContentTypeHTML(t *testing.T) {
	testContentTypeSetting(t, webutil.SetContentTypeHTML, "text/html;charset=utf-8")
}

func TestClientIP(t *testing.T) {
	tests := []struct {
		name       string
		realIP     string
		remoteAddr string
		want       string
	}{
		{
			name:       "HasRealIP",
			realIP:     "192.168.1.1",
			remoteAddr: "192.168.1.100:9876",
			want:       "192.168.1.1",
		},
		{
			name:       "NoRealIP",
			realIP:     "",
			remoteAddr: "192.168.1.100:9876",
			want:       "192.168.1.100:9876",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			r := httptest.NewRequest("GET", "/", nil)
			if tc.realIP != "" {
				r.Header.Set("X-Real-IP", tc.realIP)
			}
			r.RemoteAddr = tc.remoteAddr
			result := webutil.ClientIP(r)
			if result != tc.want {
				t.Errorf("Expected %v, got %v", tc.want, result)
			}
		})
	}
}
