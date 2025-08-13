// Use of this source code is governed by a GPL-2 license that can be found in the LICENSE file.
//
// Copyright 2024-2025 Lexer747
//
// SPDX-License-Identifier: GPL-2.0-only

package graph

import (
	"time"

	"github.com/Lexer747/acci-ping/gui/themes"
	"github.com/Lexer747/acci-ping/terminal"
	"github.com/Lexer747/acci-ping/terminal/ansi"
	"github.com/Lexer747/acci-ping/terminal/typography"
)

var spinnerArray = [...]string{
	typography.UpperLeftQuadrantCircularArc,
	typography.UpperRightQuadrantCircularArc,
	typography.LowerRightQuadrantCircularArc,
	typography.LowerLeftQuadrantCircularArc,
}

type spinner struct {
	spinnerIndex       int
	timestampLastDrawn time.Time
}

func (s *spinner) spinner(size terminal.Size) string {
	now := time.Now()
	if s.timestampLastDrawn.Add(225 * time.Millisecond).Before(now) {
		s.spinnerIndex++
		s.timestampLastDrawn = now
	}
	return ansi.CursorPosition(1, size.Width-3) + themes.Emphasis(spinnerArray[s.spinnerIndex%len(spinnerArray)])
}
