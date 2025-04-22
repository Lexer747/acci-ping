// Use of this source code is governed by a GPL-2 license that can be found in the LICENSE file.
//
// Copyright 2024-2025 Lexer747
//
// SPDX-License-Identifier: GPL-2.0-only

package graph

import (
	"bytes"
	"fmt"
	"time"

	"github.com/Lexer747/acci-ping/graph/data"
	"github.com/Lexer747/acci-ping/graph/graphdata"
	"github.com/Lexer747/acci-ping/gui/themes"
	"github.com/Lexer747/acci-ping/ping"
	"github.com/Lexer747/acci-ping/terminal"
	"github.com/Lexer747/acci-ping/terminal/ansi"
	"github.com/Lexer747/acci-ping/terminal/typography"
	"github.com/Lexer747/acci-ping/utils/numeric"
	"github.com/Lexer747/acci-ping/utils/sliceutils"
)

type drawingYAxis struct {
	size      int
	stats     *data.Stats
	labelSize int
}

type XAxisSpanInfo struct {
	spans     []*graphdata.SpanInfo
	spanStats *data.Stats
	pingStats *data.Stats
	timeSpan  *data.TimeSpan
	startX    int
	endX      int
	width     int
}

type drawingXAxis struct {
	size        int
	spans       []*XAxisSpanInfo
	overallSpan *data.TimeSpan
}

type xAxisIter struct {
	*drawingXAxis
	spanIndex int
}

func (x drawingXAxis) NewIter() *xAxisIter {
	return &xAxisIter{
		drawingXAxis: &x,
		spanIndex:    0,
	}
}

func (x *xAxisIter) Get(p ping.PingDataPoint) *XAxisSpanInfo {
	currentSpan := x.spans[x.spanIndex]
	if currentSpan.timeSpan.Contains(p.Timestamp) {
		return currentSpan
	}
	x.spanIndex++
	return x.Get(p)
}

func xAxisStartup() {
	padding = themes.Primary(typography.Horizontal)
	origin = themes.TitleHighlight(typography.Bullet) + " "
}

var padding string
var origin string

func computeXAxis(
	toWriteTo, toWriteSpanBars *bytes.Buffer,
	s terminal.Size,
	overall *data.TimeSpan,
	spans []*graphdata.SpanInfo,
	followLatestSpan bool,
	total int,
) drawingXAxis {
	space := s.Width - 6
	remaining := space
	// First add the initial dot for A E S T H E T I C S
	fmt.Fprint(toWriteTo, ansi.CursorPosition(s.Height, 1)+origin+padding+padding+padding+padding)

	var xAxisSpans []*XAxisSpanInfo
	if followLatestSpan {
		singleSpans := spans[len(spans)-1:]
		xAxisSpans = []*XAxisSpanInfo{
			{
				spans:     singleSpans,
				spanStats: singleSpans[0].SpanStats,
				pingStats: singleSpans[0].PingStats,
				timeSpan:  singleSpans[0].TimeSpan,
				startX:    1,
				endX:      s.Width,
				width:     s.Width,
			},
		}
		overall = singleSpans[0].TimeSpan
	} else {
		xAxisSpans = combineSpansPixelWise(spans, space, total)
		if space <= 1 {
			for _, span := range xAxisSpans {
				span.startX = 1
				span.endX = 1
			}
			return drawingXAxis{
				size:        s.Width,
				spans:       xAxisSpans,
				overallSpan: overall,
			}
		}
	}

	// Now we need to iterate every "span", where a span is some pre-determined gap in the pings which is
	// considered so large that we are reasonably confident that it was another recording session.
	//
	// In each iteration, we must determine the time in which the span lives and how much terminal space it
	// should take up. And then the actual values so that we actually plot against this axis accurately.
	for _, span := range xAxisSpans {
		span.startX = min(s.Width, s.Width-remaining)

		start, times := span.timeSpan.FormatDraw(span.width, 2)
		if len(times) < 1 {
			toCrop := max(min(span.width-2, len(start)-1), 0)
			cropped := start[:toCrop]
			remaining -= len(cropped) + 2
			fmt.Fprintf(toWriteTo, "%s", themes.Emphasis(cropped))
			toWriteTo.WriteString(padding + padding)
		} else {
			remaining -= len(start) + 4 + 2
			fmt.Fprint(toWriteTo, themes.Primary("[ ")+themes.Emphasis(start)+themes.Primary(" ]"))
			toWriteTo.WriteString(padding + padding)
			remaining = xAxisDrawTimes(toWriteTo, times, remaining, padding)
		}

		span.endX = min(s.Width, s.Width-remaining)
	}
	// Finally we add these vertical bars to indicate that the axis is not continuous and a new graph is
	// starting.
	if len(xAxisSpans) > 1 {
		addYAxisVerticalSpanIndicator(toWriteSpanBars, s, xAxisSpans)
	}
	return drawingXAxis{
		size:        s.Width,
		spans:       xAxisSpans,
		overallSpan: overall,
	}
}

// combineSpansPixelWise is a very crucial pre-processing step we need to do before drawing a frame, the data
// storage part [graphdata.GraphData] of the program will have made fairly sensible decisions about which
// parts of the data were actually recorded together. However this part of the program doesn't have the
// context about how much pixel real estate we can grant per recording session. Therefore we must do this
// every frame to determine which of this recording sessions must be merged for the sake of drawing. I.e. we
// have 5 recording sessions [*graphdata.SpanInfo], but the middle two are so short they would only take up 1
// pixel in the x-axis. This function has the agency to combine those middle spans when creating the
// [XAxisSpanInfo].
func combineSpansPixelWise(spans []*graphdata.SpanInfo, startingWidth, total int) []*XAxisSpanInfo {
	retSpans := make([]*XAxisSpanInfo, 0, len(spans))
	// TODO make this configurable - right now we just use a percentage of the start width or 5 when the
	// screen is small.
	minPixels := max(int(float64(startingWidth)*0.05), 5)
	acc := 0.0
	idx := 0
	for _, span := range spans {
		ratio := float64(span.Count) / (float64(total))
		width := int(float64(startingWidth) * ratio)
		if width >= minPixels && acc == 0.0 {
			retSpans = append(retSpans, &XAxisSpanInfo{
				spans:     []*graphdata.SpanInfo{span},
				spanStats: span.SpanStats,
				pingStats: span.PingStats,
				timeSpan:  span.TimeSpan,
				width:     width,
			})
			idx++
			continue
		}
		width = int(float64(startingWidth) * (acc + ratio))
		if width >= minPixels {
			retSpans[idx].spans = append(retSpans[idx].spans, span)
			retSpans[idx].spanStats = retSpans[idx].spanStats.Merge(span.SpanStats)
			retSpans[idx].pingStats = retSpans[idx].pingStats.Merge(span.PingStats)
			retSpans[idx].timeSpan = retSpans[idx].timeSpan.Merge(span.TimeSpan)
			retSpans[idx].width = width
			acc = 0.0
			idx++
			continue
		}
		if acc == 0.0 {
			retSpans = append(retSpans, &XAxisSpanInfo{
				spans:     []*graphdata.SpanInfo{span},
				spanStats: span.SpanStats,
				pingStats: span.PingStats,
				timeSpan:  span.TimeSpan,
			})
		} else {
			retSpans[idx].spans = append(retSpans[idx].spans, span)
			retSpans[idx].spanStats = retSpans[idx].spanStats.Merge(span.SpanStats)
			retSpans[idx].pingStats = retSpans[idx].pingStats.Merge(span.PingStats)
			retSpans[idx].timeSpan = retSpans[idx].timeSpan.Merge(span.TimeSpan)
		}
		acc += ratio
	}
	// TODO this width expanding finalizing still leaves some of the terminal unfilled, fix that.
	totalWidth := sliceutils.Fold(retSpans, 0, func(x *XAxisSpanInfo, acc int) int { return x.width + acc })
	delta := startingWidth - totalWidth
	toAdd := delta / len(retSpans)
	for _, span := range retSpans {
		span.width += toAdd
		totalWidth += toAdd
	}
	delta = startingWidth - totalWidth
	retSpans[len(retSpans)-1].width += delta
	return retSpans
}

func xAxisDrawTimes(b *bytes.Buffer, times []string, remaining int, padding string) int {
	for _, point := range times {
		if remaining <= len(point) {
			break
		}
		b.WriteString(themes.Highlight(point))
		remaining -= len(point)
		if remaining <= 1 {
			break
		}
		b.WriteString(padding)
		remaining--
		if remaining <= 1 {
			break
		}
		b.WriteString(padding)
		remaining--
	}
	return remaining
}

func getX(t time.Time, span *XAxisSpanInfo, y drawingYAxis, s terminal.Size) int {
	newMin := min(max(1, s.Width-1), span.endX)
	newMax := max(y.labelSize, span.startX)
	timestamp := span.timeSpan.End.Sub(t)
	if span.pingStats.GoodCount+span.spanStats.PacketsDropped <= 1 {
		// Edge case, when we have exactly one datum but the overall graph has more than one datum so we
		// didn't short-circuit and need to draw this one point in an arbitrary point between the spans start
		// and end.
		return int(numeric.NormalizeToRange(0.5, 0, 1, float64(newMin), float64(newMax)))
	}
	// These are inverted deliberately since the drawing reference is symmetric in the x
	computed := int(numeric.NormalizeToRange(
		float64(timestamp),
		0,
		float64(span.timeSpan.Duration),
		float64(newMin),
		float64(newMax),
	))
	return computed
}
