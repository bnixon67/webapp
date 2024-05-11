// Copyright 2024 Bill Nixon. All rights reserved.
// Use of this source code is governed by the license found in the LICENSE file.

package webutil

import (
	"bytes"
	"errors"
	"fmt"
	"html/template"
	"log/slog"
	"net/http"
	"strings"
	"testing"
)

// funcMapToString converts a FuncMap to a comma-separated string of
// function names.
func funcMapToString(funcMap template.FuncMap) string {
	var names []string
	for name := range funcMap {
		names = append(names, name)
	}

	return strings.Join(names, ", ")
}

// Templates parses templates from files matching the given pattern.
func Templates(pattern string) (*template.Template, error) {
	return TemplatesWithFuncs(pattern, nil)
}

// TemplatesWithFuncs parses templates from files matching the given pattern
// and applies a FuncMap.
func TemplatesWithFuncs(pattern string, funcMap template.FuncMap) (*template.Template, error) {
	tmpls, err := template.New("tmpl").Funcs(funcMap).ParseGlob(pattern)
	if err != nil {
		return nil, err
	}

	if slog.Default().Enabled(nil, slog.LevelDebug) {
		tmplNames := strings.Join(TemplateNames(tmpls), ", ")
		slog.Debug("parsed templates with functions",
			slog.String("pattern", pattern),
			slog.String("templates", tmplNames),
			slog.String("functions", funcMapToString(funcMap)),
		)
	}

	return tmpls, nil
}

const MsgTemplateError = "The server is unable to display this page."

// RenderTemplateOrError attempts to render a named template with data,
// handling errors by responding with HTTP 500.  The caller must ensure no
// further writes are done for a non-nil error.
func RenderTemplateOrError(tmpl *template.Template, w http.ResponseWriter, name string, data interface{}) error {
	if tmpl == nil {
		http.Error(w, MsgTemplateError, http.StatusInternalServerError)
		return errors.New("renderTemplate: nil template")
	}

	if tmpl.Lookup(name) == nil {
		http.Error(w, MsgTemplateError, http.StatusInternalServerError)
		return fmt.Errorf("renderTemplate: template %q not found", name)
	}

	var buffer bytes.Buffer
	if err := tmpl.ExecuteTemplate(&buffer, name, data); err != nil {
		http.Error(w, MsgTemplateError, http.StatusInternalServerError)
		return err
	}

	if _, err := buffer.WriteTo(w); err != nil {
		http.Error(w, MsgTemplateError, http.StatusInternalServerError)
		return err
	}

	return nil
}

// RenderTemplateForTest executes a template with data for testing purposes,
// returning the output as a string.
func RenderTemplateForTest(t *testing.T, tmpl *template.Template, name string, data any) string {
	var buffer bytes.Buffer

	if err := tmpl.ExecuteTemplate(&buffer, name, data); err != nil {
		t.Fatalf("failed to execute template %q: %v", name, err)
	}

	return buffer.String()
}

// TemplateNames returns a list of all template names within the template tree.
func TemplateNames(tmpl *template.Template) []string {
	if tmpl == nil {
		return nil
	}

	names := make([]string, 0, len(tmpl.Templates()))
	for _, tmpl := range tmpl.Templates() {
		names = append(names, tmpl.Name())
	}

	return names
}
