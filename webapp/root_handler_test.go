// Copyright 2024 Bill Nixon. All rights reserved.
// Use of this source code is governed by the license found in the LICENSE file.

package webapp_test

import (
	"net/http"
	"testing"

	"github.com/bnixon67/webapp/webapp"
	"github.com/bnixon67/webapp/webhandler"
	"github.com/bnixon67/webapp/webutil"
)

func TestRootHandler(t *testing.T) {
	data := webapp.RootPageData{Title: "Test App"}

	app := AppForTest(t)

	body := webutil.RenderTemplateForTest(t, app.Tmpl, webapp.RootPageName, data)

	tests := []webhandler.TestCase{
		{
			Name:          "Valid GET Request",
			Target:        "/",
			RequestMethod: http.MethodGet,
			WantStatus:    http.StatusOK,
			WantBody:      body,
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
			WantBody:      "Error: Method Not Allowed\n",
		},
	}

	webhandler.TestHandler(t, app.RootHandlerGet, tests)
}
