// Copyright 2023 Bill Nixon. All rights reserved.
// Use of this source code is governed by the license found in the LICENSE file.

package weblogin_test

import (
	"bytes"
	"net/http"
	"path/filepath"
	"testing"
	"text/template"

	"github.com/bnixon67/webapp/assets"
	"github.com/bnixon67/webapp/webhandler"
	"github.com/bnixon67/webapp/weblogin"
	"github.com/bnixon67/webapp/webutil"
)

func usersBody(t *testing.T, data weblogin.UsersPageData) string {
	tmplName := "users.html"

	// Initialize FuncMap with the custom function.
	funcMap := template.FuncMap{"ToTimeZone": webutil.ToTimeZone}

	// Directly include the name of the template in New for clarity.
	tmpl := template.New(tmplName).Funcs(funcMap)

	// Get path to template file.
	assetDir := assets.AssetPath()
	tmplFile := filepath.Join(assetDir, "tmpl", tmplName)

	// Parse the template file, checking for errors.
	tmpl, err := tmpl.ParseFiles(tmplFile)
	if err != nil {
		t.Fatalf("could not parse template file '%s': %v", tmplFile, err)
	}

	// Create a buffer to store the rendered HTML.
	var body bytes.Buffer

	// Execute the template with the data and write the result to the buffer.
	tmpl.Execute(&body, data)

	return body.String()
}

func TestUsersHandler(t *testing.T) {
	app := AppForTest(t)

	// TODO: better way to define a test user
	userToken, err := app.LoginUser("test", "password")
	if err != nil {
		t.Fatalf("could not login user to get session token")
	}
	user, err := app.DB.GetUserForSessionToken(userToken.Value)
	if err != nil {
		t.Fatalf("could not get user")
	}
	adminToken, err := app.LoginUser("admin", "password")
	if err != nil {
		t.Fatalf("could not login user to get session token")
	}
	admin, err := app.DB.GetUserForSessionToken(adminToken.Value)
	if err != nil {
		t.Fatalf("could not get user")
	}

	users, err := weblogin.GetUsers(app.DB)
	if err != nil {
		t.Fatalf("failed GetUsers: %v", err)
	}

	tests := []webhandler.TestCase{
		{
			Name:          "Invalid Method",
			Target:        "/users",
			RequestMethod: http.MethodPost,
			WantStatus:    http.StatusMethodNotAllowed,
			WantBody:      "POST Method Not Allowed\n",
		},
		{
			Name:          "Valid GET Request without Cookie",
			Target:        "/users",
			RequestMethod: http.MethodGet,
			WantStatus:    http.StatusOK,
			WantBody: usersBody(t, weblogin.UsersPageData{
				Title: app.Cfg.App.Name,
			}),
		},
		{
			Name:          "Valid GET Request with Bad Session Token",
			Target:        "/users",
			RequestMethod: http.MethodGet,
			RequestCookies: []http.Cookie{
				{Name: weblogin.SessionTokenCookieName, Value: "foo"},
			},
			WantStatus: http.StatusOK,
			WantBody: usersBody(t, weblogin.UsersPageData{
				Title: app.Cfg.App.Name,
			}),
		},
		{
			Name:          "Valid GET Request with Good Session Token - Non Admin",
			Target:        "/users",
			RequestMethod: http.MethodGet,
			RequestCookies: []http.Cookie{
				{Name: weblogin.SessionTokenCookieName, Value: userToken.Value},
			},
			WantStatus: http.StatusOK,
			WantBody: usersBody(t, weblogin.UsersPageData{
				Title: app.Cfg.App.Name, User: user, Users: users,
			}),
		},
		{
			Name:          "Valid GET Request with Good Session Token - Admin",
			Target:        "/users",
			RequestMethod: http.MethodGet,
			RequestCookies: []http.Cookie{
				{Name: weblogin.SessionTokenCookieName, Value: adminToken.Value},
			},
			WantStatus: http.StatusOK,
			WantBody: usersBody(t, weblogin.UsersPageData{
				Title: app.Cfg.App.Name, User: admin, Users: users,
			}),
		},
	}

	// Test the handler using the utility function.
	webhandler.HandlerTestWithCases(t, app.UsersHandler, tests)
}

func TestUsersCSVHandler(t *testing.T) {
	app := AppForTest(t)

	// TODO: better way to define a test user
	userToken, err := app.LoginUser("test", "password")
	if err != nil {
		t.Fatalf("could not login user to get session token")
	}
	adminToken, err := app.LoginUser("admin", "password")
	if err != nil {
		t.Fatalf("could not login user to get session token")
	}

	events, err := weblogin.GetUsers(app.DB)
	if err != nil {
		t.Fatalf("failed GetUsers: %v", err)
	}
	var eventsBody bytes.Buffer
	err = webutil.SliceOfStructsToCSV(&eventsBody, events)
	if err != nil {
		t.Fatalf("failed SliceOfStructsToCSV: %v", err)
	}

	tests := []webhandler.TestCase{
		{
			Name:          "Invalid Method",
			Target:        "/events",
			RequestMethod: http.MethodPost,
			WantStatus:    http.StatusMethodNotAllowed,
			WantBody:      "POST Method Not Allowed\n",
		},
		{
			Name:          "Valid GET Request without Cookie",
			Target:        "/events",
			RequestMethod: http.MethodGet,
			WantStatus:    http.StatusUnauthorized,
			WantBody:      "Error: Unauthorized\n",
		},
		{
			Name:          "Valid GET Request with Bad Session Token",
			Target:        "/events",
			RequestMethod: http.MethodGet,
			RequestCookies: []http.Cookie{
				{Name: weblogin.SessionTokenCookieName, Value: "foo"},
			},
			WantStatus: http.StatusUnauthorized,
			WantBody:   "Error: Unauthorized\n",
		},
		{
			Name:          "Valid GET Request with Good Session Token - Non Admin",
			Target:        "/events",
			RequestMethod: http.MethodGet,
			RequestCookies: []http.Cookie{
				{Name: weblogin.SessionTokenCookieName, Value: userToken.Value},
			},
			WantStatus: http.StatusUnauthorized,
			WantBody:   "Error: Unauthorized\n",
		},
		{
			Name:          "Valid GET Request with Good Session Token - Admin",
			Target:        "/events",
			RequestMethod: http.MethodGet,
			RequestCookies: []http.Cookie{
				{Name: weblogin.SessionTokenCookieName, Value: adminToken.Value},
			},
			WantStatus: http.StatusOK,
			WantBody:   eventsBody.String(),
		},
	}

	// Test the handler using the utility function.
	webhandler.HandlerTestWithCases(t, app.UsersCSVHandler, tests)
}
