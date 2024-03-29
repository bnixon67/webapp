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

// HeadersHandlerGet shows the headers of the request in sorted order.
func (app *WebApp) HeadersHandlerGet(w http.ResponseWriter, r *http.Request) {
	logger := webhandler.RequestLoggerWithFunc(r)

	if !webutil.EnforceMethod(w, r, http.MethodGet) {
		logger.Error("invalid method")
		return
	}

	sortedHeaders := SortHeaders(r.Header) // sort for consistent ordering

	data := HeadersPageData{
		Title:   "Request Headers",
		Headers: sortedHeaders,
	}

	err := webutil.RenderTemplate(app.Tmpl, w, HeadersPageName, data)
	if err != nil {
		logger.Error("failed to RenderTemplate", "err", err)
		return
	}

	logger.Info("done", slog.Any("data", data))
}
