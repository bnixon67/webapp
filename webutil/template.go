// Copyright 2023 Bill Nixon. All rights reserved.
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
)

// funcMapToString returns a comma separated list of names for funcMap.
func funcMapToString(funcMap template.FuncMap) string {
	var names []string
	for name := range funcMap {
		names = append(names, name)
	}

	return strings.Join(names, ", ")
}

// Templates parses the templates matching pattern.
func Templates(pattern string) (*template.Template, error) {
	return TemplatesWithFuncs(pattern, nil)
}

// TemplatesWithFuncs parses the templates with FuncMap.
func TemplatesWithFuncs(pattern string, funcMap template.FuncMap) (*template.Template, error) {
	tmpls, err := template.New("tmpl").Funcs(funcMap).ParseGlob(pattern)
	if err != nil {
		return nil, fmt.Errorf("TemplatesWithFuncs: %w", err)
	}

	tmplNames := strings.Join(TemplateNames(tmpls), ", ")
	slog.Debug("parsed templates with functions",
		slog.String("pattern", pattern),
		slog.String("templates", tmplNames),
		slog.String("functions", funcMapToString(funcMap)),
	)
	return tmpls, nil
}

const MsgTemplateError = "The server was unable to display this page."

// RenderTemplate executes the named template with the given data and writes
// the result to the provided HTTP response writer.
//
// If an error occurs, sets HTTP response status to 500 and returns the error.
//
// The caller must ensure no further writes are done for a non-nil error.
func RenderTemplate(t *template.Template, w http.ResponseWriter, name string, data interface{}) error {
	// handle nil template
	if t == nil {
		http.Error(w, MsgTemplateError, http.StatusInternalServerError)
		return errors.New("RenderTemplate: nil template")
	}

	if t.Lookup(name) == nil {
		http.Error(w, MsgTemplateError, http.StatusInternalServerError)
		return fmt.Errorf("RenderTemplate: template %s not found", name)
	}

	// Create a buffer to store the template output since if an error
	// occurs executing the template or writing its output, execution
	// stops, but partial results may already have been written to the
	// output writer.
	var tmplBuffer bytes.Buffer

	// Execute the template with the provided data.
	err := t.ExecuteTemplate(&tmplBuffer, name, data)
	if err != nil {
		http.Error(w, MsgTemplateError, http.StatusInternalServerError)
		return err
	}

	// Write the template output to response writer and check for errors.
	_, writeErr := tmplBuffer.WriteTo(w)
	if writeErr != nil {
		http.Error(w, MsgTemplateError, http.StatusInternalServerError)
		return writeErr
	}

	return nil
}

// TemplateNames returns the names of all templates for t.
func TemplateNames(t *template.Template) []string {
	if t == nil {
		return nil
	}

	names := make([]string, 0, len(t.Templates()))
	for _, tmpl := range t.Templates() {
		names = append(names, tmpl.Name())
	}

	return names
}
