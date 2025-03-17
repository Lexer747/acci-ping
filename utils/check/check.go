// Use of this source code is governed by a GPL-2 license that can be found in the LICENSE file.
//
// Copyright 2024-2025 Lexer747
//
// SPDX-License-Identifier: GPL-2.0-only

package check

import (
	"fmt"
	"reflect"
)

func Check(shouldBeTrue bool, assertMsg string) {
	if !shouldBeTrue {
		panic("check failed: " + assertMsg)
	}
}

func Checkf(shouldBeTrue bool, format string, a ...any) {
	if !shouldBeTrue {
		panic("check failed: " + fmt.Sprintf(format, a...))
	}
}

func NotNil(ptr any, assertMsg string) {
	asIntPtr, ok := getPtr(ptr)
	Checkf(ok, "Check failed: NotNilf() called on non-pointer type %T", ptr)
	Check(asIntPtr != 0, assertMsg)
}

func NotNilf(ptr any, format string, a ...any) {
	asIntPtr, ok := getPtr(ptr)
	Checkf(ok, "Check failed: NotNilf() called on non-pointer type %T", ptr)
	Checkf(asIntPtr != 0, format, a...)
}

func NoErr(err error, msg string) {
	Checkf(err == nil, "%s: %s", msg, err)
}

func NoErrf(err error, format string, args ...any) {
	Checkf(err == nil, format+": %s", append(args, err)...)
}

func Must[T any](t T, err error) T {
	NoErr(err, "Must")
	return t
}

func getPtr(a any) (ret uintptr, ok bool) {
	defer func() {
		if r := recover(); r != nil {
			return
		}
	}()
	ret = reflect.ValueOf(a).Pointer()
	ok = true
	return
}
