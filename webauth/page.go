// Copyright 2024 Bill Nixon. All rights reserved.
// Use of this source code is governed by the license found in the LICENSE file.

package webauth

import (
	"log/slog"
	"net/http"

	"github.com/bnixon67/webapp/webutil"
)

// PageData is an interface that all page data structs must implement.
type PageData interface {
	SetDefaultTitle(appName string)
}

// CommonData holds common fields for page data.
type CommonData struct {
	Title string
}

// SetDefaultTitle ensures that the Title of CommonPageData is not empty.
// If Title is empty, this method sets it to the value of appName.
func (c *CommonData) SetDefaultTitle(appName string) {
	if c.Title == "" {
		c.Title = appName
	}
}

// RenderPage renders a web page using the specified template and data.
//
// If the page cannot be rendered, http.StatusInternalServerError is
// set and the caller should ensure no further writes are done to w.
func (app *AuthApp) RenderPage(w http.ResponseWriter, logger *slog.Logger, templateName string, data PageData) {
	data.SetDefaultTitle(app.Cfg.App.Name)

	err := webutil.RenderTemplate(app.Tmpl, w, templateName, data)
	if err != nil {
		logger.Error("unable to render template", "err", err)
		webutil.HttpError(w, http.StatusInternalServerError)
	}
}
