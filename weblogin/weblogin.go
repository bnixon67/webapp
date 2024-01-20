// Copyright 2023 Bill Nixon. All rights reserved.
// Use of this source code is governed by the license found in the LICENSE file.

// Package weblogin provides common functions for web applications with form based authentication.
package weblogin

import (
	"errors"
	"fmt"
	"log/slog"
	"strings"
	"time"

	"github.com/bnixon67/webapp/webapp"
)

// LoginApp extends WebApp with additional variables.
type LoginApp struct {
	*webapp.WebApp          // Embedded WebApp
	DB             *LoginDB // DB is the database connection.
	Cfg            Config
}

func (a *LoginApp) String() string {
	if a == nil {
		return fmt.Sprintf("%v", nil)
	}

	return fmt.Sprintf("%+v", *a)
}

// Option is a function type used to apply configuration options to a LoginApp.
// This follows the Option pattern from https://commandcenter.blogspot.com/2014/01/self-referential-functions-and-design.html and elsewhere.
type Option func(*LoginApp)

// WithDB returns an Option to set the DB of a LoginApp.
func WithDB(db *LoginDB) Option {
	return func(a *LoginApp) {
		a.DB = db
	}
}

// WithConfig returns an Option to set the Config of a LoginApp.
func WithConfig(cfg Config) Option {
	return func(a *LoginApp) {
		a.Cfg = cfg
	}
}

var ErrAppInvalidConfig = errors.New("invalid config")

// New creates a new LoginApp with the given options and returns it.
// These options can be either LoginApp or WebApp Options.
func New(options ...interface{}) (*LoginApp, error) {
	// Initialize the LoginApp structure.
	loginApp := &LoginApp{}

	// Slice to store options that are applicable to WebApp.
	var webAppOpts []webapp.Option

	// Iterate over each provided option.
	for _, opt := range options {
		switch o := opt.(type) {
		case webapp.Option:
			// Append WebApp option to webAppOpts slice.
			webAppOpts = append(webAppOpts, o)
		case Option:
			// Apply LoginApp option.
			o(loginApp)
		default:
			// If the option doesn't match expected types, return an error.
			return nil, fmt.Errorf("invalid option type: %T", opt)
		}
	}

	// Initialize the WebApp portion of LoginApp.
	var err error
	loginApp.WebApp, err = webapp.New(webAppOpts...)
	if err != nil {
		return nil, fmt.Errorf("error initializing WebApp: %w", err)
	}

	isValid, missing := loginApp.Cfg.IsValid()
	if !isValid {
		return nil, fmt.Errorf("%w: %s",
			ErrAppInvalidConfig, strings.Join(missing, ", "))
	}

	_, err = time.ParseDuration(loginApp.Cfg.LoginExpires)
	if err != nil {
		return nil, fmt.Errorf("%w: %s",
			ErrAppInvalidConfig, err)
	}

	slog.Debug("created new login app", slog.String("loginApp", loginApp.String()))

	return loginApp, nil
}
