package webhandler

import (
	"fmt"
	"log/slog"
	"net/http"

	"github.com/bnixon67/webapp/webutil"
)

// HelloHandler is an HTTP handler method of the Handler type.
// It writes a "hello" message to the HTTP response writer.
// This method can be used to check if the web server is running.
func (h *Handler) HelloHandler(w http.ResponseWriter, r *http.Request) {
	logger := Logger(r.Context())

	if !webutil.ValidMethod(w, r, http.MethodGet) {
		logger.Error("invalid method for HelloHandler")
		return
	}

	logger.Info("exec", slog.String("func", "HelloHandler"))

	webutil.SetTextContentType(w)
	fmt.Fprintln(w, "hello from", h.AppName)
}
