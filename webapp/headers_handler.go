// Copyright 2024 Bill Nixon. All rights reserved.
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

// SortHeaders uses httpHeader to create a sorted list of HeaderPair structs.
// The headers are sorted alphabetically by key. If httpHeader is empty,
// it returns nil.
func SortHeaders(httpHeader http.Header) []HeaderPair {
	if len(httpHeader) == 0 {
		return nil
	}

	// Create a slice to store header pairs.
	headerList := make([]HeaderPair, 0, len(httpHeader))

	// Populate the slice with header key-value pairs.
	for key, values := range httpHeader {
		headerList = append(headerList,
			HeaderPair{Key: key, Values: values})
	}

	// Sort the slice of headers by key.
	sort.Slice(headerList, func(i, j int) bool {
		return headerList[i].Key < headerList[j].Key
	})

	return headerList
}

// HeadersHandlerGet shows the headers of the request in sorted order.
func (app *WebApp) HeadersHandlerGet(w http.ResponseWriter, r *http.Request) {
	logger := webhandler.NewRequestLoggerWithFuncName(r)

	if !webutil.IsMethodOrError(w, r, http.MethodGet) {
		logger.Error("invalid method")
		return
	}

	// Sort headers for consistent ordering.
	sortedHeaders := SortHeaders(r.Header)

	data := HeadersPageData{
		Title:   "Request Headers",
		Headers: sortedHeaders,
	}

	err := webutil.RenderTemplateOrError(app.Tmpl, w, HeadersPageName, data)
	if err != nil {
		logger.Error("failed to RenderTemplate", "err", err)
		return
	}

	logger.Info("done", slog.Any("data", data))
}
