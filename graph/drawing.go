// Use of this source code is governed by a GPL-2 license that can be found in the LICENSE file.
//
// Copyright 2024-2025 Lexer747
//
// SPDX-License-Identifier: GPL-2.0-only

package graph

import (
	"io"
	"math"
	"strings"
	"time"

	"github.com/Lexer747/acci-ping/draw"
	"github.com/Lexer747/acci-ping/graph/data"
	"github.com/Lexer747/acci-ping/graph/gradient"
	"github.com/Lexer747/acci-ping/graph/graphdata"
	"github.com/Lexer747/acci-ping/gui/themes"
	"github.com/Lexer747/acci-ping/ping"
	"github.com/Lexer747/acci-ping/terminal"
	"github.com/Lexer747/acci-ping/terminal/ansi"
	"github.com/Lexer747/acci-ping/terminal/typography"
	"github.com/Lexer747/acci-ping/utils"
	"github.com/Lexer747/acci-ping/utils/bytes"
	"github.com/Lexer747/acci-ping/utils/errors"
	"github.com/Lexer747/acci-ping/utils/numeric"
)

type computeFrameConfig struct {
	timeBetweenFrames time.Duration
	followLatestSpan  bool
	drawSpinner       bool
	yAxisScale        YAxisScale
}

func (c computeFrameConfig) Match(cfg computeFrameConfig) bool {
	return c.followLatestSpan == cfg.followLatestSpan &&
		c.yAxisScale == cfg.yAxisScale
}

const drawingDebug = false

func getTimeBetweenFrames(fps int, pingsPerMinute ping.PingsPerMinute) time.Duration {
	if fps == 0 {
		return ping.PingsPerMinuteToDuration(pingsPerMinute)
	} else {
		return time.Duration(1000/fps) * time.Millisecond
	}
}

var noFrame = func(w io.Writer) error { return nil }

func (g *Graph) computeFrame(cfg computeFrameConfig) func(io.Writer) error {
	// This is race-y so ensure a consistent size for rendering, don't allow each sub-frame to re-compute the
	// size of the terminal. Side note - we deliberately don't attach this to the terminal size channel since
	// it's locked to the targeted FPS of this frame time anyway so just adds extra work.
	s := g.Term.GetSize()
	g.data.Lock()
	count := g.data.LockFreeTotalCount()
	spinnerValue := ""
	if cfg.drawSpinner {
		spinnerValue = g.lastFrame.spinnerData.spinner(s)
		g.drawingBuffer.Get(draw.SpinnerIndex).Reset()
		g.drawingBuffer.Get(draw.SpinnerIndex).WriteString(spinnerValue)
	}
	if count == g.lastFrame.PacketCount && g.lastFrame.Match(s, cfg) {
		g.data.Unlock() // fast path the frame didn't change
		if updateGui := g.checkGUI(); updateGui != nil {
			return updateGui
		}
		// Even faster path nothing at all too update
		return func(w io.Writer) error {
			return utils.Err(w.Write([]byte(spinnerValue)))
		}
	}
	if count == 0 {
		// nothing to do
		g.data.Unlock()
		return noFrame
	}

	g.drawingBuffer.Reset(draw.GraphIndexes...)

	header := g.data.LockFreeHeader()
	iter := g.data.LockFreeIter(cfg.followLatestSpan)
	x := computeXAxis(
		g.drawingBuffer.Get(draw.XAxisIndex),
		g.drawingBuffer.Get(draw.BarIndex),
		s,
		header.TimeSpan,
		g.data.LockFreeSpanInfos(),
		cfg.followLatestSpan,
		int(iter.Total),
	)
	yStats := header.Stats
	if cfg.followLatestSpan {
		yStats = x.spans[0].pingStats
	}
	y := computeYAxis(g.drawingBuffer.Get(draw.YAxisIndex), s, yStats, g.data.LockFreeURL(), cfg.yAxisScale)
	computeFrame(
		g,
		g.drawingBuffer.Get(draw.GradientIndex),
		g.drawingBuffer.Get(draw.DataIndex),
		g.drawingBuffer.Get(draw.DroppedIndex),
		g.drawingBuffer.Get(draw.KeyIndex),
		iter,
		g.data.LockFreeRuns(),
		x, y, s,
	)
	g.drawingBuffer.Get(draw.SpinnerIndex).WriteString(spinnerValue)
	// Everything we need is now cached we can unlock a bit early while we tidy up for the next frame
	paintFrame := withGUI(g.drawingBuffer)
	noGUI := withoutGUI(g.drawingBuffer)
	g.data.Unlock()
	g.lastFrame = frame{
		PacketCount:       count,
		yAxis:             y,
		xAxis:             x,
		spinnerData:       g.lastFrame.spinnerData,
		framePainter:      paintFrame,
		framePainterNoGui: noGUI,
		cfg:               cfg,
	}

	return paintFrame
}

func (g *Graph) checkGUI() func(io.Writer) error {
	state := g.ui.GetState()
	if state.ShouldDraw() && state.ShouldInvalidate() {
		return func(w io.Writer) error {
			defer g.ui.Drawn(state)
			return g.lastFrame.framePainter(w)
		}
	} else if state.ShouldDraw() {
		return func(w io.Writer) error {
			defer g.ui.Drawn(state)
			return errors.Join(
				onlyGUI(g.drawingBuffer)(w),
			)
		}
	} else if state.ShouldInvalidate() {
		return func(w io.Writer) error {
			defer g.ui.Drawn(state)
			return g.lastFrame.framePainter(w)
		}
	}
	return nil
}

func translate(p ping.PingDataPoint, x *XAxisSpanInfo, y drawingYAxis, s terminal.Size) (xCord, yCord int) {
	xCord = getX(p.Timestamp, x, y, s)
	yCord = getY(p.Duration, y, s)
	return
}

type gradientState struct {
	lastGoodSpan           *XAxisSpanInfo
	lastGoodIndex          int64
	lastGoodTerminalWidth  int
	lastGoodTerminalHeight int
}

func (g gradientState) set(i int64, x, y int, s *XAxisSpanInfo) gradientState {
	return gradientState{
		lastGoodIndex:          i,
		lastGoodTerminalWidth:  x,
		lastGoodTerminalHeight: y,
		lastGoodSpan:           s,
	}
}

func (g gradientState) dropped() gradientState {
	return gradientState{lastGoodIndex: -1}
}
func (g gradientState) draw() bool {
	return g.lastGoodIndex != -1
}

func computeFrame(
	g *Graph,
	toWriteGradientTo, toWriteTo, toWriteDroppedTo, toWriteKeyTo *bytes.SafeBuffer,
	iter *graphdata.Iter,
	runs *data.Runs,
	xAxis drawingXAxis,
	yAxis drawingYAxis,
	s terminal.Size,
) {
	if iter.Total < 1 {
		return
	}
	centreY := (s.Height - 2) / 2
	centreX := (s.Width - 2) / 2
	if iter.Total == 1 {
		point := iter.Get(0)
		toWriteTo.WriteString(ansi.CursorPosition(centreY, centreX) + single + " " + point.Duration.String())
		return
	}

	// Now iterate over all the individual data points and add them to the graph
	lastWasDropped := false
	lastDroppedTerminalX := -1
	window := newDrawWindow(s, len(xAxis.spans), g.debugStrict)
	xAxisIter := xAxis.NewIter()
	for i := range iter.Total {
		p := iter.Get(i)
		span := xAxisIter.Get(p)
		x := getX(p.Timestamp, span, yAxis, s)
		if p.Dropped() {
			window.addDroppedBar(x, s.Height, false)
			if lastWasDropped {
				for i := min(lastDroppedTerminalX, x) + 1; i < max(lastDroppedTerminalX, x); i++ {
					window.addDroppedBar(i, s.Height, true)
				}
			}
			lastWasDropped = true
			lastDroppedTerminalX = x
			continue
		}
		lastWasDropped = false
		y := getY(p.Duration, yAxis, s)
		g.checkf(
			x > 0 && x <= s.Width && y > 0 && y <= s.Height,
			"(x, y) {%d, %d} [%s, %s] coordinate out of terminal {%s} bounds. Index: %d",
			x, y,
			p.Timestamp, p.Duration,
			s, i,
		)
		window.addPoint(p, span.pingStats, yAxis.stats, span.width, x, y, centreX)
	}
	if shouldGradient(runs) {
		drawGradients(g, window, iter, xAxis, yAxis, s)
	}
	window.draw(toWriteTo, toWriteGradientTo, toWriteDroppedTo)
	toWriteKeyTo.WriteString(ansi.CursorPosition(s.Height-1, yAxis.labelSize+1))
	window.getKey(toWriteKeyTo)
}

func drawGradients(g *Graph, dw *drawWindow, iter *graphdata.Iter, xAxis drawingXAxis, yAxis drawingYAxis, s terminal.Size) {
	gs := gradientState{}
	xAxisIter := xAxis.NewIter()

	for i := range iter.Total {
		p := iter.Get(i)
		if p.Dropped() {
			gs = gs.dropped()
			continue
		}
		span := xAxisIter.Get(p)
		x, y := translate(p, span, yAxis, s)
		g.checkf(x <= s.Width && y <= s.Height, "Computed coord out of bounds (%d,%d) vs %q", x, y, s)
		if gs.draw() && !iter.IsLast(i) {
			if span == gs.lastGoodSpan {
				lastP := iter.Get(gs.lastGoodIndex)
				drawGradient(
					g, dw,
					span, yAxis,
					x, y,
					p,
					lastP,
					gs.lastGoodTerminalWidth,
					gs.lastGoodTerminalHeight,
					s,
				)
			}
		}
		gs = gs.set(i, x, y, span)
	}
}

// A bar requires you start at the top of the terminal, general to draw a bar at coord x, do
// [ansi.CursorPosition(2, x)] before writing the bar.
func makeBar(repeating string, s terminal.Size, touchAxis bool) string {
	offset := 3
	if touchAxis {
		offset = 2
	}
	size := s.Height - offset
	if size < 0 {
		return ""
	}
	return strings.Repeat(repeating+ansi.CursorDown(1)+ansi.CursorBack(1), size)
}

func drawGradient(
	g *Graph,
	dw *drawWindow,
	xAxis *XAxisSpanInfo,
	yAxis drawingYAxis,
	x, y int,
	current ping.PingDataPoint,
	lastGood ping.PingDataPoint,
	lastGoodTerminalWidth int,
	lastGoodTerminalHeight int,
	s terminal.Size,
) {
	gradientsToDrawX := float64(numeric.Abs(lastGoodTerminalWidth - x))
	gradientsToDrawY := float64(numeric.Abs(lastGoodTerminalHeight - y))
	gradientsToDraw := math.Sqrt((gradientsToDrawX * gradientsToDrawX) + (gradientsToDrawY * gradientsToDrawY))
	stepSizeY := float64(current.Duration-lastGood.Duration) / gradientsToDraw
	stepSizeX := float64(current.Timestamp.Sub(lastGood.Timestamp)) / gradientsToDraw

	pointsX := make([]int, 0)
	pointsY := make([]int, 0)
	for toDraw := 1.5; toDraw < gradientsToDraw; toDraw++ {
		interpolatedDuration := lastGood.Duration + time.Duration(toDraw*stepSizeY)
		interpolatedStamp := lastGood.Timestamp.Add(time.Duration(toDraw * stepSizeX))
		p := ping.PingDataPoint{Duration: interpolatedDuration, Timestamp: interpolatedStamp}
		cursorX, cursorY := translate(p, xAxis, yAxis, s)
		g.checkf(cursorX <= s.Width && cursorY <= s.Height, "Computed coord out of bounds (%d,%d) vs %q", cursorY, cursorX, s)
		pointsX = append(pointsX, cursorX)
		pointsY = append(pointsY, cursorY)
	}
	gradients := gradient.Solve(pointsX, pointsY)
	for i, gradient := range gradients {
		dw.addGradient(pointsX[i], pointsY[i], gradient)
	}
}

func shouldGradient(runs *data.Runs) bool {
	return runs.GoodPackets.Longest > 2
}

func makeTitle(toWriteTo *bytes.SafeBuffer, size terminal.Size, stats *data.Stats, url string) {
	const yAxisTitle = "Ping "
	sizeStr := size.String()
	titleBegin := themes.Emphasis(url)
	titleEnd := themes.Positive(sizeStr)
	remaining := size.Width - len(yAxisTitle) - len(url) - len(sizeStr)
	statsStr := stats.PickString(remaining)
	if len(statsStr) > 0 {
		statsStr = " [" + themes.Primary(statsStr) + "] "
	}
	title := titleBegin + statsStr + titleEnd
	titleIndent := (size.Width / 2) - (len(title) / 2)
	toWriteTo.WriteString(
		ansi.Home + themes.TitleHighlight(yAxisTitle) + ansi.CursorForward(titleIndent) + title,
	)
	if drawingDebug {
		toWriteTo.WriteString(ansi.CursorPosition(1, size.Width-1) + themes.DarkNegative(typography.LightBlock))
	}
}

// withoutGUI knows how to composite the parts of a frame and the spinner, returning a lambda which will draw
// the computed frame to the given writer, with no GUI elements.
func withoutGUI(toDraw *draw.Buffer) func(io.Writer) error {
	return painter(toDraw, true, draw.GraphIndexes)
}

// withGUI knows how to composite the parts of a frame and the spinner, returning a lambda which will draw the
// computed frame to the given writer.
func withGUI(toDraw *draw.Buffer) func(io.Writer) error {
	return painter(toDraw, true, draw.PaintOrder)
}

func onlyGUI(toDraw *draw.Buffer) func(io.Writer) error {
	return painter(toDraw, false, draw.GUIIndexes)
}

func painter(toDraw *draw.Buffer, clearFrame bool, indexes []draw.Index) func(io.Writer) error {
	return func(toWriteTo io.Writer) error {
		if clearFrame {
			// First clear the screen from the last frame
			err := utils.Err(toWriteTo.Write([]byte(ansi.Clear)))
			if err != nil {
				return err
			}
		}

		// Now in paint order, simply forward the bytes onto the writer
		for _, i := range indexes {
			err := utils.Err(toWriteTo.Write(toDraw.Get(i).Bytes()))
			if err != nil {
				return err
			}
		}
		return nil
	}
}
