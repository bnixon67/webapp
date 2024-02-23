// Copyright 2023 Bill Nixon. All rights reserved.
// Use of this source code is governed by the license found in the LICENSE file.

// Package webauth provides support for web apps that use form authentication.
package webauth

import (
	"errors"
	"fmt"
	"log/slog"
	"strings"
	"time"

	"github.com/bnixon67/webapp/webapp"
)

// AuthApp extends WebApp with additional variables.
type AuthApp struct {
	*webapp.WebApp          // Embedded WebApp
	DB             *LoginDB // DB is the database connection.
	Cfg            Config
}

func (a *AuthApp) String() string {
	if a == nil {
		return fmt.Sprintf("%v", nil)
	}

	return fmt.Sprintf("%+v", *a)
}

// Option is a function type used to apply configuration options to a AuthApp.
// This follows the Option pattern from https://commandcenter.blogspot.com/2014/01/self-referential-functions-and-design.html and elsewhere.
type Option func(*AuthApp)

// WithDB returns an Option to set the DB of a AuthApp.
func WithDB(db *LoginDB) Option {
	return func(a *AuthApp) {
		a.DB = db
	}
}

// WithConfig returns an Option to set the Config of a AuthApp.
func WithConfig(cfg Config) Option {
	return func(a *AuthApp) {
		a.Cfg = cfg
	}
}

var ErrAppInvalidConfig = errors.New("invalid config")

// New creates a new AuthApp with the given options and returns it.
// These options can be either AuthApp or WebApp Options.
func New(options ...interface{}) (*AuthApp, error) {
	// Initialize the AuthApp structure.
	authApp := &AuthApp{}

	// Slice to store options that are applicable to WebApp.
	var webAppOpts []webapp.Option

	// Iterate over each provided option.
	for _, opt := range options {
		switch o := opt.(type) {
		case webapp.Option:
			// Append WebApp option to webAppOpts slice.
			webAppOpts = append(webAppOpts, o)
		case Option:
			// Apply AuthApp option.
			o(authApp)
		default:
			// If the option doesn't match expected types, return an error.
			return nil, fmt.Errorf("invalid option type: %T", opt)
		}
	}

	// Initialize the WebApp portion of AuthApp.
	var err error
	authApp.WebApp, err = webapp.New(webAppOpts...)
	if err != nil {
		return nil, fmt.Errorf("error initializing WebApp: %w", err)
	}

	isValid, missing := authApp.Cfg.IsValid()
	if !isValid {
		return nil, fmt.Errorf("%w: %s",
			ErrAppInvalidConfig, strings.Join(missing, ", "))
	}

	_, err = time.ParseDuration(authApp.Cfg.LoginExpires)
	if err != nil {
		return nil, fmt.Errorf("%w: %s",
			ErrAppInvalidConfig, err)
	}

	slog.Debug("created new auth app", slog.String("authApp", authApp.String()))

	return authApp, nil
}
