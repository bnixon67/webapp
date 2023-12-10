// Copyright 2023 Bill Nixon. All rights reserved.
// Use of this source code is governed by the license found in the LICENSE file.

package weblogin_test

import (
	"testing"
	"text/template"

	_ "github.com/go-sql-driver/mysql"

	"github.com/bnixon67/webapp/webapp"
	"github.com/bnixon67/webapp/weblog"
	"github.com/bnixon67/webapp/weblogin"
	"github.com/bnixon67/webapp/webutil"
)

func TestNew(t *testing.T) {
	cfg, err := weblogin.GetConfigFromFile(TestConfigFile)
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
				webapp.WithAppName("TestApp"),
				weblogin.WithConfig(cfg),
			},
			wantName: "TestApp",
			wantErr:  false,
		},
		{
			name: "With AppName and Foo",
			opts: []interface{}{
				webapp.WithAppName("TestApp"),
				weblogin.WithConfig(cfg),
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
			h, err := weblogin.New(tt.opts...)
			if (err != nil) != tt.wantErr {
				t.Errorf("New() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr && h.AppName != tt.wantName {
				t.Errorf("New() AppName = %v, want %v", h.AppName, tt.wantName)
			}
		})
	}
}

const (
	TestLogFile    = "test.log"
	TestConfigFile = "testdata/test_config.json"
)

// global to provide a singleton app.
var app *weblogin.LoginApp //nolint

// AppForTest is a helper function that returns an App used for testing.
func AppForTest(t *testing.T) *weblogin.LoginApp {
	if app == nil {
		// Initialize logging.
		logLevel := "INFO"
		err := weblog.Init(weblog.WithLevel(logLevel))
		if err != nil {
			t.Fatalf("failed to initialize logging: %v", err)
		}

		cfg, err := weblogin.GetConfigFromFile(TestConfigFile)
		if err != nil {
			t.Fatalf("failed to created config: %v", err)
		}

		// Define the custom function
		funcMap := template.FuncMap{
			"ToTimeZone": webutil.ToTimeZone,
		}

		// Initialize templates
		tmpl, err := webutil.InitTemplatesWithFuncMap(cfg.ParseGlobPattern, funcMap)
		if err != nil {
			t.Fatalf("failed to init templates: %v", err)
		}

		db, err := weblogin.InitDB(cfg.SQL.DriverName, cfg.SQL.DataSourceName)
		if err != nil {
			t.Fatalf("failed to init db: %v", err)
		}

		app, err = weblogin.New(
			webapp.WithTemplate(tmpl),
			webapp.WithAppName(cfg.Name),
			weblogin.WithConfig(cfg),
			weblogin.WithDB(db),
		)
		if err != nil {
			app = nil

			t.Fatalf("cannot create NewApp, %v", err)
		}
	}

	return app
}
