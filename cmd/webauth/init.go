package main

import (
	"html/template"

	"github.com/bnixon67/webapp/webauth"
	"github.com/bnixon67/webapp/webutil"
)

// Init initializes logging, templates, and database.
func Init(cfg webauth.Config) (*template.Template, *webauth.AuthDB, error) {
	// Initialize logging.
	err := cfg.Log.Init()
	if err != nil {
		return nil, nil, err
	}

	// Initialize templates with custom functions.
	tmpl, err := webutil.TemplatesWithFuncs(cfg.App.TmplPattern,
		template.FuncMap{
			"ToTimeZone": webutil.ToTimeZone,
			"Join":       webutil.Join,
		})
	if err != nil {
		return nil, nil, err
	}

	// Initialize db
	db, err := webauth.InitDB(cfg.SQL.DriverName, cfg.SQL.DataSourceName)
	if err != nil {
		return nil, nil, err
	}

	return tmpl, db, nil
}
