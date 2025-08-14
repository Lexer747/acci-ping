// Use of this source code is governed by a GPL-2 license that can be found in the LICENSE file.
//
// Copyright 2025 Lexer747
//
// SPDX-License-Identifier: GPL-2.0-only

package acciping

import (
	"context"

	"github.com/Lexer747/acci-ping/draw"
	"github.com/Lexer747/acci-ping/graph"
	"github.com/Lexer747/acci-ping/gui"
	"github.com/Lexer747/acci-ping/terminal"
	"github.com/Lexer747/acci-ping/utils/bytes"
)

// showControls which should only be called once the paint buffer is initialised.
func (app *Application) showControls(
	ctx context.Context,
	initialValues graph.Presentation,
	fromTerminal <-chan graph.Control,
	terminalSizeUpdates <-chan terminal.Size,
) {
	buffer := app.drawBuffer.Get(draw.ControlIndex)
	c := controlState{Presentation: initialValues}
	app.GUIState.Paint(c.render(app.term.GetSize(), buffer))
	for {
		select {
		case <-ctx.Done():
			return
		case newSize := <-terminalSizeUpdates:
			app.GUIState.Paint(c.render(newSize, buffer))
		case update := <-fromTerminal:
			if update.FollowLatestSpan.DidChange {
				c.Following = update.FollowLatestSpan.Value
			}
			if update.YAxisScale.DidChange {
				c.YAxisScale = update.YAxisScale.Value
			}
			app.GUIState.Paint(c.render(app.term.GetSize(), buffer))
		}
	}
}

type controlState struct {
	graph.Presentation
}

func (c controlState) render(size terminal.Size, buf *bytes.SafeBuffer) gui.PaintUpdate {
	ret := gui.None
	buf.Reset()
	if !c.Following || c.YAxisScale != graph.Logarithmic {
		ret = ret | gui.Invalidate
	}
	paint := false
	var box gui.Box
	switch {
	case c.Following && c.YAxisScale == graph.Logarithmic:
		box = c.makeFollowingAndLogarithmicBox()
		paint = true
	case c.Following:
		box = c.makeFollowingBox()
		paint = true
	case c.YAxisScale == graph.Logarithmic:
		box = c.makeLogarithmicBox()
		paint = true
	}
	if paint {
		box.Draw(size, buf)
		return ret | gui.Paint
	}
	return ret
}

var (
	following = gui.Typography{
		ToPrint:        "Following",
		LenFromToPrint: true,
		Alignment:      gui.Right,
	}
	logarithmic = gui.Typography{
		ToPrint:        "Logarithmic",
		LenFromToPrint: true,
		Alignment:      gui.Right,
	}
)

func (c controlState) makeFollowingAndLogarithmicBox() gui.Box {
	return makeControlBox(following, logarithmic)
}
func (c controlState) makeFollowingBox() gui.Box {
	return makeControlBox(following)
}
func (c controlState) makeLogarithmicBox() gui.Box {
	return makeControlBox(logarithmic)
}

func makeControlBox(ts ...gui.Typography) gui.Box {
	return gui.Box{
		BoxText: ts,
		Position: gui.Position{
			Vertical:   gui.Top,
			Horizontal: gui.Right,
			Padding:    gui.Padding{Left: 1},
		},
		Style: gui.NoBorder,
	}
}
