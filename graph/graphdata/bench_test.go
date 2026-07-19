// Use of this source code is governed by a GPL-2 license that can be found in the LICENSE file.
//
// Copyright 2026 Lexer747
//
// SPDX-License-Identifier: GPL-2.0-only

package graphdata_test

import (
	"os"
	"testing"

	"github.com/Lexer747/acci-ping/graph/data"
	"github.com/Lexer747/acci-ping/graph/graphdata"

	"gotest.tools/v3/assert"
)

// benchInput is the largest testdata capture (~60k points), exercising the point-iteration loops at scale.
const benchInput = "../data/testdata/input/sixty-k.pings"

func loadBenchData(tb testing.TB) *data.Data {
	tb.Helper()
	raw, err := os.ReadFile(benchInput)
	assert.NilError(tb, err)
	d := &data.Data{}
	_, err = d.FromCompact(raw)
	assert.NilError(tb, err)
	return d
}

// BenchmarkNewGraphData is the baseline for the ingestion loop, which walks every point via
// data.Get to build up the spans.
func BenchmarkNewGraphData(b *testing.B) {
	d := loadBenchData(b)
	b.ReportAllocs()

	for b.Loop() {
		_ = graphdata.NewGraphData(d)
	}
}
