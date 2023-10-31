package weblogin

import (
	"encoding/base64"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestGenerateRandomString(t *testing.T) {
	tests := []struct {
		name    string
		n       int
		wantErr bool
	}{
		{"ValidInput", 16, false},
		{"NegativeInput", -1, true},
		{"ZeroInput", 0, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := GenerateRandomString(tt.n)

			if (err != nil) != tt.wantErr {
				t.Errorf("GenerateRandomString() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if !tt.wantErr {
				decoded, err := base64.URLEncoding.DecodeString(got)
				if err != nil {
					t.Errorf("Generated string is not a valid base64 URL encoding")
					return
				}

				if len(decoded) != tt.n {
					t.Errorf("Generated string length = %d, expected %d", len(decoded), tt.n)
					return
				}
			}
		})
	}
}

func TestGetCookieValue(t *testing.T) {
	cases := []struct {
		request    bool
		cookie     *http.Cookie
		name, want string
		err        error
	}{
		{
			request: false, cookie: nil,
			name: "", want: "", err: ErrRequestNil,
		},
		{
			request: true, cookie: nil,
			name: "none", want: "", err: nil,
		},
		{
			request: true,
			cookie:  &http.Cookie{Name: "test", Value: "value"},
			name:    "test", want: "value", err: nil,
		},
		{
			request: true,
			cookie:  &http.Cookie{Name: "test", Value: "value"},
			name:    "none", want: "", err: nil,
		},
	}

	for _, tc := range cases {
		var r *http.Request

		if tc.request {
			r = httptest.NewRequest(http.MethodGet, "/test", nil)
		}
		if tc.cookie != nil {
			r.AddCookie(tc.cookie)
		}

		got, err := GetCookieValue(r, tc.name)
		if !errors.Is(err, tc.err) {
			t.Errorf("GetCookieValue(%v, %q)\ngot err '%v' want '%v'", r, tc.name, err, tc.err)
		}
		if got != tc.want {
			t.Errorf("GetCookieValue(%v, %q)\ngot %q want %q", r, tc.name, got, tc.want)
		}
	}
}
