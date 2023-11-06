// Copyright 2023 Bill Nixon. All rights reserved.
// Use of this source code is governed by the license found in the LICENSE file.

package weblogin_test

import (
	"bytes"
	"net/http"
	"net/url"
	"testing"
	"text/template"

	"github.com/bnixon67/webapp/webhandler"
	"github.com/bnixon67/webapp/weblogin"
)

func forgotBody(data weblogin.ForgotPageData) string {
	// Parse the HTML template from a file.
	tmpl := template.Must(template.ParseFiles("tmpl/forgot.html"))

	// Create a buffer to store the rendered HTML.
	var body bytes.Buffer

	// Execute the template with the data and write the result to the buffer.
	tmpl.Execute(&body, data)

	return body.String()
}

func sentBody(data weblogin.ForgotPageData) string {
	// Parse the HTML template from a file.
	tmpl := template.Must(template.ParseFiles("tmpl/forgot_sent.html"))

	// Create a buffer to store the rendered HTML.
	var body bytes.Buffer

	// Execute the template with the data and write the result to the buffer.
	tmpl.Execute(&body, data)

	return body.String()
}

func TestForgotHandler(t *testing.T) {
	app := AppForTest(t)

	header := http.Header{"Content-Type": {"application/x-www-form-urlencoded"}}
	tests := []webhandler.TestCase{
		{
			Name:          "Valid GET Request",
			Target:        "/forgot",
			RequestMethod: http.MethodGet,
			WantStatus:    http.StatusOK,
			WantBody: forgotBody(weblogin.ForgotPageData{
				Title: app.Cfg.Title,
			}),
		},
		{
			Name:          "Invalid Method",
			Target:        "/forgot",
			RequestMethod: http.MethodPatch,
			WantStatus:    http.StatusMethodNotAllowed,
			WantBody:      "PATCH Method Not Allowed\n",
		},
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
	}

	// Test the handler using the utility function.
	webhandler.HandlerTestWithCases(t, app.ForgotHandler, tests)
}
