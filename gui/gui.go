// Use of this source code is governed by a GPL-2 license that can be found in the LICENSE file.
//
// Copyright 2025 Lexer747
//
// SPDX-License-Identifier: GPL-2.0-only

package gui

import (
	"bytes"
	"strconv"

	"github.com/Lexer747/acci-ping/terminal"
)

// GUI is the abstract notion of all the components which make up the GUI, providing two methods, the "state"
// of the GUI informs a callee about whether this GUI would like to be drawn [Token.ShouldDraw]. And a further
// clarifying [Token.ShouldInvalidate] which tells the callee that the underlying GUI believes the components
// have changed and thus need clearing from the output, with nothing to be drawn.
//
// See the concrete implementation used by acci-ping at [cmd/subcommands/acci-ping/gui.go]
type GUI interface {
	GetState() Token
	Drawn(t Token)
}

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
