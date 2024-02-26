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

func confirmBody(data webauth.ConfirmData) string {
	// Get path to template file.
	assetDir := assets.AssetPath()
	tmplFile := filepath.Join(assetDir, "tmpl", webauth.ConfirmTmpl)

	// Parse the HTML template from a file.
	tmpl := template.Must(template.ParseFiles(tmplFile))

	// Create a buffer to store the rendered HTML.
	var body bytes.Buffer

	// Execute the template with the data and write result to the buffer.
	tmpl.Execute(&body, data)

	return body.String()
}

func sentConfirmBody(data webauth.ConfirmData) string {
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

func TestConfirmHandlerGet(t *testing.T) {
	app := AppForTest(t)

	tests := []webhandler.TestCase{
		{
			Name:          "Valid GET Request",
			Target:        "/confirm",
			RequestMethod: http.MethodGet,
			WantStatus:    http.StatusOK,
			WantBody: confirmBody(webauth.ConfirmData{
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
	webhandler.HandlerTestWithCases(t, app.ConfirmHandlerGet, tests)
}

func TestConfirmHandlerPost(t *testing.T) {
	app := AppForTest(t)

	header := http.Header{
		"Content-Type": {"application/x-www-form-urlencoded"},
	}

	utoken, err := app.DB.CreateConfirmEmailToken("unconfirmed")
	if err != nil {
		t.Fatalf("could not create confirm email token")
	}

	_, err = app.DB.Exec(
		"UPDATE users SET confirmed = 0 WHERE username = ?",
		"unconfirmed")
	if err != nil {
		t.Fatalf("could not unconfirm user")
	}

	expiredToken, err := app.DB.CreateConfirmEmailToken("expired")
	if err != nil {
		t.Fatalf("could not create expired confirm email token")
	}

	_, err = app.DB.Exec(
		"UPDATE tokens SET expires = NOW() WHERE username = ?",
		"expired")
	if err != nil {
		t.Fatalf("could not expire token: %v", err)
	}

	tests := []webhandler.TestCase{
		{
			Name:          "Invalid Method",
			Target:        "/confirm",
			RequestMethod: http.MethodGet,
			WantStatus:    http.StatusMethodNotAllowed,
			WantBody:      "Error: Method Not Allowed\n",
		},
		{
			Name:           "Missing Token",
			Target:         "/confirm",
			RequestMethod:  http.MethodPost,
			RequestHeaders: header,
			RequestBody:    url.Values{}.Encode(),
			WantStatus:     http.StatusOK,
			WantBody: confirmBody(webauth.ConfirmData{
				CommonData: webauth.CommonData{
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
			WantBody: confirmBody(webauth.ConfirmData{
				CommonData: webauth.CommonData{
					Title: app.Cfg.App.Name,
				},
				Message: webauth.MsgInvalidConfirmToken,
			}),
		},
		{
			Name:           "Expired Token",
			Target:         "/confirm",
			RequestMethod:  http.MethodPost,
			RequestHeaders: header,
			RequestBody:    url.Values{"ctoken": {expiredToken.Value}}.Encode(),
			WantStatus:     http.StatusOK,
			WantBody: confirmBody(webauth.ConfirmData{
				CommonData: webauth.CommonData{
					Title: app.Cfg.App.Name,
				},
				Message: webauth.MsgExpiredConfirmToken,
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
	}

	// Test the handler using the utility function.
	webhandler.HandlerTestWithCases(t, app.ConfirmHandlerPost, tests)
}
