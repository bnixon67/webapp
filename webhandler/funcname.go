// Copyright 2023 Bill Nixon. All rights reserved.
// Use of this source code is governed by the license found in the LICENSE file.

package webhandler

import (
	"runtime"
	"strings"
)

// funcName retrieves the name of the function at a given call stack depth.
// The name does not include any path or package information.
// For the calling function use depth 1, for its caller use depth 2, etc.
// If the function name cannot be determined, "unknown" is returned.
func funcName(depth int) string {
	// Get the program counter (PC) for function based on depth.
	pc, _, _, ok := runtime.Caller(depth)
	if !ok {
		return "unknown"
	}

	// Retrieve function information.
	fn := runtime.FuncForPC(pc)
	if fn == nil {
		return "unknown"
	}

	// Function name may include packages prior to function name.
	funcName := fn.Name()

	// Function name is after the last dot.
	if lastIndex := strings.LastIndex(funcName, "."); lastIndex >= 0 {
		return funcName[lastIndex+1:]
	}

	// In the rare case that there's no dot, return the full name.
	return funcName
}

// FuncName returns the name of the function that called it.
// If the function name cannot be determined, "unknown" is returned.
func FuncName() string {
	return funcName(2) // Depth 2 for this function and its caller.
}

// FuncNameParent returns the name of the parent of the calling function.
// If the parent function name cannot be determined, "unknown" is returned.
func FuncNameParent() string {
	return funcName(3) // Depth 3 for this function, caller, and parent.
}
