// Copyright 2024 Bill Nixon. All rights reserved.
// Use of this source code is governed by the license found in the LICENSE file.

// Package webauth provides support for web apps that use form authentication.
package webauth

import (
	"errors"
	"fmt"
	"log/slog"
	"time"

	"github.com/bnixon67/webapp/webapp"
)

// AuthApp extends the WebApp to support authentication.
type AuthApp struct {
	*webapp.WebApp         // Embedded WebApp
	DB             *AuthDB // DB is the database connection.
	Cfg            Config
}

// String returns a string representation of the AuthApp instance.
func (a *AuthApp) String() string {
	if a == nil {
		return fmt.Sprintf("%v", nil)
	}

	return fmt.Sprintf("%+v", *a)
}

// Option is a function type used to apply configuration options to a AuthApp.
// This follows the Option pattern from https://commandcenter.blogspot.com/2014/01/self-referential-functions-and-design.html and elsewhere.
type Option func(*AuthApp)

// WithDB returns an Option to set the DB for a AuthApp.
func WithDB(db *AuthDB) Option {
	return func(a *AuthApp) {
		a.DB = db
	}
}

// WithConfig returns an Option to set the Config for a AuthApp.
func WithConfig(cfg Config) Option {
	return func(a *AuthApp) {
		a.Cfg = cfg
	}
}

var ErrInvalidConfig = errors.New("invalid config")

// NewApp creates a new AuthApp with the given options and returns it.
// These options can be either AuthApp or WebApp Options.
func NewApp(options ...interface{}) (*AuthApp, error) {
	authApp := &AuthApp{}

	var webAppOpts []webapp.Option

	// Process options for both AuthApp and WebApp.
	for _, opt := range options {
		switch o := opt.(type) {
		case webapp.Option:
			// Collect WebApp options.
			webAppOpts = append(webAppOpts, o)
		case Option:
			// Apply AuthApp option.
			o(authApp)
		default:
			// Handle unexpected option types.
			return nil, fmt.Errorf("invalid option type: %T", opt)
		}
	}

	// Initialize embedded WebApp.
	var err error
	authApp.WebApp, err = webapp.New(webAppOpts...)
	if err != nil {
		return nil, fmt.Errorf("error initializing WebApp: %w", err)
	}

	// Validate configuration.
	isValid, missing, err := authApp.Cfg.IsValid()
	if err != nil {
		return nil, err
	}
	if !isValid {
		return nil, fmt.Errorf("%w: %v", ErrInvalidConfig, missing)
	}

	// Validate login expiration duration.
	_, err = time.ParseDuration(authApp.Cfg.Auth.LoginExpires)
	if err != nil {
		return nil, fmt.Errorf("%w: %s", ErrInvalidConfig, err)
	}

	slog.Debug("created new auth app",
		slog.String("authApp", authApp.String()))

	return authApp, nil
}
