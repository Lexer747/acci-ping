// Use of this source code is governed by a GPL-2 license that can be found in the LICENSE file.
//
// Copyright 2025 Lexer747
//
// SPDX-License-Identifier: GPL-2.0-only

package bugfixes_test

import (
	"io"
	"log/slog"
	"os"
	"os/exec"
	"strconv"
	"strings"
	"testing"

	"github.com/Lexer747/acci-ping/utils/check"
	"github.com/Lexer747/acci-ping/utils/env"
	"gotest.tools/v3/assert"
)

//nolint:gochecknoinits
func init() {
	var err error
	path, err = preBuild()
	check.NoErr(err, "could not build acci-ping binary for test")
}

var path string

func preBuild() (string, error) {
	build := exec.Command("/bin/bash", "./build.sh", "unit-tests")
	build.Dir = "../" + build.Dir
	// TODO support other CI envs
	return "./../out/linux/amd64/", build.Run()
}

func TestBugfixes(t *testing.T) {
	t.Parallel()
	f, err := os.Open("./")
	assert.NilError(t, err)
	dirs, err := f.Readdirnames(0)
	assert.NilError(t, err)
	for _, dir := range dirs {
		runTestCase(t, dir)
	}
}

func runTestCase(t *testing.T, dir string) {
	t.Helper()
	reproScript := "./" + dir + "/repro.sh"
	testName := "Bugfixes-'" + reproScript + "'"
	t.Run(testName, func(t *testing.T) {
		t.Parallel()
		_, err := strconv.Atoi(dir)
		if err != nil {
			return
		}
		d, err := os.Open("./" + dir)
		assert.NilError(t, err)
		results, err := d.Readdirnames(0)
		assert.NilError(t, err)
		runScriptAndAssert(t, reproScript, parseTestType(t, results, dir, testName))
	})
}

func parseTestType(t *testing.T, results []string, dir string, testName string) testType {
	t.Helper()
	for _, r := range results {
		if r == "test-type" {
			testFile, err := os.Open("./" + dir + "/" + r)
			assert.NilError(t, err)
			return parseFileTestType(t, testFile, testName)
		}
	}
	return none
}

func parseFileTestType(t *testing.T, testFile io.Reader, testName string) testType {
	t.Helper()
	bs, err := io.ReadAll(testFile)
	assert.NilError(t, err)
	switch string(bs) {
	case "panic":
		return panicAssert
	default:
		slog.Debug("No assert done for " + testName)
	}
	return none
}

func runScriptAndAssert(t *testing.T, reproScript string, testType testType) {
	t.Helper()
	// (G204) this a directly controlled part of the repo in a test not a security risk
	cmd := exec.Command("/bin/bash", reproScript) //nolint:gosec
	stderr := strings.Builder{}
	cmd.Stderr = &stderr
	stdout := strings.Builder{}
	cmd.Stderr = &stdout
	cmd.Env = env.AddTo_PATH(os.Environ(), path)
	err := cmd.Run()
	switch testType {
	case none:
	case panicAssert:
		if err != nil {
			stdoutCompleted := stdout.String()
			assert.Assert(
				t,
				!strings.Contains(stdoutCompleted, "panic"),
				"Expected *NO* panic message from repro case: %s. Got:\n %s",
				reproScript,
				stdoutCompleted,
			)
		}
	}
}

type testType int

const (
	none testType = iota
	panicAssert
)
