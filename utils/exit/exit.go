// Use of this source code is governed by a GPL-2 license that can be found in the LICENSE file.
//
// Copyright 2025 Lexer747
//
// SPDX-License-Identifier: GPL-2.0-only

package exit

import (
	"fmt"
	"log/slog"
	"os"
)

// OnError should be called when there is no way from the program to continue functioning normally, if err is
// not nil the program will exit and print the error which caused the issue.
func OnError(err error) {
	if err != nil {
		slog.Error(fmt.Sprintf("Exiting with %d", errCode), "err", err.Error())
		fmt.Fprintln(os.Stderr, err.Error())
		os.Exit(errCode)
	}
}

// OnErrorMsg is like [OnError] but has a custom message when err is not nil.
func OnErrorMsg(err error, msg string) {
	if err != nil {
		slog.Error(fmt.Sprintf("Exiting with %d", errCode), "err", err.Error(), "msg", msg)
		fmt.Fprintf(os.Stderr, msg+": %s", err.Error())
		os.Exit(errCode)
	}
}

// OnErrorMsgf is like [OnErrorMsg] but will format the string according to printf before writing it.
func OnErrorMsgf(err error, format string, args ...any) {
	if err != nil {
		slog.Error(fmt.Sprintf("Exiting with %d", errCode), "err", err.Error(), "msg", fmt.Sprintf(format, args...))
		fmt.Fprintf(os.Stderr, fmt.Sprintf(format, args...)+": %s", err.Error())
		os.Exit(errCode)
	}
}

// Success is a alias for [os.Exit(0)].
func Success() {
	os.Exit(0)
}

func Silent() {
	os.Exit(1)
}

const errCode = -1
