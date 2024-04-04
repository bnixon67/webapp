// Copyright 2024 Bill Nixon. All rights reserved.
// Use of this source code is governed by the license found in the LICENSE file.

// Package assets provides access to embeded assets.
package assets

import (
	_ "embed"
	"path/filepath"
	"runtime"
)

//go:embed html/hello.html
var HelloHTML string // Embedded HTML page for a simple greeting.

// AssetPath returns the directory of the file that calls this function.
// It's useful for determining the path context in runtime, especially for
// locating assets relative to executing code.  Returns an empty string if
// the caller's path cannot be determined.
func AssetPath() string {
	_, file, _, ok := runtime.Caller(0)
	if !ok {
		return ""
	}

	return filepath.Dir(file)
}
