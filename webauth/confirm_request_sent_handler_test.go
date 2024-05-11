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

func confirmRequestSentBody(data webauth.ConfirmRequestSentData) string {
	assetDir := assets.AssetPath()
	tmplFile := filepath.Join(assetDir, "tmpl", webauth.ConfirmRequestSentTmpl)

	tmpl := template.Must(template.ParseFiles(tmplFile))

	var body bytes.Buffer
	tmpl.Execute(&body, data)

	return body.String()
}

func TestConfirmRequestSentHandlerGet(t *testing.T) {
	app := AppForTest(t)

	tests := []webhandler.TestCase{
		{
			Name:          "Valid GET Request",
			Target:        "/confirm_request_sent",
			RequestMethod: http.MethodGet,
			WantStatus:    http.StatusOK,
			WantBody: confirmRequestSentBody(
				webauth.ConfirmRequestSentData{
					CommonData: webauth.CommonData{
						Title: app.Cfg.App.Name,
					},
					EmailFrom: app.Cfg.SMTP.Username,
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
	webhandler.HandlerTestWithCases(t, app.ConfirmRequestSentHandlerGet, tests)
}
