// Copyright 2024 Bill Nixon. All rights reserved.
// Use of this source code is governed by the license found in the LICENSE file.

// Package webapp provides common functions and types for web applications.
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

// WebApp encapsulates common web application configurations and state,
// including configuration settings, templates, and build information.
type WebApp struct {
	Config                           // Provides embedded AppConfig.
	Tmpl          *template.Template // Tmpl holds parsed templates.
	BuildDateTime time.Time          // Time executable last modified.
}

// String returns a string representation of WebApp.
func (app *WebApp) String() string {
	if app == nil {
		return "<nil>"
	}

	return fmt.Sprintf("%+v", *app)
}

// Option defines a function type for configuring a WebApp instance,
// adhering to the functional options pattern.
type Option func(*WebApp)

// WithName creates an Option to set the name of the WebApp.
func WithName(name string) Option {
	return func(app *WebApp) {
		app.Config.App.Name = name
	}
}

// WithTemplate creates an Option to set the template of the WebApp.
func WithTemplate(tmpl *template.Template) Option {
	return func(app *WebApp) {
		app.Tmpl = tmpl
	}
}

// New creates a new WebApp instance with the provided options,
// initializing its BuildDateTime to the executable's modification time.
//
// It returns an error if the name is not specified or if it encounters
// issues determining the build time.
func New(opts ...Option) (*WebApp, error) {
	dt, err := ExecutableModTime()
	if err != nil {
		return nil, fmt.Errorf("failed to get build time: %s", err)
	}

	app := &WebApp{BuildDateTime: dt}
	for _, opt := range opts {
		opt(app)
	}

	// Ensure AppName is set.
	if app.Config.App.Name == "" {
		return nil, errors.New("missing Name")
	}

	logIfDebug(app)

	return app, nil
}

// logIfDebug logs info about the WebApp instance if log level is debug.
func logIfDebug(app *WebApp) {
	if slog.Default().Enabled(nil, slog.LevelDebug) {
		tmplNames := strings.Join(webutil.TemplateNames(app.Tmpl), ", ")
		slog.Debug("WebApp creation details",
			slog.String("name", app.Config.App.Name),
			slog.String("templates", tmplNames),
			slog.Time("buildDateTime", app.BuildDateTime))
	}
}

// ExecutableModTime returns the modification time of the current executable.
func ExecutableModTime() (time.Time, error) {
	execPath, err := os.Executable()
	if err != nil {
		return time.Time{}, fmt.Errorf("failed to get executable: %w", err)
	}

	fileInfo, err := os.Stat(execPath)
	if err != nil {
		return time.Time{}, fmt.Errorf("failed to stat exec: %w", err)
	}

	return fileInfo.ModTime(), nil
}
