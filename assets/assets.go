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
var HelloHTML string

// AssetPath returns the directory path of the file that calls this function.
// If the file's path cannot be determined, it returns an empty string.
func AssetPath() string {
	_, file, _, ok := runtime.Caller(0)
	if !ok {
		return ""
	}

	return filepath.Dir(file)
}
