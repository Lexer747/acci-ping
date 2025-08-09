// Use of this source code is governed by a GPL-2 license that can be found in the LICENSE file.
//
// Copyright 2025 Lexer747
//
// SPDX-License-Identifier: GPL-2.0-only

package acciping

import (
	"context"
	"log/slog"
	"strings"
	"time"

	"github.com/Lexer747/acci-ping/draw"
	"github.com/Lexer747/acci-ping/gui"
	"github.com/Lexer747/acci-ping/gui/themes"
	"github.com/Lexer747/acci-ping/ping"
	"github.com/Lexer747/acci-ping/terminal"
	"github.com/Lexer747/acci-ping/terminal/typography"
)

func (app *Application) showSpeedChanges(
	ctx context.Context,
	guiSpeedChange chan ping.Speed,
	terminalSizeUpdates <-chan terminal.Size,
) {
	store := gui.NewNotification(app.term.GetSize(), makeSpeedChangeBox)
	buffer := app.drawBuffer.Get(draw.EmojiIndex)
	for {
		select {
		case <-ctx.Done():
			return
		case newSize := <-terminalSizeUpdates:
			store.NewSize(app.GUIState, newSize, buffer)
		case speed := <-guiSpeedChange:
			slog.Info("changing ping speed:", "PingRate", speed)
			var toShow string
			switch speed {
			case ping.Faster:
				toShow = themes.Positive(typography.UpArrow)
			case ping.Slower:
				toShow = themes.Negative(typography.DownArrow)
			case ping.Fastest:
				toShow = themes.Emphasis(typography.DoubleUpArrow)
			default:
				continue
			}

			store.NewValue(app.GUIState, toShow, buffer, time.Second)
		}
	}
}

func makeSpeedChangeBox(_ terminal.Size, es []string) gui.Draw {
	text := gui.Typography{
		ToPrint:        strings.Join(es, ""),
		TextLen:        len(es),
		LenFromToPrint: false,
		Alignment:      gui.Centre,
	}

	return &gui.Box{
		BoxText: []gui.Typography{text},
		Position: gui.Position{
			Vertical:   gui.Top,
			Horizontal: gui.Centre,
			Padding: gui.Padding{
				Bottom: 1,
			},
		},
		Style:         gui.NoBorder,
		Configuration: gui.BoxCfg{},
	}
}
