// Copyright 2023 Bill Nixon. All rights reserved.
// Use of this source code is governed by the license found in the LICENSE file.

package webhandler

import (
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"time"

	"github.com/bnixon67/webapp/webutil"
)

// BuildDateTimeFormat can be used to format a time as "YYYY-MM-DD HH:MM:SS"
const BuildDateTimeFormat = "2006-01-02 15:04:05"

// BuildHandler responds with the executable modification date and time.
func (h *Handler) BuildHandler(w http.ResponseWriter, r *http.Request) {
	// Get logger with request info from request context and add calling function name.
	logger := Logger(r.Context()).With(slog.String("func", FuncName()))

	// Check if the HTTP method is valid.
	if !webutil.ValidMethod(w, r, http.MethodGet) {
		logger.Error("invalid method")
		return
	}

	// Set no-cache headers to prevent caching of the response.
	webutil.SetNoCacheHeaders(w)

	// Format the time as a string.
	build := h.BuildDateTime.Format(BuildDateTimeFormat)

	logger.Debug("response", slog.String("build", build))

	// Set the content type of the response to text.
	webutil.SetTextContentType(w)

	// Write the build time to the response.
	fmt.Fprintln(w, build)
}

// ExecutableModTime returns the modification time of the executable file.
func ExecutableModTime() (time.Time, error) {
	// Get path of the current executable.
	execPath, err := os.Executable()
	if err != nil {
		return time.Time{}, fmt.Errorf("error getting executable path: %w", err)
	}

	// Get file information of the executable.
	fileInfo, err := os.Stat(execPath)
	if err != nil {
		return time.Time{}, fmt.Errorf("error getting file info: %w", err)
	}

	// Return modification time.
	return fileInfo.ModTime(), nil
}
