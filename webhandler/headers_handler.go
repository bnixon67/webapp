// Copyright 2023 Bill Nixon. All rights reserved.
// Use of this source code is governed by the license found in the LICENSE file.

package webhandler

import (
	"log/slog"
	"net/http"
	"sort"

	"github.com/bnixon67/webapp/webutil"
)

// HeadersPageName is the name of the HTTP template to execute.
const HeadersPageName = "headers.html"

// HeaderPair represents a key-value pair in an HTTP header.
type HeaderPair struct {
	Key   string
	Value []string
}

// HeadersPageData holds the data passed to the HTML template.
type HeadersPageData struct {
	Title   string       // Title of the page.
	Headers []HeaderPair // Sorted list of the request headers.
}

// SortedHeaders returns a slice of headers sorted by header keys.
func SortedHeaders(httpHeader http.Header) []HeaderPair {
	if len(httpHeader) == 0 {
		return nil
	}

	// Initialize the slice with the exact size needed.
	headerList := make([]HeaderPair, len(httpHeader))

	// Iterate over the map to fill the slice using the index 'i'.
	i := 0
	for key, values := range httpHeader {
		headerList[i] = HeaderPair{Key: key, Value: values}
		i++
	}

	// Sort the slice of headers by key name.
	sort.Slice(headerList, func(i, j int) bool {
		return headerList[i].Key < headerList[j].Key
	})

	return headerList
}

// HeadersHandler prints the headers of the request in sorted order.
func (h *Handler) HeadersHandler(w http.ResponseWriter, r *http.Request) {
	// Get logger with request info from request context and add calling function name.
	logger := LoggerFromContext(r.Context()).With(slog.String("func", FuncName()))

	// Check if the HTTP method is valid.
	if !webutil.ValidMethod(w, r, http.MethodGet) {
		logger.Error("invalid method")
		return
	}

	// Sort the headers from the request for consistent ordering.
	sortedHeaders := SortedHeaders(r.Header)

	// Prepare the data for rendering the template.
	data := HeadersPageData{
		Title:   "Request Headers",
		Headers: sortedHeaders,
	}

	logger.Debug("response", slog.Any("data", data))

	// Render the template with the data.
	err := webutil.RenderTemplate(h.Tmpl, w, HeadersPageName, data)
	if err != nil {
		logger.Error("failed to RenderTemplate", "err", err)
		return
	}
}
