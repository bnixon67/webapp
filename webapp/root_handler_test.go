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

// rootBody is a utility function that renders an HTML template for the given hedata.
func rootBody(t *testing.T, data webapp.RootPageData, funcMap template.FuncMap) string {
	tmplName := webapp.RootPageName

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

	// Execute template with data and write the result to the buffer.
	tmpl.Execute(&body, data)

	return body.String()
}

func TestRootHandler(t *testing.T) {
	data := webapp.RootPageData{Title: "Test App"}

	// Define the custom functions
	funcMap := template.FuncMap{
		"ToTimeZone": webutil.ToTimeZone,
		"Join":       webutil.Join,
	}

	tests := []webhandler.TestCase{
		{
			Name:          "Valid GET Request",
			Target:        "/",
			RequestMethod: http.MethodGet,
			WantStatus:    http.StatusOK,
			WantBody:      rootBody(t, data, funcMap),
		},
		{
			Name:          "Inavlid Path",
			Target:        "/invalid",
			RequestMethod: http.MethodGet,
			WantStatus:    http.StatusNotFound,
			WantBody:      "404 page not found\n",
		},
		{
			Name:          "Invalid POST Request",
			Target:        "/",
			RequestMethod: http.MethodPost,
			WantStatus:    http.StatusMethodNotAllowed,
			WantBody:      "POST Method Not Allowed\n",
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
	webhandler.HandlerTestWithCases(t, app.RootHandler, tests)
}
