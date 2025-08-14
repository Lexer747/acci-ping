// Use of this source code is governed by a GPL-2 license that can be found in the LICENSE file.
//
// Copyright 2024-2025 Lexer747
//
// SPDX-License-Identifier: GPL-2.0-only

package graph

import (
	"fmt"
	"math"
	"time"

	"github.com/Lexer747/acci-ping/graph/data"
	"github.com/Lexer747/acci-ping/gui/themes"
	"github.com/Lexer747/acci-ping/terminal"
	"github.com/Lexer747/acci-ping/terminal/ansi"
	"github.com/Lexer747/acci-ping/terminal/typography"
	"github.com/Lexer747/acci-ping/utils/bytes"
	"github.com/Lexer747/acci-ping/utils/numeric"
	"github.com/Lexer747/acci-ping/utils/timeutils"
)

type drawingYAxis struct {
	size      int
	stats     *data.Stats
	labelSize int
	scale     YAxisScale
}

type YAxisScale byte

const (
	Linear      YAxisScale = 0
	Logarithmic YAxisScale = 1
)

func (y YAxisScale) String() string {
	switch y {
	case Linear:
		return "Linear"
	case Logarithmic:
		return "Logarithmic"
	default:
		panic("exhaustive:enforce")
	}
}

func yAxisStartup() {
	spanBar = themes.Emphasis(typography.DoubleVertical)
}

var spanBar string

func addYAxisVerticalSpanIndicator(bars *bytes.SafeBuffer, s terminal.Size, spans []*XAxisSpanInfo) {
	spanSeparator := makeBar(spanBar, s, true)
	// Don't draw the last span since this is implied by the end of the terminal
	for _, span := range spans[:len(spans)-1] {
		if span.endX >= (s.Width - 1) {
			continue
			// Don't draw on-top of the y-axis
		}
		bars.WriteString(ansi.CursorPosition(2, span.endX+1) + spanSeparator)
	}
	// Reset the cursor back to the start of the axis
	bars.WriteString(ansi.CursorPosition(s.Height, 1))
}

func computeYAxis(
	toWriteTo *bytes.SafeBuffer,
	size terminal.Size,
	stats *data.Stats,
	url string,
	scale YAxisScale,
) drawingYAxis {
	toWriteTo.Grow(size.Height)

	makeTitle(toWriteTo, size, stats, url)

	gapSize := 2
	if size.Height > 20 {
		gapSize++
	}
	durationSize := (gapSize * 3) / 2

	// We skip the first and last two lines
	for i := range size.Height - 3 {
		h := i + 2
		fmt.Fprint(toWriteTo, ansi.CursorPosition(h, 1))
		if i%gapSize == 1 {
			var scaledDuration float64
			switch scale {
			case Linear:
				scaledDuration = numeric.NormalizeToRange(float64(i+1), float64(size.Height-3), 2, float64(stats.Min), float64(stats.Max))
			case Logarithmic:
				logMin := math.Log(float64(stats.Min))
				logMax := math.Log(float64(stats.Max))
				logScaledTime := numeric.NormalizeToRange(float64(i+1), float64(size.Height-3), 2, logMin, logMax)
				scaledDuration = math.Pow(math.E, logScaledTime)
			}
			toPrint := timeutils.HumanString(time.Duration(scaledDuration), durationSize)
			fmt.Fprint(toWriteTo, themes.Highlight(toPrint))
		} else {
			fmt.Fprint(toWriteTo, themes.Primary(typography.Vertical))
		}
	}
	// Last line is always a bar
	fmt.Fprint(toWriteTo, ansi.CursorPosition(max(1, size.Height-1), 1)+themes.Primary(typography.Vertical))
	return drawingYAxis{
		size:      size.Height,
		stats:     stats,
		labelSize: min(durationSize+4, size.Width),
		scale:     scale,
	}
}

func getY(dur time.Duration, yAxis drawingYAxis, s terminal.Size) int {
	newMin := max(1, s.Height-2)
	newMax := min(2, s.Height)
	switch yAxis.scale {
	case Linear:
		return int(numeric.NormalizeToRange(
			float64(dur),
			float64(yAxis.stats.Min),
			float64(yAxis.stats.Max),
			float64(newMin),
			float64(newMax),
		))
	case Logarithmic:
		return int(numeric.NormalizeToRange(
			math.Log(float64(dur)),
			math.Log(float64(yAxis.stats.Min)),
			math.Log(float64(yAxis.stats.Max)),
			float64(newMin),
			float64(newMax),
		))
	}
	panic("exhaustive:enforce")
}
