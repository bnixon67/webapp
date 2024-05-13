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

func confirmRequestBody(data webauth.ConfirmRequestPageData) string {
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

func sentConfirmRequestBody(data webauth.ConfirmRequestPageData) string {
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

func TestConfirmRequestHandlerGet(t *testing.T) {
	app := AppForTest(t)

	tests := []webhandler.TestCase{
		{
			Name:          "validRequest",
			Target:        "/confirm_request",
			RequestMethod: http.MethodGet,
			WantStatus:    http.StatusOK,
			WantBody: confirmRequestBody(webauth.ConfirmRequestPageData{
				CommonData: webauth.CommonData{
					Title: app.Cfg.App.Name,
				},
			}),
		},
		{
			Name:          "invalidMethod",
			Target:        "/confirm_request",
			RequestMethod: http.MethodPatch,
			WantStatus:    http.StatusMethodNotAllowed,
			WantBody:      "Error: Method Not Allowed\n",
		},
	}

	// Test the handler using the utility function.
	webhandler.TestHandler(t, app.ConfirmRequestHandlerGet, tests)
}

func TestConfirmRequestHandlerPost(t *testing.T) {
	app := AppForTest(t)

	header := http.Header{"Content-Type": {"application/x-www-form-urlencoded"}}
	tests := []webhandler.TestCase{
		{
			Name:          "invalidMethod",
			Target:        "/confirm_request",
			RequestMethod: http.MethodPatch,
			WantStatus:    http.StatusMethodNotAllowed,
			WantBody:      "Error: Method Not Allowed\n",
		},
		{
			Name:           "missingEmail",
			Target:         "/confirm_request",
			RequestMethod:  http.MethodPost,
			RequestHeaders: header,
			WantStatus:     http.StatusOK,
			WantBody: confirmRequestBody(webauth.ConfirmRequestPageData{
				CommonData: webauth.CommonData{
					Title: app.Cfg.App.Name,
				},
				Message: webauth.MsgMissingEmail,
			}),
		},
		{
			Name:           "unknownEmail",
			Target:         "/confirm_request",
			RequestMethod:  http.MethodPost,
			RequestHeaders: header,
			RequestBody: url.Values{
				"email": {"unknown@email"},
			}.Encode(),
			WantStatus: http.StatusSeeOther,
		},
	}

	// Test the handler using the utility function.
	webhandler.TestHandler(t, app.ConfirmRequestHandlerPost, tests)
}
