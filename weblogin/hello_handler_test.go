// Copyright 2023 Bill Nixon. All rights reserved.
// Use of this source code is governed by the license found in the LICENSE file.

package weblogin_test

import (
	"bytes"
	"net/http"
	"testing"
	"text/template"

	"github.com/bnixon67/webapp/webhandler"
	"github.com/bnixon67/webapp/weblogin"
)

func helloBody(data weblogin.HelloPageData) string {
	// Parse the HTML template from a file.
	tmpl := template.Must(template.ParseFiles("tmpl/hello.html"))

	// Create a buffer to store the rendered HTML.
	var body bytes.Buffer

	// Execute the template with the data and write the result to the buffer.
	tmpl.Execute(&body, data)

	return body.String()
}

func TestHelloHandler(t *testing.T) {
	app := AppForTest(t)

	// TODO: better way to define a test user
	token, err := app.LoginUser("test", "password")
	if err != nil {
		t.Errorf("could not login user to get session token")
	}
	user, err := weblogin.GetUserForSessionToken(app.DB, token.Value)
	if err != nil {
		t.Errorf("could not get user")
	}

	tests := []webhandler.TestCase{
		{
			Name:          "Invalid Method",
			Target:        "/hello",
			RequestMethod: http.MethodPost,
			WantStatus:    http.StatusMethodNotAllowed,
			WantBody:      "POST Method Not Allowed\n",
		},
		{
			Name:          "Valid GET Request without Cookie",
			Target:        "/hello",
			RequestMethod: http.MethodGet,
			WantStatus:    http.StatusOK,
			WantBody: helloBody(weblogin.HelloPageData{
				Title: app.Cfg.Title,
			}),
		},
		{
			Name:          "Valid GET Request with Bad Session Token",
			Target:        "/hello",
			RequestMethod: http.MethodGet,
			RequestCookies: []http.Cookie{
				{Name: weblogin.SessionTokenCookieName, Value: "foo"},
			},
			WantStatus: http.StatusOK,
			WantBody: helloBody(weblogin.HelloPageData{
				Title: app.Cfg.Title,
			}),
		},
		{
			Name:          "Valid GET Request with Good Session Token",
			Target:        "/hello",
			RequestMethod: http.MethodGet,
			RequestCookies: []http.Cookie{
				{Name: weblogin.SessionTokenCookieName, Value: token.Value},
			},
			WantStatus: http.StatusOK,
			WantBody: helloBody(weblogin.HelloPageData{
				Title: app.Cfg.Title, User: user,
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
					Title:   app.Cfg.Title,
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
					Title:   app.Cfg.Title,
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
					Title:   app.Cfg.Title,
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
					Title:     app.Cfg.Title,
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
					Title:     app.Cfg.Title,
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
					Title:     app.Cfg.Title,
					EmailFrom: app.Cfg.SMTP.User,
				}),
			},
		*/
	}

	// Test the handler using the utility function.
	webhandler.HandlerTestWithCases(t, app.HelloHandler, tests)
}
