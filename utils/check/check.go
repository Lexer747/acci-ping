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

// Check asserts that the given condition is true, if it is not this is assumed to be an unrecoverable
// violation of the state of the program and will result in a panic. E.g.
//
//	for i, item := range items {
//		check.Check(i < len(items), "i should never be larger than the number of items")
//	}
func Check(shouldBeTrue bool, assertMsg string) {
	if !shouldBeTrue {
		panic("check failed: " + assertMsg)
	}
}

// Checkf asserts that the given condition is true, if it is not this is assumed to be an unrecoverable
// violation of the state of the program and will result in a panic. E.g.
//
//	for i, item := range items {
//		check.Checkf(i < len(items), "i:%d should never be larger than the number of items", i)
//	}
//
// Checkf is a variant which will format the message according to normal go printf semantics.
func Checkf(shouldBeTrue bool, format string, a ...any) {
	if !shouldBeTrue {
		panic("check failed: " + fmt.Sprintf(format, a...))
	}
}

// NotNil asserts that the given pointer is not nil, otherwise this is assumed to be an unrecoverable error
// and will result in a panic. Usually a nicer panic than just dereferencing the nil pointer.
func NotNil(ptr any, assertMsg string) {
	asIntPtr, ok := getPtr(ptr)
	Checkf(ok, "Check failed: NotNilf() called on non-pointer type %T", ptr)
	Check(asIntPtr != 0, assertMsg)
}

// NotNil asserts that the given pointer is not nil, otherwise this is assumed to be an unrecoverable error
// and will result in a panic. Usually a nicer panic than just dereferencing the nil pointer. Formats the
// message according to normal go printf semantics.
func NotNilf(ptr any, format string, a ...any) {
	asIntPtr, ok := getPtr(ptr)
	Checkf(ok, "Check failed: NotNilf() called on non-pointer type %T", ptr)
	Checkf(asIntPtr != 0, format, a...)
}

// NoErr asserts that the given error is in fact nil, if it is not error then it's assumed to be an
// unrecoverable error and will result in a panic.
func NoErr(err error, msg string) {
	Checkf(err == nil, "%s: %s", msg, err)
}

// NoErr asserts that the given error is in fact nil, if it is not error then it's assumed to be an
// unrecoverable error and will result in a panic. Formatting the message according normal go printf
// semantics.
func NoErrf(err error, format string, args ...any) {
	Checkf(err == nil, format+": %s", append(args, err)...)
}

// Must takes the result of a tuple function, e.g.
//
//	Write() (int, error)
//
// And will check the error, panicking if it is not nil, otherwise returning the value. (int in this example).
func Must[T any](t T, err error) T {
	NoErr(err, "Must")
	return t
}

// getPtr is a helper function that lets the library assert that a pointer was actually passed to the library.
// If it's not a pointer the callee should just remove the call and the associated extra work.
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
