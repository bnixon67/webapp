// Copyright 2024 Bill Nixon. All rights reserved.
// Use of this source code is governed by the license found in the LICENSE file.

package webauth_test

import (
	"bytes"
	"net/http"
	"net/url"
	"path/filepath"
	"testing"
	"text/template"

	"github.com/bnixon67/webapp/assets"
	"github.com/bnixon67/webapp/webauth"
	"github.com/bnixon67/webapp/webhandler"
)

func forgotBody(data webauth.ForgotPageData) string {
	// Get path to template file.
	assetDir := assets.AssetPath()
	tmplFile := filepath.Join(assetDir, "tmpl", "forgot.html")

	// Parse the HTML template from a file.
	tmpl := template.Must(template.ParseFiles(tmplFile))

	// Create a buffer to store the rendered HTML.
	var body bytes.Buffer

	// Execute the template with the data and write the result to the buffer.
	tmpl.Execute(&body, data)

	return body.String()
}

func sentBody(data webauth.ForgotPageData) string {
	// Get path to template file.
	assetDir := assets.AssetPath()
	tmplFile := filepath.Join(assetDir, "tmpl", "forgot_sent.html")

	// Parse the HTML template from a file.
	tmpl := template.Must(template.ParseFiles(tmplFile))

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
			WantBody: forgotBody(webauth.ForgotPageData{
				CommonData: webauth.CommonData{Title: app.Cfg.App.Name},
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
			WantBody: forgotBody(webauth.ForgotPageData{
				CommonData: webauth.CommonData{Title: app.Cfg.App.Name},
				Message:    webauth.MsgMissingEmail,
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
			WantBody: forgotBody(webauth.ForgotPageData{
				CommonData: webauth.CommonData{Title: app.Cfg.App.Name},
				Message:    webauth.MsgMissingAction,
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
			WantBody: forgotBody(webauth.ForgotPageData{
				CommonData: webauth.CommonData{Title: app.Cfg.App.Name},
				Message:    webauth.MsgInvalidAction,
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
			WantBody: sentBody(webauth.ForgotPageData{
				CommonData: webauth.CommonData{Title: app.Cfg.App.Name},
				EmailFrom:  app.Cfg.SMTP.User,
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
			WantBody: sentBody(webauth.ForgotPageData{
				CommonData: webauth.CommonData{Title: app.Cfg.App.Name},
				EmailFrom:  app.Cfg.SMTP.User,
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
			WantBody: sentBody(webauth.ForgotPageData{
				CommonData: webauth.CommonData{Title: app.Cfg.App.Name},
				EmailFrom:  app.Cfg.SMTP.User,
			}),
		},
	}

	// Test the handler using the utility function.
	webhandler.HandlerTestWithCases(t, app.ForgotHandler, tests)
}
