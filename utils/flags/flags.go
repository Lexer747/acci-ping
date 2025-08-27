// Use of this source code is governed by a GPL-2 license that can be found in the LICENSE file.
//
// Copyright 2025 Lexer747
//
// SPDX-License-Identifier: GPL-2.0-only

package flags

import (
	"flag"
	"fmt"
	"io"
	"reflect"
	"strings"
)

// flag package doesn't trivially provide a way to override *which* flags to be printed so instead I fork the
// original definition and provide a way to control which flags are being visited. go version v1.24.3

type FilterFunc func(f *flag.Flag) bool

func ExcludePrefix(prefix string) FilterFunc {
	return func(f *flag.Flag) bool {
		return strings.HasPrefix(f.Name, prefix)
	}
}

func NoFilter() FilterFunc {
	return func(*flag.Flag) bool { return false }
}

type VisitLike interface {
	VisitAll(fn func(*flag.Flag))
	Output() io.Writer
}

func PrintFlagsFilter(fs VisitLike, filterFunction FilterFunc) {
	var isZeroValueErrs []error
	fs.VisitAll(func(f *flag.Flag) {
		if filterFunction(f) {
			// Don't print values which are filtered
			return
		}
		isZeroValueErrs = forkedInLoopImpl(f, isZeroValueErrs, fs)
	})
	forkedReturnImpl(isZeroValueErrs, fs)
}

// Copyright 2009 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the golang repository.

// forked from go version v1.24.3
// nolint
func forkedInLoopImpl(f *flag.Flag, isZeroValueErrs []error, fs VisitLike) []error {
	var b strings.Builder
	fmt.Fprintf(&b, "  -%s", f.Name) // Two spaces before -; see next two comments.
	name, usage := flag.UnquoteUsage(f)
	if len(name) > 0 {
		b.WriteString(" ")
		b.WriteString(name)
	}
	// Boolean flags of one ASCII letter are so common we
	// treat them specially, putting their usage on the same line.

	if b.Len() <= 4 { // space, space, '-', 'x'.
		b.WriteString("\t")
	} else {
		// Four spaces before the tab triggers good alignment
		// for both 4- and 8-space tab stops.
		b.WriteString("\n    \t")
	}
	b.WriteString(strings.ReplaceAll(usage, "\n", "\n    \t"))

	// Print the default value only if it differs to the zero value
	// for this flag type.
	if isZero, err := forkIsZeroValue(f, f.DefValue); err != nil {
		isZeroValueErrs = append(isZeroValueErrs, err)
	} else if !isZero {
		if _, ok := f.Value.(*stringValue); ok {
			// put quotes on the value
			fmt.Fprintf(&b, " (default %q)", f.DefValue)
		} else {
			fmt.Fprintf(&b, " (default %v)", f.DefValue)
		}
	}
	fmt.Fprint(fs.Output(), b.String(), "\n")
	return isZeroValueErrs
}

// nolint
func forkedReturnImpl(isZeroValueErrs []error, fs VisitLike) {
	// If calling String on any zero flag.Values triggered a panic, print
	// the messages after the full set of defaults so that the programmer
	// knows to fix the panic.
	if errs := isZeroValueErrs; len(errs) > 0 {
		fmt.Fprintln(fs.Output())
		for _, err := range errs {
			fmt.Fprintln(fs.Output(), err)
		}
	}
}

// forked from go version v1.24.3
// nolint
func forkIsZeroValue(f *flag.Flag, value string) (ok bool, err error) {
	// Build a zero value of the flag's Value type, and see if the
	// result of calling its String method equals the value passed in.
	// This works unless the Value type is itself an interface type.
	typ := reflect.TypeOf(f.Value)
	var z reflect.Value
	if typ.Kind() == reflect.Pointer {
		z = reflect.New(typ.Elem())
	} else {
		z = reflect.Zero(typ)
	}
	// Catch panics calling the String method, which shouldn't prevent the
	// usage message from being printed, but that we should report to the
	// user so that they know to fix their code.
	defer func() {
		if e := recover(); e != nil {
			if typ.Kind() == reflect.Pointer {
				typ = typ.Elem()
			}
			err = fmt.Errorf("panic calling String method on zero %v for flag %s: %v", typ, f.Name, e)
		}
	}()
	return value == z.Interface().(flag.Value).String(), nil
}

// forked from go version v1.24.3
type stringValue string

func (s *stringValue) Set(val string) error {
	*s = stringValue(val)
	return nil
}

func (s *stringValue) Get() any { return string(*s) }

func (s *stringValue) String() string { return string(*s) }
