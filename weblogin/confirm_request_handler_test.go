// Copyright 2024 Bill Nixon. All rights reserved.
// Use of this source code is governed by the license found in the LICENSE file.

package weblogin_test

import (
	"bytes"
	"net/http"
	"net/url"
	"path/filepath"
	"testing"
	"text/template"

	"github.com/bnixon67/webapp/assets"
	"github.com/bnixon67/webapp/webhandler"
	"github.com/bnixon67/webapp/weblogin"
)

func confirmRequestBody(data weblogin.ConfirmRequestPageData) string {
	// Get path to template file.
	assetDir := assets.AssetPath()
	tmplFile := filepath.Join(assetDir, "tmpl", "confirm_request.html")

	// Parse the HTML template from a file.
	tmpl := template.Must(template.ParseFiles(tmplFile))

	// Create a buffer to store the rendered HTML.
	var body bytes.Buffer

	// Execute the template with the data and write result to the buffer.
	tmpl.Execute(&body, data)

	return body.String()
}

func sentConfirmRequestBody(data weblogin.ConfirmRequestPageData) string {
	// Get path to template file.
	assetDir := assets.AssetPath()
	tmplFile := filepath.Join(assetDir, "tmpl", "confirm_request_sent.html")

	// Parse the HTML template from a file.
	tmpl := template.Must(template.ParseFiles(tmplFile))

	// Create a buffer to store the rendered HTML.
	var body bytes.Buffer

	// Execute the template with the data and write the result to the buffer.
	tmpl.Execute(&body, data)

	return body.String()
}

func TestConfirmRequestHandler(t *testing.T) {
	app := AppForTest(t)

	header := http.Header{"Content-Type": {"application/x-www-form-urlencoded"}}
	tests := []webhandler.TestCase{
		{
			Name:          "Valid GET Request",
			Target:        "/confirm_request",
			RequestMethod: http.MethodGet,
			WantStatus:    http.StatusOK,
			WantBody: confirmRequestBody(weblogin.ConfirmRequestPageData{
				Title: app.Cfg.App.Name,
			}),
		},
		{
			Name:          "Invalid Method",
			Target:        "/confirm_request",
			RequestMethod: http.MethodPatch,
			WantStatus:    http.StatusMethodNotAllowed,
			WantBody:      "PATCH Method Not Allowed\n",
		},
		{
			Name:           "Missing Email",
			Target:         "/confirm_request",
			RequestMethod:  http.MethodPost,
			RequestHeaders: header,
			WantStatus:     http.StatusOK,
			WantBody: confirmRequestBody(weblogin.ConfirmRequestPageData{
				Title:   app.Cfg.App.Name,
				Message: weblogin.MsgMissingEmail,
			}),
		},
		{
			Name:           "Unknown Email",
			Target:         "/confirm_request",
			RequestMethod:  http.MethodPost,
			RequestHeaders: header,
			RequestBody: url.Values{
				"email": {"unknown@email"},
			}.Encode(),
			WantStatus: http.StatusOK,
			WantBody: sentConfirmRequestBody(weblogin.ConfirmRequestPageData{
				Title:     app.Cfg.App.Name,
				EmailFrom: app.Cfg.SMTP.User,
			}),
		},
		{
			Name:           "Unknown Email",
			Target:         "/confirm_request",
			RequestMethod:  http.MethodPost,
			RequestHeaders: header,
			RequestBody: url.Values{
				"email": {"test@email"},
			}.Encode(),
			WantStatus: http.StatusOK,
			WantBody: sentConfirmRequestBody(weblogin.ConfirmRequestPageData{
				Title:     app.Cfg.App.Name,
				EmailFrom: app.Cfg.SMTP.User,
			}),
		},
	}

	// Test the handler using the utility function.
	webhandler.HandlerTestWithCases(t, app.ConfirmRequestHandler, tests)
}
