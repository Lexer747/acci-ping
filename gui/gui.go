// Use of this source code is governed by a GPL-2 license that can be found in the LICENSE file.
//
// Copyright 2025 Lexer747
//
// SPDX-License-Identifier: GPL-2.0-only

package gui

import (
	"bytes"
	"strconv"
	"sync"

	"github.com/Lexer747/acci-ping/terminal"
	"github.com/Lexer747/acci-ping/utils/check"
)

// GUI is the abstract notion of all the components which make up the GUI, providing two methods, the "state"
// of the GUI informs a callee about whether this GUI would like to be drawn [Token.ShouldDraw]. And a further
// clarifying [Token.ShouldInvalidate] which tells the callee that the underlying GUI believes the components
// have changed and thus need clearing from the output, with nothing to be drawn.
type GUI interface {
	GetState() Token
	Drawn(t Token)
	Paint(update PaintUpdate)
}

// GUIState is a concrete implementation of GUI providing facilities to invalidate and redraw as required.
type GUIState struct {
	m *sync.Mutex

	// We track all of these arbitrary id's so that we can hand them out to the callee of this GUI. We choose
	// a monotonic implementation so that we can easily determine which token we should expect back.

	paintIdx, drawnPaintIdx           uint64
	invalidateIdx, drawnInvalidateIdx uint64
}

func NewGUIState() *GUIState {
	return &GUIState{m: &sync.Mutex{}}
}

// PaintUpdate represents what about a paint update has occurred, it's a bit field so should be bitwise OR'd
// to combine operations.
type PaintUpdate int

const (
	None       PaintUpdate = 0b000000000000000
	Paint      PaintUpdate = 0b000000000000001
	Invalidate PaintUpdate = 0b000000000000010 // TODO invalidate invalidates all components but should only per remove GUI element
)

// Paint tells the GUI that a certain update is required.
func (g *GUIState) Paint(update PaintUpdate) {
	if update == None {
		return
	}
	g.m.Lock()
	defer g.m.Unlock()
	// If we have painted then increment
	if (update & Paint) != None {
		g.paintIdx++
	}
	// If we have invalidate then increment
	if (update & Invalidate) != None {
		g.invalidateIdx++
	}
}

// Drawn implements [g.GUI]. In this case our GUI will be informed by the callee which token was actually
// consumed and so we can use this in future to determine if we need to draw/invalidate still.
func (g *GUIState) Drawn(t Token) {
	token, ok := t.(*paintToken)
	check.Check(ok, "should only be called with original gui token")
	g.drawnInvalidateIdx = token.invalidateIdx
	g.drawnPaintIdx = token.paintIdx
}

// GetState implements [GUI].
func (g *GUIState) GetState() Token {
	g.m.Lock()
	defer g.m.Unlock()
	return &paintToken{
		paintIdx:            g.paintIdx,
		drawnPaintIdx:       g.drawnPaintIdx,
		invalidateIdx:       g.invalidateIdx,
		drawnInvalidatedIdx: g.drawnInvalidateIdx,
	}
}

// paintToken can easily determine if we need to draw or invalidate because we know indexes are monotonic if
// current index is larger than the last seen index from [Drawn] then we know we need to do that action.
type paintToken struct {
	paintIdx, drawnPaintIdx            uint64
	invalidateIdx, drawnInvalidatedIdx uint64
}

func (p *paintToken) ShouldDraw() bool {
	return p.paintIdx > p.drawnPaintIdx
}

func (p *paintToken) ShouldInvalidate() bool {
	return p.invalidateIdx > p.drawnInvalidatedIdx
}

var _ Token = (&paintToken{})

// Token represents the state of the GUI and can be queried about what it desires. If [ShouldDraw] returns
// true the GUI would like to be drawn. [ShouldInvalidate] however returns true if the component should be
// removed from the canvas.
//
// If both return false the GUI has no work to be done and last known frame is correct.
//
// Once the callee has rendered the GUI it should return this token back to the [GUI] with [GUI.Drawn] so that
// the GUI can keep track of what frame was drawn.
type Token interface {
	ShouldDraw() bool
	ShouldInvalidate() bool
}

// Draw is the high level interface that any GUI component should implement which will draw itself to the byte
// buffer.
type Draw interface {
	Draw(size terminal.Size, b *bytes.Buffer)
}

var _ Draw = (&Box{})
var _ Draw = (&initialisedTypography{})

// A Position contains all the GUI data required to know where a GUI component should be drawn.
type Position struct {
	Vertical   VerticalAlignment
	Horizontal HorizontalAlignment
	Padding    Padding
}

func (p Position) String() string {
	return "{Vertical: " + p.Vertical.String() +
		" Horizontal: " + p.Horizontal.String() +
		" Padding: " + p.Padding.String() + "}"
}

// Padding represents in pixels how many pixels should be offset in the given direction.
type Padding struct {
	Top, Bottom, Left, Right int
}

func (p Padding) String() string {
	return "T:" + strconv.Itoa(p.Top) +
		" B:" + strconv.Itoa(p.Bottom) +
		" L:" + strconv.Itoa(p.Left) +
		" R:" + strconv.Itoa(p.Right)
}

func (p Padding) Equal(other Padding) bool {
	return p.Top == other.Top &&
		p.Bottom == other.Bottom &&
		p.Left == other.Left &&
		p.Right == other.Right
}

var NoPadding Padding = Padding{}

type HorizontalAlignment int
type VerticalAlignment int

const (
	Left   HorizontalAlignment = 1
	Centre HorizontalAlignment = 2
	Right  HorizontalAlignment = 3

	Top    VerticalAlignment = 1
	Middle VerticalAlignment = 2
	Bottom VerticalAlignment = 3
)

func (a HorizontalAlignment) String() string {
	switch a {
	case Left:
		return "Left"
	case Centre:
		return "Centre"
	case Right:
		return "Right"
	default:
		return "Unknown Alignment: " + strconv.Itoa(int(a))
	}
}

func (a VerticalAlignment) String() string {
	switch a {
	case Top:
		return "Top"
	case Middle:
		return "Middle"
	case Bottom:
		return "Bottom"
	default:
		return "Unknown Alignment: " + strconv.Itoa(int(a))
	}
}

func (p PaintUpdate) String() string {
	switch p {
	case None:
		return "None"
	case Paint:
		return "Paint"
	case Invalidate:
		return "Invalidate"
	}
	panic("exhaustive:enforce")
}
