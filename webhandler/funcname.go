// Copyright 2024 Bill Nixon. All rights reserved.
// Use of this source code is governed by the license found in the LICENSE file.

package webhandler

import (
	"runtime"
	"strings"
)

// FuncNameAtDepth retrieves the name of the function at a specified call
// stack depth without including package or path information.
//
// depth specifies the call stack level: 1 for the calling function, 2 for
// its caller, and so on.
//
// Returns "unknown" if the function name cannot be determined.
func FuncNameAtDepth(depth int) string {
	// Get the program counter (PC), confirming the function call exists.
	pc, _, _, ok := runtime.Caller(depth)
	if !ok {
		return "unknown"
	}

	// Retrieve and validate function information from the PC.
	fn := runtime.FuncForPC(pc)
	if fn == nil {
		return "unknown"
	}

	// Extract the simple function name from the full function signature.
	funcName := fn.Name()
	if lastIndex := strings.LastIndex(funcName, "."); lastIndex >= 0 {
		return funcName[lastIndex+1:]
	}

	// If there is no dot, return the entire function name.
	return funcName
}

// FuncName returns the name of the function that called it.
//
// If the function name cannot be determined, "unknown" is returned.
func FuncName() string {
	return FuncNameAtDepth(2) // wo levels up, i.e., function, caller.
}

// FuncNameParent returns the name of the parent of the calling function.
//
// If the parent function name cannot be determined, "unknown" is returned.
func FuncNameParent() string {
	return FuncNameAtDepth(3) // Three levels up, i.e., function, caller, parent.
}
