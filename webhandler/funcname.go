// Copyright 2023 Bill Nixon. All rights reserved.
// Use of this source code is governed by the license found in the LICENSE file.

package webhandler

import (
	"runtime"
	"strings"
)

// funcName retrieves the name of the function at a given call stack depth.
// 'depth' levels up from the current stack frame. For the calling function use depth 1,
// for its caller use depth 2, and so on.
// If the function name cannot be determined, "unknown" is returned.
func funcName(depth int) string {
	// Get the program counter (PC) for the function that called this function.
	pc, _, _, ok := runtime.Caller(depth)
	if !ok {
		return "unknown"
	}

	// Retrieve function information.
	fn := runtime.FuncForPC(pc)
	if fn == nil {
		return "unknown"
	}

	// Get the full function name, which includes package and function names.
	fullFuncName := fn.Name()

	// The last part of the full function name after the last dot is the actual function name.
	if lastIndex := strings.LastIndex(fullFuncName, "."); lastIndex >= 0 {
		return fullFuncName[lastIndex+1:]
	}

	// In the rare case that there's no dot, return the full name.
	return fullFuncName
}

// FuncName returns the name of the function that called it.
// If the function name cannot be determined, "unknown" is returned.
func FuncName() string {
	return funcName(2) // Depth 2 accounts for this function and its caller.
}

// FuncNameParent returns the name of the parent of the calling function.
// If the parent function name cannot be determined, "unknown" is returned.
func FuncNameParent() string {
	return funcName(3) // Depth 3 accounts for this function, caller, and parent.
}
