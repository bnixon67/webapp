// Copyright 2023 Bill Nixon. All rights reserved.
// Use of this source code is governed by the license found in the LICENSE file.

// Package webhandler provides handlers, middleware, and utilities for web applications.
// It simplifies common tasks, enhances request processing, and includes features like request logging, unique request IDs, and HTML template rendering.
package webhandler

import (
	"errors"
	"fmt"
	"html/template"
	"log/slog"
	"time"
)

// Handler represents a web handler with configuration options.
type Handler struct {
	AppName       string             // AppName is the application's name.
	Tmpl          *template.Template // Tmpl stores parsed templates.
	BuildDateTime time.Time          // BuildDateTime is the executable's modification time.
}

// Option is a function type used to apply configuration options to a Handler.
// This follows the Option pattern from https://commandcenter.blogspot.com/2014/01/self-referential-functions-and-design.html and elsewhere.
type Option func(*Handler)

// WithAppName returns an Option to set the AppName of a Handler.
func WithAppName(appName string) Option {
	return func(h *Handler) {
		h.AppName = appName
	}
}

// WithTemplate returns an Option to set the Tmpl of a Handler.
func WithTemplate(tmpl *template.Template) Option {
	return func(h *Handler) {
		h.Tmpl = tmpl
	}
}

// New creates a new Handler with the given options and returns it.
// It returns an error if no AppName is provided through the options or other error occurs.
// The BuildDateTime is also set.
func New(opts ...Option) (*Handler, error) {
	// Retrieve the executable's modification time.
	dt, err := ExecutableModTime()
	if err != nil {
		return nil, fmt.Errorf("failed to get executable modification time: %v", err)
	}

	h := &Handler{
		BuildDateTime: dt,
	}

	// Apply configuration options to the Handler.
	for _, opt := range opts {
		opt(h)
	}

	// Ensure AppName is set.
	if h.AppName == "" {
		return nil, errors.New("AppName is required")
	}

	slog.Debug("created new handler",
		slog.Group("handler",
			slog.String("AppName", h.AppName),
			slog.Time("BuildDateTime", h.BuildDateTime),
		),
	)

	return h, nil
}
