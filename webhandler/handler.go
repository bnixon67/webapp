// Package webhandler provides utilities for working with web handlers.
package webhandler

import (
	"errors"
	"html/template"
	"log/slog"
)

// Handler represents a web handler with configuration options.
type Handler struct {
	AppName string             // AppName is the name of the application.
	Tmpl    *template.Template // Tmpl holds parsed templates.
}

// Option is a function type used to apply configuration options to a Handler.
type Option func(*Handler)

// WithAppName returns an Option to set the AppName of a Handler.
func WithAppName(appName string) Option {
	return func(h *Handler) {
		h.AppName = appName
	}
}

// WithTemplate returns an Option to set the AppName of a Handler.
func WithTemplate(tmpl *template.Template) Option {
	return func(h *Handler) {
		h.Tmpl = tmpl
	}
}

// New creates a new Handler with the given options and returns it.
// It returns an error if no AppName is provided through the options.
func New(opts ...Option) (*Handler, error) {
	h := &Handler{}

	// Apply configuration options to the Handler.
	for _, opt := range opts {
		opt(h)
	}

	// Ensure AppName is set.
	if h.AppName == "" {
		return nil, errors.New("AppName is required")
	}

	slog.Debug("new handler",
		slog.Group("handler",
			slog.String("AppName", h.AppName),
		),
	)

	return h, nil
}
