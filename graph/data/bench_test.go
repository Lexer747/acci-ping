// Use of this source code is governed by a GPL-2 license that can be found in the LICENSE file.
//
// Copyright 2026 Lexer747
//
// SPDX-License-Identifier: GPL-2.0-only

package data_test

import (
	"os"
	"testing"
	"time"

	"github.com/Lexer747/acci-ping/graph/data"

	"gotest.tools/v3/assert"
)

// benchInput is the largest testdata capture (~60k points), exercising the point-iteration loops at scale.
const benchInput = "testdata/input/sixty-k.pings"

func loadBenchData(tb testing.TB) *data.Data {
	tb.Helper()
	raw, err := os.ReadFile(benchInput)
	assert.NilError(tb, err)
	d := &data.Data{}
	_, err = d.FromCompact(raw)
	assert.NilError(tb, err)
	return d
}

// BenchmarkDataIn is the baseline for [data.Data.In], which walks every point to re-timezone it.
func BenchmarkDataIn(b *testing.B) {
	d := loadBenchData(b)
	tz := time.UTC
	b.ReportAllocs()

	for b.Loop() {
		_ = d.In(tz)
	}
}
