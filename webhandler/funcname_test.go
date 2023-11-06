// Copyright 2023 Bill Nixon. All rights reserved.
// Use of this source code is governed by the license found in the LICENSE file.

package webhandler_test

import (
	"testing"

	"github.com/bnixon67/webapp/webhandler"
)

func TestGetCallingFuncName(t *testing.T) {
	funcName := webhandler.FuncName()
	want := "TestGetCallingFuncName"

	if funcName != want {
		t.Errorf("Expected: %s, Got: %s", want, funcName)
	}
}

func testParentFunc() string {
	return webhandler.FuncNameParent()
}

func TestGetCallingFuncNameParent(t *testing.T) {
	funcName := testParentFunc()
	want := "TestGetCallingFuncNameParent"

	if funcName != want {
		t.Errorf("Expected: %s, Got: %s", want, funcName)
	}
}
