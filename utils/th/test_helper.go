// Use of this source code is governed by a GPL-2 license that can be found in the LICENSE file.
//
// Copyright 2024-2025 Lexer747
//
// SPDX-License-Identifier: GPL-2.0-only

// th stands for "test helper"
package th

import (
	"reflect"

	"github.com/Lexer747/acci-ping/utils/numeric"
	"github.com/google/go-cmp/cmp"
	"gotest.tools/v3/assert"
	is "gotest.tools/v3/assert/cmp"
	"pgregory.net/rapid"
)

func AssertFloatEqual(t T, expected float64, actual float64, sigFigs int, msgAndArgs ...any) {
	t.Helper()
	a := numeric.RoundToNearestSigFig(actual, sigFigs)
	e := numeric.RoundToNearestSigFig(expected, sigFigs)
	assert.Check(t, is.Equal(e, a), msgAndArgs...)
}

var AllowAllUnexported = cmp.Exporter(func(reflect.Type) bool { return true })

// T is the "current" acci-ping most generic test interface for the most use with all test frameworks and
// third party helpers. [*testing.T] is safer and easier to use if in doubt, you will know when you need this
// helper because certain cross dependency tests have stopped compiling.
type T interface {
	rapid.TB
}
