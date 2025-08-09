// Use of this source code is governed by a GPL-2 license that can be found in the LICENSE file.
//
// Copyright 2024-2025 Lexer747
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
	"github.com/Lexer747/acci-ping/terminal"
	"github.com/Lexer747/acci-ping/utils/sliceutils"
)

// toastNotifications which should only be called once the paint buffer is initialised.
func (app *Application) toastNotifications(ctx context.Context, terminalSizeUpdates <-chan terminal.Size) {
	store := gui.NewNotification(app.term.GetSize(), makeToastBox)
	toastBuffer := app.drawBuffer.Get(draw.ToastIndex)
	for {
		select {
		case <-ctx.Done():
			return
		case newSize := <-terminalSizeUpdates:
			store.NewSize(app.GUIState, newSize, toastBuffer)
		case toShow := <-app.errorChannel:
			if toShow == nil {
				continue
			}
			slog.Info("New Error being shown", "err", toShow)
			store.NewValue(app.GUIState, toShow.Error(), toastBuffer, 20*time.Second)
		}
	}
}

const title = " An Error Occurred "

func makeToastBox(size terminal.Size, errors []string) gui.Draw {
	text := make([]gui.Typography, 0, len(errors)+1)
	text = append(text, gui.Typography{ToPrint: themes.Negative(title), TextLen: 19, Alignment: gui.Centre})
	// TODO wrap differently when this might be ontop/underneath the help box.
	maxSize := (size.Width * 3) / 4
	for _, e := range errors {
		for line := range strings.SplitSeq(e, "\n") {
			if len([]rune(line)) >= maxSize {
				// TODO split on spaces instead ...
				for _, splitLine := range sliceutils.SplitN([]rune(line), maxSize) {
					text = append(text, gui.Typography{
						ToPrint:        string(splitLine),
						LenFromToPrint: true,
						Alignment:      gui.Centre,
					})
				}
			} else {
				text = append(text, gui.Typography{
					ToPrint:        line,
					LenFromToPrint: true,
					Alignment:      gui.Centre,
				})
			}
		}
	}
	return &gui.Box{
		BoxText: text,
		Position: gui.Position{
			Vertical:   gui.Middle,
			Horizontal: gui.Centre,
			Padding:    gui.NoPadding,
		},
		Style:         gui.RoundedCorners,
		Configuration: gui.BoxCfg{},
	}
}
