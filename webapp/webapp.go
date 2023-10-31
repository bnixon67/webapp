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
	"time"
)

// WebApp contains common variables used across various components of the web application.
// This struct can be used as a receiver in handler or other functions.
type WebApp struct {
	AppName       string             // AppName is the application's name.
	Tmpl          *template.Template // Tmpl stores parsed templates.
	BuildDateTime time.Time          // BuildDateTime is the executable's modification time.
}

func (a *WebApp) String() string {
	if a == nil {
		return fmt.Sprintf("%v", nil)
	}

	return fmt.Sprintf("%+v", *a)
}

// Option is a function type used to apply configuration options to a WebApp.
// This follows the Option pattern from https://commandcenter.blogspot.com/2014/01/self-referential-functions-and-design.html and elsewhere.
type Option func(*WebApp)

// WithAppName returns an Option to set the AppName of a WebApp.
func WithAppName(appName string) Option {
	return func(a *WebApp) {
		a.AppName = appName
	}
}

// WithTemplate returns an Option to set the Tmpl of a WebApp.
func WithTemplate(tmpl *template.Template) Option {
	return func(a *WebApp) {
		a.Tmpl = tmpl
	}
}

// New creates a new WebApp with the given options and returns it.
// It returns an error if no AppName is provided through the options or other error occurs.
// The BuildDateTime is also set.
func New(opts ...Option) (*WebApp, error) {
	// Retrieve the executable's modification time.
	dt, err := ExecutableModTime()
	if err != nil {
		return nil, fmt.Errorf("failed to get executable modification time: %v", err)
	}

	a := &WebApp{
		BuildDateTime: dt,
	}

	// Apply configuration options to the Handler.
	for _, opt := range opts {
		opt(a)
	}

	// Ensure AppName is set.
	if a.AppName == "" {
		return nil, errors.New("AppName is required")
	}

	slog.Debug("created new webapp", "webapp", a)

	return a, nil
}

// ExecutableModTime returns the modification time of the executable file.
func ExecutableModTime() (time.Time, error) {
	// Get path of the current executable.
	execPath, err := os.Executable()
	if err != nil {
		return time.Time{}, fmt.Errorf("error getting executable path: %w", err)
	}

	// Get file information of the executable.
	fileInfo, err := os.Stat(execPath)
	if err != nil {
		return time.Time{}, fmt.Errorf("error getting file info: %w", err)
	}

	// Return modification time.
	return fileInfo.ModTime(), nil
}
