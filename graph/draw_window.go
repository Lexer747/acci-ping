// Use of this source code is governed by a GPL-2 license that can be found in the LICENSE file.
//
// Copyright 2024-2025 Lexer747
//
// SPDX-License-Identifier: GPL-2.0-only

package graph

import (
	"bytes"
	"fmt"
	"log/slog"

	"github.com/Lexer747/acci-ping/graph/data"
	"github.com/Lexer747/acci-ping/graph/terminal/ansi"
	"github.com/Lexer747/acci-ping/graph/terminal/typography"
	"github.com/Lexer747/acci-ping/ping"
	"github.com/Lexer747/acci-ping/utils/check"
)

// drawWindow is an optimiser and beautifier for the actual points being drawn to a given frame. The
// optimisation is that instead of creating more terminal output by painting over dots we cache all the
// results into this buffer, then draw the cache after iterating all the points.
//
// This also directly enables not painting over labels and the other span based printing choosing which points
// to highlight.
type drawWindow struct {
	cache       map[coords]drawnData
	labels      []label
	debugStrict bool
	max         int
}

// coords are the unique key to identify some data to be drawn
type coords struct {
	x, y int
}

// drawnData is the actual data we wish to draw, [isLabel] is an indirect pointer of sorts which tells the
// overall library to look at the [drawWindow.labels] instead this draw data.
type drawnData struct {
	pingCount int
	isLabel   bool
}

type label struct {
	coords
	symbol      string
	text        string
	leftJustify bool
	colour      colour
}

type colour int

const (
	red colour = iota
	green
)

func newDrawWindow(total int64, spans int, debugStrict bool) *drawWindow {
	return &drawWindow{
		cache:       make(map[coords]drawnData, int(total)),
		labels:      make([]label, 0, spans*2),
		debugStrict: debugStrict,
	}
}

func (dw *drawWindow) draw(toWriteTo *bytes.Buffer) {
	// These can be indeterministically (map order) drawn since we guarantee uniqueness of the coords,
	// therefore meaning no map [drawnData] will ever contain the same coords which have different ping counts
	for c, point := range dw.cache {
		if point.isLabel {
			// labels are drawn separately, but are in the cache for [addPoint]
			continue
		}
		toWriteTo.WriteString(ansi.CursorPosition(c.y, c.x) + dw.getOverlap(c.x, c.y))
	}
	// If these are drawn indeterministically then we will get shimmer as labels may be fighting for Z-Preference
	for _, l := range dw.labels {
		var addColour func(string) string
		if l.colour == red {
			addColour = ansi.Red
		} else {
			addColour = ansi.Green
		}
		if l.leftJustify {
			// ensure that we don't write to a negative coord should the terminal be very small.
			xPos := max(1, l.x-len(l.text))
			toWriteTo.WriteString(ansi.CursorPosition(l.y, xPos) + addColour(l.text+" "+l.symbol))
		} else /* rightJustify */ {
			toWriteTo.WriteString(ansi.CursorPosition(l.y, l.x) + addColour(l.symbol+" "+l.text))
		}
	}
}

// e.g. 22.12434ms, 8.359131ms, 7.406686ms
const averageLabelSize = 30

func (dw *drawWindow) addPoint(
	p ping.PingDataPoint,
	spanStats, stats *data.Stats,
	spanWidth int,
	x, y, centreX int,
) {
	isMin := p.Duration == stats.Min
	isMax := p.Duration == stats.Max
	isMinWithinSpan := p.Duration == spanStats.Min
	isMaxWithinSpan := p.Duration == spanStats.Max
	wideEnough := spanWidth > averageLabelSize
	needsLabel := (wideEnough && (isMinWithinSpan || isMaxWithinSpan)) || isMin || isMax
	dw.add(x, y, needsLabel)
	if !needsLabel {
		return
	}
	leftJustify := x > centreX
	var symbol string
	var colour colour
	if isMinWithinSpan {
		colour = green
		symbol = typography.HollowUpTriangle
	}
	if isMaxWithinSpan {
		colour = red
		symbol = typography.HollowDownTriangle
	}
	if isMin {
		colour = green
		symbol = typography.FilledUpTriangle
	}
	if isMax {
		colour = red
		symbol = typography.FilledDownTriangle
	}
	if needsLabel {
		label := p.Duration.String()
		dw.addLabel(x, y, leftJustify, symbol, label, colour)
	}
}

func (dw *drawWindow) add(x, y int, label bool) {
	dw.checkf(x > 0 && y > 0, "(x, y): {%d, %d} being added to draw window out of bounds", x, y)
	c := coords{x, y}
	if drawData, found := dw.cache[c]; found {
		if drawData.isLabel {
			// Don't double count label, labels only need a count of 1 to be drawn
			return
		}
		count := drawData.pingCount + 1
		dw.cache[c] = drawnData{
			pingCount: count,
			isLabel:   drawData.isLabel || label,
		}
		dw.max = max(count, dw.max)
	} else {
		dw.cache[c] = drawnData{
			pingCount: 1,
			isLabel:   label,
		}
	}
}

func (dw *drawWindow) checkf(shouldBeTrue bool, format string, a ...any) {
	if dw.debugStrict {
		check.Checkf(shouldBeTrue, format, a...)
	} else if !shouldBeTrue {
		slog.Error("check failed: " + fmt.Sprintf(format, a...))
	}
}

// addLabel will spread over the drawWindow all the coords which the label will occupy this ensures we draw on
// top of data points and the text is legible.
func (dw *drawWindow) addLabel(x, y int, leftJustify bool, symbol, labelStr string, colour colour) {
	c := coords{x, y}
	dir := 1
	if leftJustify {
		dir = -1
	}
	for i := range len(labelStr) {
		extendedX := (x + 2) + (dir * i)
		if extendedX == x {
			// Don't double count the point itself
			continue
		}
		if extendedX < 1 {
			break // don't add label text outside the coordinate system
		}
		dw.add(extendedX, y, true)
	}
	dw.labels = append(dw.labels, label{
		coords:      c,
		symbol:      symbol,
		text:        labelStr,
		leftJustify: leftJustify,
		colour:      colour,
	})
}

const (
	fewThreshold   = 1
	manyThreshold  = 5
	loadsThreshold = 25
)

var (
	single = ansi.White(typography.Multiply)
	few    = ansi.White(typography.SmallSquare)
	many   = ansi.White(typography.Diamond)
	loads  = ansi.White(typography.Square)

	bar = ansi.Gray("|")
)

func (dw *drawWindow) getOverlap(x, y int) string {
	c := coords{x, y}
	dd := dw.cache[c]
	switch {
	case dd.pingCount <= fewThreshold:
		return single
	case dd.pingCount <= manyThreshold:
		return few
	case dd.pingCount <= loadsThreshold:
		return many
	default:
		return loads
	}
}

// getKey will write to the draw buffer the key needed for this draw window, where is minimizes the amount of
// text needed to show the key for all the points drawn.
func (dw *drawWindow) getKey(toWriteTo *bytes.Buffer) {
	if dw.max > loadsThreshold {
		fmt.Fprintf(toWriteTo, ansi.Gray("Key")+ansi.White(": ")+
			single+" = %d "+bar+" "+few+" = %d-%d "+bar+" "+many+" = %d-%d "+bar+" "+loads+" = %d+    ",
			fewThreshold, fewThreshold+1, manyThreshold, manyThreshold+1, loadsThreshold, loadsThreshold+1)
		return
	}
	if dw.max > manyThreshold {
		fmt.Fprintf(toWriteTo, ansi.Gray("Key")+ansi.White(": ")+
			single+" = %d "+bar+" "+few+" = %d-%d "+bar+" "+many+" = %d-%d    ",
			fewThreshold, fewThreshold+1, manyThreshold, manyThreshold+1, loadsThreshold)
		return
	}
	if dw.max > fewThreshold {
		fmt.Fprintf(toWriteTo, ansi.Gray("Key")+ansi.White(": ")+
			single+" = %d "+bar+" "+few+" = %d-%d    ",
			fewThreshold, fewThreshold+1, manyThreshold)
		return
	}
}
