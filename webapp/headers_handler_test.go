// Copyright 2023 Bill Nixon. All rights reserved.
// Use of this source code is governed by the license found in the LICENSE file.

package webapp_test

import (
	"bytes"
	"html/template"
	"net/http"
	"path/filepath"
	"testing"

	"github.com/bnixon67/webapp/assets"
	"github.com/bnixon67/webapp/webapp"
	"github.com/bnixon67/webapp/webhandler"
	"github.com/bnixon67/webapp/webutil"
)

// headersBody is a utility function that renders an HTML template for the given headers.
func headersBody(t *testing.T, headers http.Header, funcMap template.FuncMap) string {
	tmplName := "headers.html"

	// Directly include the name of the template in New for clarity.
	tmpl := template.New(tmplName).Funcs(funcMap)

	// Get path to template file.
	assetDir := assets.AssetPath()
	tmplFile := filepath.Join(assetDir, "tmpl", tmplName)

	// Parse the template file, checking for errors.
	tmpl, err := tmpl.ParseFiles(tmplFile)
	if err != nil {
		t.Fatalf("could not parse template file '%s': %v", tmplFile, err)
	}

	// Create a buffer to store the rendered HTML.
	var body bytes.Buffer

	// Define the template data.
	data := webapp.HeadersPageData{
		Title:   "Request Headers",
		Headers: webapp.SortHeaders(headers),
	}

	// Execute the template with the data and write the result to the buffer.
	tmpl.Execute(&body, data)

	return body.String()
}

func TestGetHeaders(t *testing.T) {
	noHeaders := http.Header{}

	typicalHeaders := http.Header{
		"Content-Type":    {"application/json"},
		"X-Custom-Header": {"value"},
		"Accept-Encoding": {"gzip"},
	}

	multiHeaders := http.Header{
		"Content-Type":    {"application/json"},
		"X-Custom-Header": {"value1", "value2"},
		"Accept-Encoding": {"gzip"},
	}

	// Define the custom functions
	funcMap := template.FuncMap{
		"ToTimeZone": webutil.ToTimeZone,
		"Join":       webutil.Join,
	}

	tests := []webhandler.TestCase{
		{
			Name:           "Valid GET Request with no headers",
			RequestMethod:  http.MethodGet,
			RequestHeaders: noHeaders,
			WantStatus:     http.StatusOK,
			WantBody:       headersBody(t, noHeaders, funcMap),
		},
		{
			Name:           "Valid GET Request with typical headers",
			RequestMethod:  http.MethodGet,
			RequestHeaders: typicalHeaders,
			WantStatus:     http.StatusOK,
			WantBody:       headersBody(t, typicalHeaders, funcMap),
		},
		{
			Name:           "Valid GET Request with multiple header values",
			RequestMethod:  http.MethodGet,
			RequestHeaders: multiHeaders,
			WantStatus:     http.StatusOK,
			WantBody:       headersBody(t, multiHeaders, funcMap),
		},
		{
			Name:          "Invalid POST Request",
			RequestMethod: http.MethodPost,
			WantStatus:    http.StatusMethodNotAllowed,
			WantBody:      "Error: Method Not Allowed\n",
		},
	}

	// Initialize templates
	tmpls, err := template.New("html").Funcs(funcMap).ParseGlob("../assets/tmpl/*.html")
	if err != nil {
		t.Fatalf("could not create initialize templates: %v", err)
	}

	// Create a web app instance for testing.
	app, err := webapp.New(webapp.WithName("Test App"), webapp.WithTemplate(tmpls))
	if err != nil {
		t.Fatalf("could not create web handler: %v", err)
	}

	// Test the handler using the utility function.
	webhandler.HandlerTestWithCases(t, app.HeadersHandlerGet, tests)
}
