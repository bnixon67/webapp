// Copyright 2023 Bill Nixon. All rights reserved.
// Use of this source code is governed by the license found in the LICENSE file.

package webhandler_test

import (
	"bytes"
	"html/template"
	"net/http"
	"testing"

	"github.com/bnixon67/webapp/webhandler"
)

// headersBody is a utility function that renders an HTML template for the given headers.
func headersBody(headers http.Header) string {
	// Parse the HTML template from a file.
	tmpl := template.Must(template.ParseFiles("../assets/tmpl/headers.html"))

	// Create a buffer to store the rendered HTML.
	var body bytes.Buffer

	// Define the template data.
	data := webhandler.HeadersPageData{
		Title:   "Request Headers",
		Headers: webhandler.SortedHeaders(headers),
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

	tests := []TestCase{
		{
			name:           "Valid GET Request with no headers",
			requestMethod:  http.MethodGet,
			requestHeaders: noHeaders,
			wantStatus:     http.StatusOK,
			wantBody:       headersBody(noHeaders),
		},
		{
			name:           "Valid GET Request with typical headers",
			requestMethod:  http.MethodGet,
			requestHeaders: typicalHeaders,
			wantStatus:     http.StatusOK,
			wantBody:       headersBody(typicalHeaders),
		},
		{
			name:           "Valid GET Request with multiple header values",
			requestMethod:  http.MethodGet,
			requestHeaders: multiHeaders,
			wantStatus:     http.StatusOK,
			wantBody:       headersBody(multiHeaders),
		},
		{
			name:          "Invalid POST Request",
			requestMethod: http.MethodPost,
			wantStatus:    http.StatusMethodNotAllowed,
			wantBody:      "POST Method Not Allowed\n",
		},
	}

	// Initialize templates
	tmpls, err := template.New("html").ParseGlob("../assets/tmpl/*.html")
	if err != nil {
		t.Fatalf("could not create initialize templates: %v", err)
	}

	// Create a web handler instance for testing.
	handler, err := webhandler.New(webhandler.WithAppName("Test App"), webhandler.WithTemplate(tmpls))
	if err != nil {
		t.Fatalf("could not create web handler: %v", err)
	}

	// Test the handler using the utility function.
	HandlerTestWithCases(t, handler.HeadersHandler, tests)
}
