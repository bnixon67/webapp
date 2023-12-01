// Copyright 2023 Bill Nixon. All rights reserved.
// Use of this source code is governed by the license found in the LICENSE file.

package webapp_test

import (
	"bytes"
	"html/template"
	"net/http"
	"testing"

	"github.com/bnixon67/webapp/webapp"
	"github.com/bnixon67/webapp/webhandler"
	"github.com/bnixon67/webapp/webutil"
)

// headersBody is a utility function that renders an HTML template for the given headers.
func headersBody(headers http.Header) string {
	// Parse the HTML template from a file.
	tmpl := template.Must(template.ParseFiles("../assets/tmpl/headers.html"))

	// Create a buffer to store the rendered HTML.
	var body bytes.Buffer

	// Define the template data.
	data := webapp.HeadersPageData{
		Title:   "Request Headers",
		Headers: webapp.SortedHeaders(headers),
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

	tests := []webhandler.TestCase{
		{
			Name:           "Valid GET Request with no headers",
			RequestMethod:  http.MethodGet,
			RequestHeaders: noHeaders,
			WantStatus:     http.StatusOK,
			WantBody:       headersBody(noHeaders),
		},
		{
			Name:           "Valid GET Request with typical headers",
			RequestMethod:  http.MethodGet,
			RequestHeaders: typicalHeaders,
			WantStatus:     http.StatusOK,
			WantBody:       headersBody(typicalHeaders),
		},
		{
			Name:           "Valid GET Request with multiple header values",
			RequestMethod:  http.MethodGet,
			RequestHeaders: multiHeaders,
			WantStatus:     http.StatusOK,
			WantBody:       headersBody(multiHeaders),
		},
		{
			Name:          "Invalid POST Request",
			RequestMethod: http.MethodPost,
			WantStatus:    http.StatusMethodNotAllowed,
			WantBody:      "POST Method Not Allowed\n",
		},
	}

	// Define the custom function
	funcMap := template.FuncMap{
		"ToTimeZone": webutil.ToTimeZone,
	}

	// Initialize templates
	tmpls, err := template.New("html").Funcs(funcMap).ParseGlob("../assets/tmpl/*.html")
	if err != nil {
		t.Fatalf("could not create initialize templates: %v", err)
	}

	// Create a web app instance for testing.
	app, err := webapp.New(webapp.WithAppName("Test App"), webapp.WithTemplate(tmpls))
	if err != nil {
		t.Fatalf("could not create web handler: %v", err)
	}

	// Test the handler using the utility function.
	webhandler.HandlerTestWithCases(t, app.HeadersHandler, tests)
}
