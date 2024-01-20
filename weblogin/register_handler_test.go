// Copyright 2023 Bill Nixon. All rights reserved.
// Use of this source code is governed by the license found in the LICENSE file.

package weblogin_test

import (
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"

	"github.com/bnixon67/webapp/weblogin"
)

func TestRegisterHandlerInvalidMethod(t *testing.T) {
	app := AppForTest(t)

	w := httptest.NewRecorder()
	r := httptest.NewRequest(http.MethodPatch, "/register", nil)

	app.RegisterHandler(w, r)

	expectedStatus := http.StatusMethodNotAllowed
	if w.Code != expectedStatus {
		t.Errorf("got status %d %q, expected %d %q", w.Code, http.StatusText(w.Code), expectedStatus, http.StatusText(expectedStatus))
	}
}

func TestRegisterHandlerGet(t *testing.T) {
	app := AppForTest(t)

	w := httptest.NewRecorder()
	r := httptest.NewRequest(http.MethodGet, "/register", nil)

	app.RegisterHandler(w, r)

	expectedStatus := http.StatusOK
	if w.Code != expectedStatus {
		t.Errorf("got status %d %q, expected %d %q", w.Code, http.StatusText(w.Code), expectedStatus, http.StatusText(expectedStatus))
	}

	expectedInBody := "Register"
	if !strings.Contains(w.Body.String(), expectedInBody) {
		t.Errorf("got body %q, expected %q in body", w.Body, expectedInBody)
	}

	got := w.Header().Get("Location")
	expected := ""
	if got != expected {
		t.Errorf("got location %q, expected %q", got, expected)
	}
}

func TestRegisterHandlerPostMissingValues(t *testing.T) {
	app := AppForTest(t)

	w := httptest.NewRecorder()
	r := httptest.NewRequest(http.MethodPost, "/register", nil)

	app.RegisterHandler(w, r)

	expectedStatus := http.StatusOK
	if w.Code != expectedStatus {
		t.Errorf("got status %d %q, expected %d %q", w.Code, http.StatusText(w.Code), expectedStatus, http.StatusText(expectedStatus))
	}

	expectedInBody := weblogin.MsgMissingRequired
	if !strings.Contains(w.Body.String(), expectedInBody) {
		t.Errorf("got body %q, expected %q in body", w.Body, expectedInBody)
	}

	got := w.Header().Get("Location")
	expected := ""
	if got != expected {
		t.Errorf("got location %q, expected %q", got, expected)
	}
}

func TestRegisterHandlerPostExistingUser(t *testing.T) {
	data := url.Values{
		"userName":  {"test"},
		"fullName":  {"full name"},
		"email":     {"email"},
		"password1": {"password one"},
		"password2": {"password one"},
	}

	app := AppForTest(t)

	w := httptest.NewRecorder()
	r := httptest.NewRequest(http.MethodPost, "/register", strings.NewReader(data.Encode()))
	r.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	app.RegisterHandler(w, r)

	expectedStatus := http.StatusOK
	if w.Code != expectedStatus {
		t.Errorf("got status %d %q, expected %d %q", w.Code, http.StatusText(w.Code), expectedStatus, http.StatusText(expectedStatus))
	}

	expectedInBody := weblogin.MsgUsernameExists
	if !strings.Contains(w.Body.String(), expectedInBody) {
		t.Errorf("got body %q, expected %q in body", w.Body, expectedInBody)
	}

	got := w.Header().Get("Location")
	expected := ""
	if got != expected {
		t.Errorf("got location %q, expected %q", got, expected)
	}
}

func TestRegisterHandlerPostExistingEmail(t *testing.T) {
	randomUsername, err := weblogin.GenerateRandomString(8)
	if err != nil {
		t.Errorf("could not GenerateRandomString")
	}
	data := url.Values{
		"userName":  {randomUsername},
		"fullName":  {"full name"},
		"email":     {"test@email"},
		"password1": {"password one"},
		"password2": {"password one"},
	}

	app := AppForTest(t)

	w := httptest.NewRecorder()
	r := httptest.NewRequest(http.MethodPost, "/register", strings.NewReader(data.Encode()))
	r.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	app.RegisterHandler(w, r)

	expectedStatus := http.StatusOK
	if w.Code != expectedStatus {
		t.Errorf("got status %d %q, expected %d %q", w.Code, http.StatusText(w.Code), expectedStatus, http.StatusText(expectedStatus))
	}

	expectedInBody := weblogin.MsgEmailExists
	if !strings.Contains(w.Body.String(), expectedInBody) {
		t.Errorf("got body %q, expected %q in body", w.Body, expectedInBody)
	}

	got := w.Header().Get("Location")
	expected := ""
	if got != expected {
		t.Errorf("got location %q, expected %q", got, expected)
	}
}

func TestRegisterHandlerPostMismatchedPassword(t *testing.T) {
	randomUsername, err := weblogin.GenerateRandomString(8)
	if err != nil {
		t.Errorf("could not GenerateRandomString")
	}
	data := url.Values{
		"userName":  {randomUsername},
		"fullName":  {"full name"},
		"email":     {randomUsername + "@email"},
		"password1": {"password one"},
		"password2": {"password two"},
	}

	app := AppForTest(t)

	w := httptest.NewRecorder()
	r := httptest.NewRequest(http.MethodPost, "/register", strings.NewReader(data.Encode()))
	r.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	app.RegisterHandler(w, r)

	expectedStatus := http.StatusOK
	if w.Code != expectedStatus {
		t.Errorf("got status %d %q, expected %d %q", w.Code, http.StatusText(w.Code), expectedStatus, http.StatusText(expectedStatus))
	}

	expectedInBody := weblogin.MsgPasswordsDifferent
	if !strings.Contains(w.Body.String(), expectedInBody) {
		t.Errorf("got body %q, expected %q in body", w.Body, expectedInBody)
	}

	got := w.Header().Get("Location")
	expected := ""
	if got != expected {
		t.Errorf("got location %q, expected %q", got, expected)
	}
}

func TestRegisterHandlerPostMissingPassword1(t *testing.T) {
	randomUsername, err := weblogin.GenerateRandomString(8)
	if err != nil {
		t.Errorf("could not GenerateRandomString")
	}
	data := url.Values{
		"userName":  {randomUsername},
		"fullName":  {"full name"},
		"email":     {randomUsername + "@email"},
		"password1": {""},
		"password2": {"password two"},
	}

	app := AppForTest(t)

	w := httptest.NewRecorder()
	r := httptest.NewRequest(http.MethodPost, "/register", strings.NewReader(data.Encode()))
	r.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	app.RegisterHandler(w, r)

	expectedStatus := http.StatusOK
	if w.Code != expectedStatus {
		t.Errorf("got status %d %q, expected %d %q", w.Code, http.StatusText(w.Code), expectedStatus, http.StatusText(expectedStatus))
	}

	expectedInBody := weblogin.MsgMissingRequired
	if !strings.Contains(w.Body.String(), expectedInBody) {
		t.Errorf("got body %q, expected %q in body", w.Body, expectedInBody)
	}

	got := w.Header().Get("Location")
	expected := ""
	if got != expected {
		t.Errorf("got location %q, expected %q", got, expected)
	}
}

func TestRegisterHandlerPostValid(t *testing.T) {
	randomUsername, err := weblogin.GenerateRandomString(8)
	if err != nil {
		t.Errorf("could not GenerateRandomString")
	}
	data := url.Values{
		"userName":  {randomUsername},
		"fullName":  {"full name"},
		"email":     {randomUsername + "@email"},
		"password1": {"password"},
		"password2": {"password"},
	}

	app := AppForTest(t)

	w := httptest.NewRecorder()
	r := httptest.NewRequest(http.MethodPost, "/register", strings.NewReader(data.Encode()))
	r.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	app.RegisterHandler(w, r)

	expectedStatus := http.StatusSeeOther
	if w.Code != expectedStatus {
		t.Errorf("got status %d %q, expected %d %q", w.Code, http.StatusText(w.Code), expectedStatus, http.StatusText(expectedStatus))
	}

	expectedInBody := ""
	if !strings.Contains(w.Body.String(), expectedInBody) {
		t.Errorf("got body %q, expected %q in body", w.Body, expectedInBody)
	}

	got := w.Header().Get("Location")
	expected := "/login"
	if got != expected {
		t.Errorf("got location %q, expected %q", got, expected)
	}
}
