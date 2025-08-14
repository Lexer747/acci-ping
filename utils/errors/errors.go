// Use of this source code is governed by a GPL-2 license that can be found in the LICENSE file.
//
// Copyright 2024-2025 Lexer747
//
// SPDX-License-Identifier: GPL-2.0-only

package errors

import (
	stderrors "errors" //nolint:depguard
	"fmt"
)

var New = stderrors.New
var As = stderrors.As
var Is = stderrors.Is
var Join = stderrors.Join
var Unwrap = stderrors.Unwrap

func Errorf(format string, args ...interface{}) error {
	return New(fmt.Sprintf(format, args...))
}

func Wrap(err error, wrapping string) error {
	if err == nil {
		return nil
	}
	return &wrapErr{cause: err, messageErr: New(wrapping)}
}

func Wrapf(err error, format string, args ...interface{}) error {
	return Wrap(err, fmt.Sprintf(format, args...))
}

func WrapErr(err error, wrapping error) error {
	if err == nil {
		return nil
	}
	return &wrapErr{cause: err, messageErr: wrapping}
}

type wrapErr struct {
	cause      error
	messageErr error
}

func (e *wrapErr) Error() string {
	return e.messageErr.Error() + " caused by: " + e.cause.Error()
}

func (e *wrapErr) Format(s fmt.State, verb rune) {
	switch verb {
	case 'v':
		if s.Flag('+') {
			fmt.Fprint(s, e.messageErr)
			fmt.Fprintf(s, " caused by: %+v", e.cause)
			return
		}
		fallthrough
	case 's', 'q':
		fmt.Fprint(s, e.Error())
	}
}

func (e *wrapErr) Unwrap() []error {
	return []error{e.messageErr, e.cause}
}
