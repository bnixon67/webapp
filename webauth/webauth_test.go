// Copyright 2024 Bill Nixon. All rights reserved.
// Use of this source code is governed by the license found in the LICENSE file.

package webauth_test

import (
	"testing"
	"text/template"

	_ "github.com/go-sql-driver/mysql"

	"github.com/bnixon67/webapp/webapp"
	"github.com/bnixon67/webapp/webauth"
	"github.com/bnixon67/webapp/weblog"
	"github.com/bnixon67/webapp/webutil"
)

func TestNew(t *testing.T) {
	cfg, err := webauth.ConfigFromJSONFile(TestConfigFile)
	if err != nil {
		t.Fatalf("failed to created config: %v", err)
	}

	tests := []struct {
		name     string
		opts     []interface{}
		wantName string
		wantErr  bool
	}{
		{
			name: "With AppName",
			opts: []interface{}{
				webapp.WithName("TestApp"),
				webauth.WithConfig(cfg),
			},
			wantName: "TestApp",
			wantErr:  false,
		},
		{
			name: "With AppName and Foo",
			opts: []interface{}{
				webapp.WithName("TestApp"),
				webauth.WithConfig(cfg),
			},
			wantName: "TestApp",
			wantErr:  false,
		},
		{
			name:     "Without AppName",
			wantName: "",
			wantErr:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			h, err := webauth.New(tt.opts...)
			if (err != nil) != tt.wantErr {
				t.Errorf("New() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr && h.Name != tt.wantName {
				t.Errorf("New() AppName = %v, want %v", h.Name, tt.wantName)
			}
		})
	}
}

const (
	TestLogFile    = "test.log"
	TestConfigFile = "testdata/test_config.json"
)

// global to provide a singleton app.
var app *webauth.AuthApp //nolint

// AppForTest is a helper function that returns an App used for testing.
func AppForTest(t *testing.T) *webauth.AuthApp {
	if app == nil {
		// Initialize logging.
		logLevel := "INFO"
		err := weblog.Init(weblog.WithLevel(logLevel))
		if err != nil {
			t.Fatalf("failed to initialize logging: %v", err)
		}

		cfg, err := webauth.ConfigFromJSONFile(TestConfigFile)
		if err != nil {
			t.Fatalf("failed to created config: %v", err)
		}

		// Define the custom function
		funcMap := template.FuncMap{
			"ToTimeZone": webutil.ToTimeZone,
			"Join":       webutil.Join,
		}

		// Initialize templates
		tmpl, err := webutil.TemplatesWithFuncs(cfg.ParseGlobPattern, funcMap)
		if err != nil {
			t.Fatalf("failed to init templates: %v", err)
		}

		db, err := webauth.InitDB(cfg.SQL.DriverName, cfg.SQL.DataSourceName)
		if err != nil {
			t.Fatalf("failed to init db: %v", err)
		}

		app, err = webauth.New(
			webapp.WithTemplate(tmpl),
			webapp.WithName(cfg.App.Name),
			webauth.WithConfig(cfg),
			webauth.WithDB(db),
		)
		if err != nil {
			app = nil

			t.Fatalf("cannot create NewApp, %v", err)
		}
	}

	return app
}
