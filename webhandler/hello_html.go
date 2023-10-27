package webhandler

import (
	"fmt"
	"log/slog"
	"net/http"

	"github.com/bnixon67/webapp/assets"
	"github.com/bnixon67/webapp/webutil"
)

// HelloHTMLHandler responds with a simple "hello" message in HTML format.
func (h *Handler) HelloHTMLHandler(w http.ResponseWriter, r *http.Request) {
	logger := Logger(r.Context())

	if !webutil.ValidMethod(w, r, http.MethodGet) {
		logger.Error("invalid method")
		return
	}

	logger.Info("exec", slog.String("func", "HelloHTMLHandler"))

	webutil.SetNoCacheHeaders(w)

	// Write the HTML content to the response
	fmt.Fprint(w, assets.HelloHTML)
}
