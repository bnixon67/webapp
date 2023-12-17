// Copyright 2023 Bill Nixon. All rights reserved.
// Use of this source code is governed by the license found in the LICENSE file.

// Package webapp provides a common functions for web applications.
package webapp

import (
	"errors"
	"fmt"
	"html/template"
	"log/slog"
	"os"
	"strings"
	"time"

	"github.com/bnixon67/webapp/webutil"
)

// WebApp contains common variables used across the web application.
// This is used in handlers or other functions to avoid global variables.
type WebApp struct {
	Name          string             // Name is the name of the app.
	Tmpl          *template.Template // Tmpl stores parsed templates.
	BuildDateTime time.Time          // BuildDateTime is the executable's modification time.
}

// String returns a string representation of the webapp.
func (webapp *WebApp) String() string {
	if webapp == nil {
		return fmt.Sprintf("%v", nil)
	}

	return fmt.Sprintf("%+v", *webapp)
}

// Option is a function type used to apply configuration options to a WebApp.
// This follows the Option pattern from https://commandcenter.blogspot.com/2014/01/self-referential-functions-and-design.html and elsewhere.
type Option func(*WebApp)

// WithName returns an Option to set the Name of a WebApp.
func WithName(name string) Option {
	return func(webapp *WebApp) {
		webapp.Name = name
	}
}

// WithTemplate returns an Option to set the Tmpl of a WebApp.
func WithTemplate(tmpl *template.Template) Option {
	return func(webapp *WebApp) {
		webapp.Tmpl = tmpl
	}
}

// New return a new WebApp with the given options and BuildDateTime.
// An error is returned if Name is not provided or other errors occur.
func New(opts ...Option) (*WebApp, error) {
	// Retrieve the executable's modification time.
	dt, err := ExecutableModTime()
	if err != nil {
		return nil, fmt.Errorf("failed to get build date time: %v", err)
	}

	// Create webapp with build date time.
	webapp := &WebApp{
		BuildDateTime: dt,
	}

	// Apply configuration options.
	for _, opt := range opts {
		opt(webapp)
	}

	// Ensure AppName is set.
	if webapp.Name == "" {
		return nil, errors.New("missing Name")
	}

	tmplNames := strings.Join(webutil.TemplateNames(webapp.Tmpl), ", ")
	slog.Debug("created webapp",
		slog.Group("webapp",
			slog.String("Name", webapp.Name),
			slog.String("Templates", tmplNames),
			slog.Time("BuildDateTime", webapp.BuildDateTime),
		),
	)

	return webapp, nil
}

// ExecutableModTime returns the modification time of the current executable.
func ExecutableModTime() (time.Time, error) {
	var nilTime time.Time

	// Get path of current executable.
	execPath, err := os.Executable()
	if err != nil {
		return nilTime, fmt.Errorf("failed to get exec path: %w", err)
	}

	// Get file information of the executable.
	fileInfo, err := os.Stat(execPath)
	if err != nil {
		return nilTime, fmt.Errorf("failed to stat exec: %w", err)
	}

	// Return modification time.
	return fileInfo.ModTime(), nil
}
