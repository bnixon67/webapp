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
	"github.com/bnixon67/webapp/webauth"
	"github.com/bnixon67/webapp/webhandler"
)

func userBody(data webauth.UserPageData) string {
	// Get path to template file.
	assetDir := assets.AssetPath()
	tmplFile := filepath.Join(assetDir, "tmpl", "user.html")

	// Parse the HTML template from a file.
	tmpl := template.Must(template.ParseFiles(tmplFile))

	// Create a buffer to store the rendered HTML.
	var body bytes.Buffer

	// Execute the template with the data and write the result to the buffer.
	tmpl.Execute(&body, data)

	return body.String()
}

func TestUserGetHandler(t *testing.T) {
	app := AppForTest(t)

	// TODO: better way to define a test user
	token, err := app.LoginUser("test", "password")
	if err != nil {
		t.Errorf("could not login user to get login token")
	}
	user, err := app.DB.UserForLoginToken(token.Value)
	if err != nil {
		t.Errorf("could not get user")
	}

	tests := []webhandler.TestCase{
		{
			Name:          "Invalid Method",
			Target:        "/user",
			RequestMethod: http.MethodPost,
			WantStatus:    http.StatusMethodNotAllowed,
			WantBody:      "Error: Method Not Allowed\n",
		},
		{
			Name:          "Valid GET without Cookie",
			Target:        "/user",
			RequestMethod: http.MethodGet,
			WantStatus:    http.StatusOK,
			WantBody: userBody(webauth.UserPageData{
				Title: app.Cfg.App.Name,
			}),
		},
		{
			Name:          "Valid GET with Bad Login Token",
			Target:        "/user",
			RequestMethod: http.MethodGet,
			RequestCookies: []http.Cookie{
				{
					Name:  webauth.LoginTokenCookieName,
					Value: "foo",
				},
			},
			WantStatus: http.StatusOK,
			WantBody: userBody(webauth.UserPageData{
				Title: app.Cfg.App.Name,
			}),
			WantCookies: []http.Cookie{
				{
					Name:   "login",
					MaxAge: -1,
					Raw:    "login=; Max-Age=0",
				},
			},
		},
		{
			Name:          "Valid GET with Good Login Token",
			Target:        "/user",
			RequestMethod: http.MethodGet,
			RequestCookies: []http.Cookie{
				{
					Name:  webauth.LoginTokenCookieName,
					Value: token.Value,
				},
			},
			WantStatus: http.StatusOK,
			WantBody: userBody(webauth.UserPageData{
				Title: app.Cfg.App.Name, User: user,
			}),
		},
	}

	// Test the handler using the utility function.
	webhandler.HandlerTestWithCases(t, app.UserGetHandler, tests)
}
