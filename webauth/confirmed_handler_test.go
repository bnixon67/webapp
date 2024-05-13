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

func confirmedBody(data webauth.ConfirmedData) string {
	// Get path to template file.
	assetDir := assets.AssetPath()
	tmplFile := filepath.Join(assetDir, "tmpl", webauth.ConfirmedTmpl)

	// Parse the HTML template from a file.
	tmpl := template.Must(template.ParseFiles(tmplFile))

	// Create a buffer to store the rendered HTML.
	var body bytes.Buffer

	// Execute the template with the data and write result to the buffer.
	tmpl.Execute(&body, data)

	return body.String()
}

func TestConfirmedHandlerGet(t *testing.T) {
	app := AppForTest(t)

	tests := []webhandler.TestCase{
		{
			Name:          "Valid GET Request",
			Target:        "/confirmed",
			RequestMethod: http.MethodGet,
			WantStatus:    http.StatusOK,
			WantBody: confirmedBody(webauth.ConfirmedData{
				CommonData: webauth.CommonData{
					Title: app.Cfg.App.Name,
				},
			}),
		},
		{
			Name:          "Invalid Method",
			Target:        "/confirm",
			RequestMethod: http.MethodPost,
			WantStatus:    http.StatusMethodNotAllowed,
			WantBody:      "Error: Method Not Allowed\n",
		},
	}

	// Test the handler using the utility function.
	webhandler.TestHandler(t, app.ConfirmedHandlerGet, tests)
}
