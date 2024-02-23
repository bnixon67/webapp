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

func confirmBody(data webauth.ConfirmPageData) string {
	// Get path to template file.
	assetDir := assets.AssetPath()
	tmplFile := filepath.Join(assetDir, "tmpl", "confirm.html")

	// Parse the HTML template from a file.
	tmpl := template.Must(template.ParseFiles(tmplFile))

	// Create a buffer to store the rendered HTML.
	var body bytes.Buffer

	// Execute the template with the data and write result to the buffer.
	tmpl.Execute(&body, data)

	return body.String()
}

func sentConfirmBody(data webauth.ConfirmPageData) string {
	// Get path to template file.
	assetDir := assets.AssetPath()
	tmplFile := filepath.Join(assetDir, "tmpl", "confirm_request.html")

	// Parse the HTML template from a file.
	tmpl := template.Must(template.ParseFiles(tmplFile))

	// Create a buffer to store the rendered HTML.
	var body bytes.Buffer

	// Execute the template with the data and write the result to the buffer.
	tmpl.Execute(&body, data)

	return body.String()
}

func TestConfirmHandler(t *testing.T) {
	app := AppForTest(t)

	header := http.Header{
		"Content-Type": {"application/x-www-form-urlencoded"},
	}

	ctoken, err := app.DB.CreateConfirmEmailToken("confirmed")
	if err != nil {
		t.Fatalf("could not create confirm email token")
	}

	utoken, err := app.DB.CreateConfirmEmailToken("unconfirmed")
	if err != nil {
		t.Fatalf("could not create confirm email token")
	}

	const qry = "UPDATE users SET confirmed = 0 WHERE username = ?"
	_, err = app.DB.Exec(qry, "unconfirmed")
	if err != nil {
		t.Fatalf("could not unconfirm user")
	}

	tests := []webhandler.TestCase{
		{
			Name:          "Valid GET Request",
			Target:        "/confirm",
			RequestMethod: http.MethodGet,
			WantStatus:    http.StatusOK,
			WantBody: confirmBody(webauth.ConfirmPageData{
				CommonPageData: webauth.CommonPageData{
					Title: app.Cfg.App.Name,
				},
			}),
		},
		{
			Name:          "Invalid Method",
			Target:        "/confirm",
			RequestMethod: http.MethodPatch,
			WantStatus:    http.StatusMethodNotAllowed,
			WantBody:      "PATCH Method Not Allowed\n",
		},
		{
			Name:           "Missing Token",
			Target:         "/confirm",
			RequestMethod:  http.MethodPost,
			RequestHeaders: header,
			RequestBody:    url.Values{}.Encode(),
			WantStatus:     http.StatusOK,
			WantBody: confirmBody(webauth.ConfirmPageData{
				CommonPageData: webauth.CommonPageData{
					Title: app.Cfg.App.Name,
				},
				Message: webauth.MsgMissingConfirmToken,
			}),
		},
		{
			Name:           "Invalid Token",
			Target:         "/confirm",
			RequestMethod:  http.MethodPost,
			RequestHeaders: header,
			RequestBody:    url.Values{"ctoken": {"foo"}}.Encode(),
			WantStatus:     http.StatusOK,
			WantBody: confirmBody(webauth.ConfirmPageData{
				CommonPageData: webauth.CommonPageData{
					Title: app.Cfg.App.Name,
				},
				Message:      webauth.MsgInvalidConfirmToken,
				ConfirmToken: "foo",
			}),
		},
		{
			Name:           "Unconfirmed User",
			Target:         "/confirm",
			RequestMethod:  http.MethodPost,
			RequestHeaders: header,
			RequestBody:    url.Values{"ctoken": {utoken.Value}}.Encode(),
			WantStatus:     http.StatusSeeOther,
		},
		{
			Name:           "Confirmed User",
			Target:         "/confirm",
			RequestMethod:  http.MethodPost,
			RequestHeaders: header,
			RequestBody:    url.Values{"ctoken": {ctoken.Value}}.Encode(),
			WantStatus:     http.StatusOK,
			WantBody: confirmBody(webauth.ConfirmPageData{
				CommonPageData: webauth.CommonPageData{
					Title: app.Cfg.App.Name,
				},
				Message: webauth.MsgUserAlreadyConfirmed,
			}),
		},
	}

	// Test the handler using the utility function.
	webhandler.HandlerTestWithCases(t, app.ConfirmHandler, tests)
}
