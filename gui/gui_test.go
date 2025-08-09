// Use of this source code is governed by a GPL-2 license that can be found in the LICENSE file.
//
// Copyright 2025 Lexer747
//
// SPDX-License-Identifier: GPL-2.0-only

package gui_test

import (
	"testing"

	"github.com/Lexer747/acci-ping/gui"
	"gotest.tools/v3/assert"
)

func TestTrivialGuiInterface(t *testing.T) {
	t.Parallel()
	guis := []gui.GUI{
		gui.NoGUI(),
		gui.NewGUIState(),
	}
	for _, underTest := range guis {
		state := underTest.GetState()
		assert.Equal(t, state.ShouldDraw(), false)
		assert.Equal(t, state.ShouldInvalidate(), false)
		underTest.Drawn(state)
		underTest.Paint(gui.None)
		state2 := underTest.GetState()
		assert.Equal(t, state2.ShouldDraw(), false)
		assert.Equal(t, state2.ShouldInvalidate(), false)
	}
}

// This test would never pass with NoGUI.
func TestBasicGui(t *testing.T) {
	t.Parallel()
	underTest := gui.NewGUIState()
	{
		// paint nothing
		underTest.Paint(gui.None)
		state := underTest.GetState()
		assert.Equal(t, state.ShouldDraw(), false)
		assert.Equal(t, state.ShouldInvalidate(), false)
	}
	{
		// paint something
		underTest.Paint(gui.Paint)
		state := underTest.GetState()
		assert.Equal(t, state.ShouldDraw(), true)
		assert.Equal(t, state.ShouldInvalidate(), false)
	}
	{
		// invalidate something
		underTest.Paint(gui.Invalidate)
		state := underTest.GetState()
		assert.Equal(t, state.ShouldDraw(), true, "we haven't drawn yet so this should still be true")
		assert.Equal(t, state.ShouldInvalidate(), true)
	}
	{
		// paint nothing
		underTest.Paint(gui.None)
		state := underTest.GetState()
		assert.Equal(t, state.ShouldDraw(), true, "we haven't drawn yet so this should still be true")
		assert.Equal(t, state.ShouldInvalidate(), true, "we haven't invalidated yet so this should still be true")
	}
	// do a paint, cheating and simply using the current state
	underTest.Drawn(underTest.GetState())
	newState := underTest.GetState()
	assert.Equal(t, newState.ShouldDraw(), false, "paint should be completed")
	assert.Equal(t, newState.ShouldInvalidate(), false, "invalidate should be completed")
}

func TestStateTrackingGui(t *testing.T) {
	t.Parallel()
	underTest := gui.NewGUIState()
	underTest.Paint(gui.Paint)
	stateBeforeInvalidate := underTest.GetState()
	underTest.Paint(gui.Paint | gui.Invalidate)
	// Notice that we haven't acquired a new state after calling paint with the new flags
	underTest.Drawn(stateBeforeInvalidate)
	assert.Equal(t, underTest.GetState().ShouldInvalidate(), true, "invalidate is not be completed")
}
