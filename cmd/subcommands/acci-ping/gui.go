// Use of this source code is governed by a GPL-2 license that can be found in the LICENSE file.
//
// Copyright 2024-2025 Lexer747
//
// SPDX-License-Identifier: GPL-2.0-only

package acciping

import (
	"sync"

	"github.com/Lexer747/acci-ping/gui"
	"github.com/Lexer747/acci-ping/terminal"
	"github.com/Lexer747/acci-ping/utils/check"
)

type GUI struct {
	listeningChars map[rune]terminal.ConditionalListener
	fallbacks      []terminal.Listener

	m sync.Mutex

	// We track all of these arbitrary id's so that we can hand them out to the callee of this GUI. We choose
	// a monotonic implementation so that we can easily determine which token we should expect back.

	paintIdx, drawnPaintIdx           uint64
	invalidateIdx, drawnInvalidateIdx uint64
}

var _ gui.GUI = (&GUI{})

func newGUIState() *GUI {
	return &GUI{
		listeningChars: map[rune]terminal.ConditionalListener{},
		fallbacks:      []terminal.Listener{},
	}
}

type paintUpdate int

const (
	None       paintUpdate = 0b000000000000000
	Paint      paintUpdate = 0b000000000000001
	Invalidate paintUpdate = 0b000000000000010 // TODO invalidate invalidates all components but should only per remove GUI element
)

func (p paintUpdate) String() string {
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

func (g *GUI) paint(update paintUpdate) {
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

// Drawn implements [gui.GUI]. In this case our GUI will be informed by the callee which token was actually
// consumed and so we can use this in future to determine if we need to draw/invalidate still.
func (g *GUI) Drawn(t gui.Token) {
	token, ok := t.(*paintToken)
	check.Check(ok, "should only be called with original gui token")
	g.drawnInvalidateIdx = token.invalidateIdx
	g.drawnPaintIdx = token.paintIdx
}

// GetState implements [gui.GUI].
func (g *GUI) GetState() gui.Token {
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

var _ gui.Token = (&paintToken{})
