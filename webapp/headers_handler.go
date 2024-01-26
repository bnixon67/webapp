// Copyright 2023 Bill Nixon. All rights reserved.
// Use of this source code is governed by the license found in the LICENSE file.

package webapp

import (
	"log/slog"
	"net/http"
	"sort"

	"github.com/bnixon67/webapp/webhandler"
	"github.com/bnixon67/webapp/webutil"
)

// HeadersPageName is the name of the HTTP template to execute.
const HeadersPageName = "headers.html"

// HeaderPair represents a key-value pair in an HTTP header.
type HeaderPair struct {
	Key    string
	Values []string
}

// HeadersPageData holds the data passed to the HTML template.
type HeadersPageData struct {
	Title   string       // Title of the page.
	Headers []HeaderPair // Sorted list of the request headers.
}

// SortHeaders returns a slice of header pairs sorted by header keys.
func SortHeaders(httpHeader http.Header) []HeaderPair {
	if len(httpHeader) == 0 {
		return nil
	}

	// Pre-allocate slice.
	headerList := make([]HeaderPair, len(httpHeader))

	// Fill the slice with header pairs, flattening multiple values.
	i := 0
	for key, values := range httpHeader {
		headerList[i].Key = key
		headerList[i].Values = values
		i++
	}

	// Sort the slice of headers by key name.
	sort.Slice(headerList, func(i, j int) bool {
		return headerList[i].Key < headerList[j].Key
	})

	return headerList
}

// HeadersHandler prints the headers of the request in sorted order.
func (app *WebApp) HeadersHandler(w http.ResponseWriter, r *http.Request) {
	// Get logger with request info and function name.
	logger := webhandler.RequestLoggerWithFunc(r)

	// Check if the HTTP method is valid.
	if !webutil.ValidMethod(w, r, http.MethodGet) {
		logger.Error("invalid method")
		return
	}

	// Sort the headers from the request for consistent ordering.
	sortedHeaders := SortHeaders(r.Header)

	// Prepare the data for rendering the template.
	data := HeadersPageData{
		Title:   "Request Headers",
		Headers: sortedHeaders,
	}

	// Render the template with the data.
	err := webutil.RenderTemplate(app.Tmpl, w, HeadersPageName, data)
	if err != nil {
		logger.Error("failed to RenderTemplate", "err", err)
		return
	}

	logger.Info("success", slog.Any("data", data))
}
