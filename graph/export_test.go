// Use of this source code is governed by a GPL-2 license that can be found in the LICENSE file.
//
// Copyright 2024-2026 Lexer747
//
// SPDX-License-Identifier: GPL-2.0-only

package graph

import (
	"strings"

	"github.com/Lexer747/acci-ping/ping"
	"github.com/Lexer747/acci-ping/terminal"
	"github.com/Lexer747/acci-ping/utils/bytes"
	"github.com/Lexer747/acci-ping/utils/check"
)

// This file contains various helper methods for unit tests but which are not safe public API methods.

func (g *Graph) AddPoint(p ping.PingResults) {
	g.data.AddPoint(p)
}

func (g *Graph) ComputeFrame() string {
	var b strings.Builder
	painter := g.computeFrame(computeFrameConfig{
		followLatestSpan: false,
		drawSpinner:      false,
		yAxisScale:       g.presentation.Get().YAxisScale,
	})
	err := painter(&b)
	check.NoErr(err, "While painting frame to string buffer")
	return b.String()
}

func (g *Graph) Size() int64 {
	return g.data.TotalCount()
}

type XAxisSpanBounds struct {
	StartX, EndX, Width int
}

// ComputeXAxisBounds runs the internal x-axis layout and returns the per-span pixel bounds plus the axis size,
// letting tests assert layout invariants (e.g. the drawable area is fully used) without golden files.
func (g *Graph) ComputeXAxisBounds(s terminal.Size, following bool) []XAxisSpanBounds {
	g.data.Lock()
	defer g.data.Unlock()
	header := g.data.LockFreeHeader()
	iter := g.data.LockFreeIter(following)
	x := computeXAxis(
		bytes.NewConcurrentBuf(),
		bytes.NewConcurrentBuf(),
		s,
		header.TimeSpan,
		g.data.LockFreeSpanInfos(),
		following,
		int(iter.Total),
	)
	bounds := make([]XAxisSpanBounds, len(x.spans))
	for i, span := range x.spans {
		bounds[i] = XAxisSpanBounds{StartX: span.startX, EndX: span.endX, Width: span.width}
	}
	return bounds
}
