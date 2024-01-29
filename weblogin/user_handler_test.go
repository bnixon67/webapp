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
)

func userBody(data weblogin.UserPageData) string {
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

func TestUserHandler(t *testing.T) {
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
			WantBody:      "POST Method Not Allowed\n",
		},
		{
			Name:          "Valid GET Request without Cookie",
			Target:        "/user",
			RequestMethod: http.MethodGet,
			WantStatus:    http.StatusOK,
			WantBody: userBody(weblogin.UserPageData{
				Title: app.Cfg.App.Name,
			}),
		},
		{
			Name:          "Valid GET Request with Bad Login Token",
			Target:        "/user",
			RequestMethod: http.MethodGet,
			RequestCookies: []http.Cookie{
				{Name: weblogin.LoginTokenCookieName, Value: "foo"},
			},
			WantStatus: http.StatusOK,
			WantBody: userBody(weblogin.UserPageData{
				Title: app.Cfg.App.Name,
			}),
			WantCookies: []http.Cookie{http.Cookie{Name: "login", MaxAge: -1, Raw: "login=; Max-Age=0"}},
		},
		{
			Name:          "Valid GET Request with Good Login Token",
			Target:        "/user",
			RequestMethod: http.MethodGet,
			RequestCookies: []http.Cookie{
				{Name: weblogin.LoginTokenCookieName, Value: token.Value},
			},
			WantStatus: http.StatusOK,
			WantBody: userBody(weblogin.UserPageData{
				Title: app.Cfg.App.Name, User: user,
			}),
		},
		/*
			{
				Name:           "Missing Email",
				Target:         "/forgot",
				RequestMethod:  http.MethodPost,
				RequestHeaders: header,
				RequestBody: url.Values{
					"action": {"user"},
				}.Encode(),
				WantStatus: http.StatusOK,
				WantBody: forgotBody(weblogin.ForgotPageData{
					Title:   app.Cfg.App.Name,
					Message: weblogin.MsgMissingEmail,
				}),
			},
			{
				Name:           "Missing Action",
				Target:         "/forgot",
				RequestMethod:  http.MethodPost,
				RequestHeaders: header,
				RequestBody: url.Values{
					"email": {"test@email"},
				}.Encode(),
				WantStatus: http.StatusOK,
				WantBody: forgotBody(weblogin.ForgotPageData{
					Title:   app.Cfg.App.Name,
					Message: weblogin.MsgMissingAction,
				}),
			},
			{
				Name:           "Invalid Action",
				Target:         "/forgot",
				RequestMethod:  http.MethodPost,
				RequestHeaders: header,
				RequestBody: url.Values{
					"email":  {"test@email"},
					"action": {"invalid"},
				}.Encode(),
				WantStatus: http.StatusOK,
				WantBody: forgotBody(weblogin.ForgotPageData{
					Title:   app.Cfg.App.Name,
					Message: weblogin.MsgInvalidAction,
				}),
			},
			{
				Name:           "Valid User Action",
				Target:         "/forgot",
				RequestMethod:  http.MethodPost,
				RequestHeaders: header,
				RequestBody: url.Values{
					"action": {"user"},
					"email":  {"test@email"},
				}.Encode(),
				WantStatus: http.StatusOK,
				WantBody: sentBody(weblogin.ForgotPageData{
					Title:     app.Cfg.App.Name,
					EmailFrom: app.Cfg.SMTP.User,
				}),
			},
			{
				Name:           "Valid Password Action",
				Target:         "/forgot",
				RequestMethod:  http.MethodPost,
				RequestHeaders: header,
				RequestBody: url.Values{
					"action": {"password"},
					"email":  {"test@email"},
				}.Encode(),
				WantStatus: http.StatusOK,
				WantBody: sentBody(weblogin.ForgotPageData{
					Title:     app.Cfg.App.Name,
					EmailFrom: app.Cfg.SMTP.User,
				}),
			},
			{
				Name:           "Unknown Email",
				Target:         "/forgot",
				RequestMethod:  http.MethodPost,
				RequestHeaders: header,
				RequestBody: url.Values{
					"action": {"user"},
					"email":  {"unknown@email"},
				}.Encode(),
				WantStatus: http.StatusOK,
				WantBody: sentBody(weblogin.ForgotPageData{
					Title:     app.Cfg.App.Name,
					EmailFrom: app.Cfg.SMTP.User,
				}),
			},
		*/
	}

	// Test the handler using the utility function.
	webhandler.HandlerTestWithCases(t, app.UserHandler, tests)
}
