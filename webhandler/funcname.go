// Copyright 2023 Bill Nixon. All rights reserved.
// Use of this source code is governed by the license found in the LICENSE file.

package webhandler

import (
	"runtime"
	"strings"
)

// FuncName returns the name of the calling function.
// If it cannot determine the calling function, it returns "unknown".
func FuncName() string {
	// Get the program counter (PC) for the function that called this function.
	// The depth of 1 indicates the immediate caller.
	pc, _, _, _ := runtime.Caller(1)

	// Retrieve function information for the calling function.
	callingFunction := runtime.FuncForPC(pc)

	// If the calling function information is not available, return "unknown."
	if callingFunction == nil {
		return "unknown"
	}

	// Get the full function name, which includes package and function names.
	fullFuncName := callingFunction.Name()

	// Split the full name by the dot separator.
	parts := strings.Split(fullFuncName, ".")

	// Return just the function name.
	return parts[len(parts)-1]
}
