// Copyright 2023 Bill Nixon. All rights reserved.
// Use of this source code is governed by the license found in the LICENSE file.

package weblogin_test

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/bnixon67/webapp/weblogin"
)

func TestLogoutHandlerInvalidMethod(t *testing.T) {
	app := AppForTest(t)

	w := httptest.NewRecorder()
	r := httptest.NewRequest(http.MethodPost, "/logout", nil)

	app.LogoutHandler(w, r)

	expectedStatus := http.StatusMethodNotAllowed
	if w.Code != expectedStatus {
		t.Errorf("got status %d %q, expected %d %q",
			w.Code, http.StatusText(w.Code), expectedStatus, http.StatusText(expectedStatus))
	}
}

func getCookie(name string, cookies []*http.Cookie) (*http.Cookie, error) {
	for _, c := range cookies {
		if name == c.Name {
			return c, nil
		}
	}
	return nil, http.ErrNoCookie
}

func TestLogoutHandlerGetNoSessionToken(t *testing.T) {
	app := AppForTest(t)

	w := httptest.NewRecorder()
	r := httptest.NewRequest(http.MethodGet, "/logout", nil)

	app.LogoutHandler(w, r)

	expectedStatus := http.StatusOK
	if w.Code != expectedStatus {
		t.Errorf("got status %d %q, expected %d %q",
			w.Code, http.StatusText(w.Code), expectedStatus, http.StatusText(expectedStatus))
	}

	expectedInBody := "You have been logged out."
	if !strings.Contains(w.Body.String(), expectedInBody) {
		t.Errorf("got body %q, expected %q in body",
			w.Body, expectedInBody)
	}

	c, err := getCookie(weblogin.SessionTokenCookieName, w.Result().Cookies())
	if err != nil {
		t.Errorf("sessionToken cookie missing")
	}
	if c.Value != "" {
		t.Errorf("sessionToken not empty")
	}
	if c.MaxAge != -1 {
		t.Errorf("sessionToken.MaxAge not -1")
	}
}

func TestLogoutHandlerGetWithGoodSessionToken(t *testing.T) {
	app := AppForTest(t)

	// TODO: better way to define a test user
	token, err := app.LoginUser("test", "password")
	if err != nil {
		t.Errorf("could not login user to get session token")
	}

	w := httptest.NewRecorder()
	r := httptest.NewRequest(http.MethodGet, "/logout", nil)
	r.AddCookie(&http.Cookie{
		Name: weblogin.SessionTokenCookieName, Value: token.Value,
	})

	app.LogoutHandler(w, r)

	expectedStatus := http.StatusOK
	if w.Code != expectedStatus {
		t.Errorf("got status %d %q, expected %d %q",
			w.Code, http.StatusText(w.Code), expectedStatus, http.StatusText(expectedStatus))
	}

	expectedInBody := "You have been logged out."
	if !strings.Contains(w.Body.String(), expectedInBody) {
		t.Errorf("got body %q, expected %q in body",
			w.Body, expectedInBody)
	}

	c, err := getCookie(weblogin.SessionTokenCookieName, w.Result().Cookies())
	if err != nil {
		t.Errorf("sessionToken cookie missing")
	}
	if c.Value != "" {
		t.Errorf("sessionToken not empty")
	}
	if c.MaxAge != -1 {
		t.Errorf("sessionToken.MaxAge not -1")
	}
}
