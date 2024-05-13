// Copyright 2024 Bill Nixon. All rights reserved.
// Use of this source code is governed by the license found in the LICENSE file.

package webauth_test

import (
	"bytes"
	"net/http"
	"path/filepath"
	"testing"
	"text/template"

	"github.com/bnixon67/webapp/assets"
	"github.com/bnixon67/webapp/csv"
	"github.com/bnixon67/webapp/webauth"
	"github.com/bnixon67/webapp/webhandler"
	"github.com/bnixon67/webapp/webutil"
)

func eventsBody(t *testing.T, data webauth.EventsPageData) string {
	tmplName := "events.html"

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

func TestEventsHandler(t *testing.T) {
	app := AppForTest(t)

	// TODO: better way to define a test user
	userToken, err := app.LoginUser("test", "password")
	if err != nil {
		t.Fatalf("could not login user to get login token")
	}
	user, err := app.DB.UserForLoginToken(userToken.Value)
	if err != nil {
		t.Fatalf("could not get user")
	}
	adminToken, err := app.LoginUser("admin", "password")
	if err != nil {
		t.Fatalf("could not login user to get login token")
	}
	admin, err := app.DB.UserForLoginToken(adminToken.Value)
	if err != nil {
		t.Fatalf("could not get user")
	}

	events, err := app.DB.GetEvents()
	if err != nil {
		t.Fatalf("failed GetEvents: %v", err)
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
			WantStatus:    http.StatusOK,
			WantBody: eventsBody(t, webauth.EventsPageData{
				CommonData: webauth.CommonData{
					Title: app.Cfg.App.Name,
				},
			}),
		},
		{
			Name:          "Valid GET Request with Bad Login Token",
			Target:        "/events",
			RequestMethod: http.MethodGet,
			RequestCookies: []http.Cookie{
				{Name: webauth.LoginTokenCookieName, Value: "foo"},
			},
			WantStatus: http.StatusOK,
			WantBody: eventsBody(t, webauth.EventsPageData{
				CommonData: webauth.CommonData{
					Title: app.Cfg.App.Name,
				},
			}),
			WantCookies: []http.Cookie{http.Cookie{Name: "login", MaxAge: -1, Raw: "login=; Max-Age=0"}},
		},
		{
			Name:          "Valid GET Request with Good Login Token - Non Admin",
			Target:        "/events",
			RequestMethod: http.MethodGet,
			RequestCookies: []http.Cookie{
				{Name: webauth.LoginTokenCookieName, Value: userToken.Value},
			},
			WantStatus: http.StatusOK,
			WantBody: eventsBody(t, webauth.EventsPageData{
				CommonData: webauth.CommonData{
					Title: app.Cfg.App.Name,
				},
				User: user,
			}),
		},
		{
			Name:          "Valid GET Request with Good Login Token - Admin",
			Target:        "/events",
			RequestMethod: http.MethodGet,
			RequestCookies: []http.Cookie{
				{Name: webauth.LoginTokenCookieName, Value: adminToken.Value},
			},
			WantStatus: http.StatusOK,
			WantBody: eventsBody(t, webauth.EventsPageData{
				CommonData: webauth.CommonData{
					Title: app.Cfg.App.Name,
				},
				User: admin, Events: events,
			}),
		},
	}

	// Test the handler using the utility function.
	webhandler.TestHandler(t, app.EventsHandler, tests)
}

func TestEventsCSVHandler(t *testing.T) {
	app := AppForTest(t)

	// TODO: better way to define a test user
	userToken, err := app.LoginUser("test", "password")
	if err != nil {
		t.Fatalf("could not login user to get login token")
	}
	adminToken, err := app.LoginUser("admin", "password")
	if err != nil {
		t.Fatalf("could not login user to get login token")
	}

	events, err := app.DB.GetEvents()
	if err != nil {
		t.Fatalf("failed GetEvents: %v", err)
	}
	var eventsBody bytes.Buffer
	err = csv.SliceOfStructsToCSV(&eventsBody, events)
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
			Name:          "Valid GET Request with Bad Login Token",
			Target:        "/events",
			RequestMethod: http.MethodGet,
			RequestCookies: []http.Cookie{
				{Name: webauth.LoginTokenCookieName, Value: "foo"},
			},
			WantStatus:  http.StatusUnauthorized,
			WantBody:    "Error: Unauthorized\n",
			WantCookies: []http.Cookie{http.Cookie{Name: "login", MaxAge: -1, Raw: "login=; Max-Age=0"}},
		},
		{
			Name:          "Valid GET Request with Good Login Token - Non Admin",
			Target:        "/events",
			RequestMethod: http.MethodGet,
			RequestCookies: []http.Cookie{
				{Name: webauth.LoginTokenCookieName, Value: userToken.Value},
			},
			WantStatus: http.StatusUnauthorized,
			WantBody:   "Error: Unauthorized\n",
		},
		{
			Name:          "Valid GET Request with Good Login Token - Admin",
			Target:        "/events",
			RequestMethod: http.MethodGet,
			RequestCookies: []http.Cookie{
				{Name: webauth.LoginTokenCookieName, Value: adminToken.Value},
			},
			WantStatus: http.StatusOK,
			WantBody:   eventsBody.String(),
		},
	}

	// Test the handler using the utility function.
	webhandler.TestHandler(t, app.EventsCSVHandler, tests)
}
