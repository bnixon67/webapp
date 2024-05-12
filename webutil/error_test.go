// Copyright 2024 Bill Nixon. All rights reserved.
// Use of this source code is governed by the license found in the LICENSE file.

package webutil_test

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/bnixon67/webapp/webutil"
)

// TestRespondWithError checks if the function sends the correct HTTP status
// code and the appropriate error message.
func TestRespondWithError(t *testing.T) {
	tests := []struct {
		name     string
		code     int
		wantCode int
		wantBody string
	}{
		{
			name:     "BadRequest",
			code:     http.StatusBadRequest,
			wantCode: http.StatusBadRequest,
			wantBody: "Error: Bad Request",
		},
		{
			name:     "NotFound",
			code:     http.StatusNotFound,
			wantCode: http.StatusNotFound,
			wantBody: "Error: Not Found",
		},
		{
			name:     "InternalServerError",
			code:     http.StatusInternalServerError,
			wantCode: http.StatusInternalServerError,
			wantBody: "Error: Internal Server Error",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			w := httptest.NewRecorder()

			webutil.RespondWithError(w, tc.code)

			if w.Code != tc.wantCode {
				t.Errorf("got code %v, want %v", w.Code, tc.wantCode)
			}

			gotBody := w.Body.String()
			wantBody := tc.wantBody + "\n" // http.Error adds newline
			if gotBody != wantBody {
				t.Errorf("got body %q, want %q", gotBody, wantBody)
			}
		})
	}
}
