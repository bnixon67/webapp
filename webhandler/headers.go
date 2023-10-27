package webhandler

import (
	"log/slog"
	"net/http"
	"sort"

	"github.com/bnixon67/webapp/webutil"
)

// HeadersPageName is the name of the HTTP template to execute.
const HeadersPageName = "headers.html"

// HeaderInfo contains individual header details.
type HeaderInfo struct {
	Key   string
	Value []string
}

// HeadersPageData holds the data passed to the HTML template.
type HeadersPageData struct {
	Title   string       // Title of the page.
	Headers []HeaderInfo // Sorted list of the request headers.
}

// NewHeaderInfo inits and returns a sorted array of HeaderInfo from httpHeader.
func NewHeaderInfo(httpHeader http.Header) []HeaderInfo {
	headerList := make([]HeaderInfo, 0, len(httpHeader))
	for key, values := range httpHeader {
		headerList = append(headerList, HeaderInfo{Key: key, Value: values})
	}

	// Sort headers based on their key names
	sort.Slice(headerList, func(i, j int) bool { return headerList[i].Key < headerList[j].Key })

	return headerList
}

// HeadersHandler prints the headers of the request in sorted order.
func (h *Handler) HeadersHandler(w http.ResponseWriter, r *http.Request) {
	// get the logger from the context, which include request information
	logger := Logger(r.Context())

	// check for valid methods
	if !webutil.ValidMethod(w, r, http.MethodGet) {
		logger.Error("invalid method")
		return
	}

	sortedHeaders := NewHeaderInfo(r.Header)

	data := HeadersPageData{
		Title:   "Request Headers",
		Headers: sortedHeaders,
	}

	logger.Info("exec",
		slog.String("func", "HeadersHandler"),
		slog.Any("data", data),
	)

	err := webutil.RenderTemplate(h.Tmpl, w, HeadersPageName, data)
	if err != nil {
		logger.Error("failed to RenderTemplate", "err", err)
		return
	}
}
