// Use of this source code is governed by a GPL-2 license that can be found in the LICENSE file.
//
// Copyright 2025 Lexer747
//
// SPDX-License-Identifier: GPL-2.0-only

package gui

import (
	"strings"

	"github.com/Lexer747/acci-ping/terminal"
	"github.com/Lexer747/acci-ping/utils/bytes"
)

// Typography implements the [Draw] interface and knows how to draw text to the screen. Note that it only
// knows how to do a single unbroken line of text, for multi line strings you should use a [Box].
type Typography struct {
	// ToPrint is the actual text to print.
	ToPrint string
	// TextLen isn't always equal to len(ToPrint) because of unicode characters and ansi control characters
	// hence why it's a separate field.
	TextLen int
	// LenFromToPrint if true will cause the draw call to always overwrite TextLen with len(ToPrint), safe to
	// use if no ansi (e.g. [ansi.Green]) colours are used.
	LenFromToPrint bool
	Alignment      HorizontalAlignment
}

func (t Typography) Len() int {
	if t.LenFromToPrint {
		return len(t.ToPrint)
	}
	return t.TextLen
}

func (t Typography) init(maxTextLength int) initialisedTypography {
	return initialisedTypography{
		Typography:    t,
		maxTextLength: maxTextLength,
	}
}

type initialisedTypography struct {
	Typography

	maxTextLength int
}

func (t initialisedTypography) Draw(size terminal.Size, b *bytes.SafeBuffer) {
	if t.Len() > t.maxTextLength {
		b.WriteString(t.ToPrint)
		return
	}
	switch t.Alignment {
	case Centre:
		padding := (t.maxTextLength - t.Len()) / 2
		leftPadding, rightPadding := getLeftRightPadding(padding, padding, t.Len(), t.maxTextLength)
		b.WriteString(strings.Repeat(" ", leftPadding) + t.ToPrint + strings.Repeat(" ", rightPadding))
	case Left:
		padding := t.maxTextLength - t.Len()
		b.WriteString(t.ToPrint + strings.Repeat(" ", padding))
	case Right:
		padding := t.maxTextLength - t.Len()
		b.WriteString(strings.Repeat(" ", padding) + t.ToPrint)
	default:
		panic("unknown Alignment: " + t.Alignment.String())
	}
}

func getLeftRightPadding(leftPadding, rightPadding, cur, maxLen int) (int, int) {
	for leftPadding+rightPadding+cur > maxLen {
		if leftPadding+rightPadding+cur%2 == 0 {
			leftPadding--
		} else {
			rightPadding--
		}
	}
	for leftPadding+rightPadding+cur < maxLen {
		if leftPadding+rightPadding+cur%2 == 0 {
			leftPadding++
		} else {
			rightPadding++
		}
	}
	return leftPadding, rightPadding
}
