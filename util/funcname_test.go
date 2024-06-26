// Copyright 2023 Bill Nixon. All rights reserved.
// Use of this source code is governed by the license found in the LICENSE file.

package util_test

import (
	"testing"

	"github.com/bnixon67/webapp/util"
)

func TestFuncName(t *testing.T) {
	funcName := util.FuncName()
	want := "TestFuncName"

	if funcName != want {
		t.Errorf("got %v, want %v", funcName, want)
	}
}

func testFuncNameParent() string {
	return util.FuncNameParent()
}

func TestFuncNameParent(t *testing.T) {
	funcName := testFuncNameParent()
	want := "TestFuncNameParent"

	if funcName != want {
		t.Errorf("got %v, want %v", funcName, want)
	}
}
