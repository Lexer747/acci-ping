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
	"github.com/Lexer747/acci-ping/gui"
	"github.com/Lexer747/acci-ping/gui/themes"
	"github.com/Lexer747/acci-ping/terminal"
)

// help which should only be called once the paint buffer is initialised.
func (app *Application) help(
	ctx context.Context,
	startShowHelp bool,
	helpChannel <-chan rune,
	terminalSizeUpdates <-chan terminal.Size,
) {
	helpBuffer := app.drawBuffer.Get(draw.HelpIndex)
	h := help{showHelp: startShowHelp}
	app.paint(h.render(app.term.GetSize(), helpBuffer))
	for {
		select {
		case <-ctx.Done():
			return
		case newSize := <-terminalSizeUpdates:
			app.paint(h.render(newSize, helpBuffer))
		case toShow := <-helpChannel:
			switch toShow {
			case 'h':
				h.showHelp = !h.showHelp
				app.paint(h.render(app.term.GetSize(), helpBuffer))
			default:
			}
		}
	}
}

type help struct {
	showHelp bool
}

func (h help) render(size terminal.Size, buf *bytes.Buffer) paintUpdate {
	ret := None
	shouldInvalidate := buf.Len() != 0
	if shouldInvalidate {
		ret = ret | Invalidate
	}
	buf.Reset()
	if h.showHelp {
		box := h.makeHelpBox()
		box.Draw(size, buf)
		return ret | Paint
	}
	return ret
}

func helpAction(ch chan<- rune) func(r rune) error {
	return func(r rune) error {
		ch <- r
		return nil
	}
}

func (h help) makeHelpBox() gui.Box {
	return gui.Box{
		BoxText: helpCopy,
		Position: gui.Position{
			Vertical:   gui.Middle,
			Horizontal: gui.Right,
			Padding:    gui.Padding{Left: 4},
		},
		Style: gui.SharpCorners,
	}
}

func helpStartup() {
	helpText := themes.Highlight("Help")
	ctrlCText := themes.Positive("ctrl+c")
	keyBindF := themes.Positive("f")
	keyBindH := themes.Positive("h")

	helpCopy = []gui.Typography{
		{ToPrint: helpText, TextLen: 4, Alignment: gui.Centre},
		{ToPrint: "", TextLen: 0, Alignment: gui.Centre},
		{ToPrint: themes.Primary("Press ") + ctrlCText + themes.Primary(" to exit."),
			TextLen: 6 + 6 + 9, Alignment: gui.Left},
		{ToPrint: themes.Primary("Press ") + keyBindF + themes.Primary(" to follow the most recent data."),
			TextLen: 6 + 1 + 32, Alignment: gui.Left},
		{ToPrint: themes.Primary("Press ") + keyBindH + themes.Primary(" to open/close this window."),
			TextLen: 6 + 1 + 27, Alignment: gui.Left},
	}
}

var helpCopy = make([]gui.Typography, 0, 10)
