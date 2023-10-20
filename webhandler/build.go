package webhandler

import (
	"fmt"
	"net/http"
	"os"
	"time"

	"github.com/bnixon67/webapp/webutils"
)

const MsgExecDateTimeErr = "Cannot get executable datetime"

// BuildHandler responds with the executable modification date and time.
func (h *Handler) BuildHandler(w http.ResponseWriter, r *http.Request) {
	logger := Logger(r.Context())

	if !webutils.ValidMethod(w, r, http.MethodGet) {
		logger.Error("invalid method")
		return
	}

	webutils.SetNoCacheHeaders(w)

	// get executable date/time
	dt, err := ExecutableDateTime()
	if err != nil {
		logger.Error(MsgExecDateTimeErr, "err", err)
		http.Error(w, MsgExecDateTimeErr, http.StatusInternalServerError)
		return
	}

	build := dt.Format("2006-01-02 15:04:05")

	logger.Info("BuildHandler", "build", build)

	webutils.SetTextContentType(w)
	fmt.Fprintln(w, build)
}

// ExecutableDateTime returns the modification date/time of the executable file.
func ExecutableDateTime() (time.Time, error) {
	// Get the path of the executable
	executablePath, err := os.Executable()
	if err != nil {
		return time.Time{}, fmt.Errorf("error getting executable path: %w", err)
	}

	// Get file information
	fileInfo, err := os.Stat(executablePath)
	if err != nil {
		return time.Time{}, fmt.Errorf("error getting file info: %w", err)
	}

	return fileInfo.ModTime(), nil
}
