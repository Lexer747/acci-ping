// Use of this source code is governed by a GPL-2 license that can be found in the LICENSE file.
//
// Copyright 2025 Lexer747
//
// SPDX-License-Identifier: GPL-2.0-only

package acciping

import (
	"bytes"
	"context"

	"github.com/Lexer747/acci-ping/draw"
	"github.com/Lexer747/acci-ping/graph"
	"github.com/Lexer747/acci-ping/gui"
	"github.com/Lexer747/acci-ping/terminal"
)

// showControls which should only be called once the paint buffer is initialised.
func (app *Application) showControls(
	ctx context.Context,
	controls <-chan graph.Control,
	terminalSizeUpdates <-chan terminal.Size,
) {
	buffer := app.drawBuffer.Get(draw.ControlIndex)
	c := controlState{following: false} // TODO when added to config populate it here too
	app.paint(c.render(app.term.GetSize(), buffer))
	for {
		select {
		case <-ctx.Done():
			return
		case newSize := <-terminalSizeUpdates:
			app.paint(c.render(newSize, buffer))
		case update := <-controls:
			if update.FollowLatestSpan {
				c.following = !c.following
			}
			app.paint(c.render(app.term.GetSize(), buffer))
		}
	}
}

type controlState struct {
	following bool
}

func (c controlState) render(size terminal.Size, buf *bytes.Buffer) paintUpdate {
	ret := None
	buf.Reset()
	if !c.following {
		ret = ret | Invalidate
	}
	if c.following {
		box := c.makeControlBox()
		box.Draw(size, buf)
		return ret | Paint
	}
	return ret
}

func (c controlState) makeControlBox() gui.Box {
	return gui.Box{
		BoxText: []gui.Typography{{
			ToPrint:        "Following",
			LenFromToPrint: true,
			Alignment:      gui.Right,
		}},
		Position: gui.Position{
			Vertical:   gui.Top,
			Horizontal: gui.Right,
			Padding:    gui.Padding{Bottom: 1},
		},
		Style: gui.NoBorder,
	}
}
