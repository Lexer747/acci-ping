// Use of this source code is governed by a GPL-2 license that can be found in the LICENSE file.
//
// Copyright 2025 Lexer747
//
// SPDX-License-Identifier: GPL-2.0-only

//nolint:staticcheck
package env

import (
	"os"
	"strings"
)

func SHOULD_TEST_NETWORK() bool {
	str := os.Getenv("SHOULD_TEST_NETWORK")
	return str == "1"
}

func LOCAL_FRAME_DIFFS() bool {
	str := os.Getenv("LOCAL_FRAME_DIFFS")
	return str == "1"
}

// AddTo_PATH will in memory only, adding the given [toAdd] file path to the [initialEnv] of key value pairs
// (see [Cmd.Env]), finding the PATH key and adding to that key or creating it if no PATH key was found.
//
// Note that this doesn't set the current processes environment you can achieve that with [os.Setenv] more
// easily instead.
//
// [Cmd.Env]: https://pkg.go.dev/os/exec#Cmd.Env
func AddTo_PATH(initialEnv []string, toAdd string) []string {
	for i, keyValue := range initialEnv {
		split := strings.Split(keyValue, "=")
		if len(split) <= 1 {
			continue
		}
		key := split[0]
		value := split[1]
		if key != "PATH" {
			continue
		}
		initialEnv[i] = "PATH=" + value + ":" + toAdd
		return initialEnv
	}
	initialEnv = append(initialEnv, "PATH="+toAdd)
	return initialEnv
}
