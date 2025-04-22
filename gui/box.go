// Use of this source code is governed by a GPL-2 license that can be found in the LICENSE file.
//
// Copyright 2025 Lexer747
//
// SPDX-License-Identifier: GPL-2.0-only

package gui

import (
	"bytes"
	"fmt"
	"strconv"
	"strings"

	"github.com/Lexer747/acci-ping/gui/themes"
	"github.com/Lexer747/acci-ping/terminal"
	"github.com/Lexer747/acci-ping/terminal/ansi"
)

type Box struct {
	// BoxText is the slice of text to show where each element represents a separate line
	BoxText       []Typography
	Position      Position
	Style         Style
	Configuration BoxCfg
}

type Style int

const (
	RoundedCorners Style = 1
	SharpCorners   Style = 2
	NoBorder       Style = 3
)

func (s Style) String() string {
	switch s {
	case RoundedCorners:
		return "RoundedCorners"
	case SharpCorners:
		return "SharpCorners"
	case NoBorder:
		return "NoBorder"
	default:
		return "Unknown Style: " + strconv.Itoa(int(s))
	}
}

type BoxCfg struct {
	DefaultWidth int
}

func (b Box) Draw(size terminal.Size, buf *bytes.Buffer) {
	p := b.position(size)
	bar := strings.Repeat(b.Style.getHorizontal(), b.boxTextWidth(size))
	corners := b.Style.getCorner()
	buf.WriteString(ansi.CursorPosition(p.startY, p.startX) + corners.TopLeft + bar + corners.TopRight)
	end := 0
	for i, t := range b.BoxText {
		end = i
		if i >= size.Height {
			break
		}
		buf.WriteString(ansi.CursorPosition(p.startY+i+1, p.startX) + b.Style.getVertical())
		t.init(b.boxTextWidth(size)).Draw(size, buf)
		buf.WriteString(b.Style.getVertical())
	}
	buf.WriteString(ansi.CursorPosition(p.startY+end+2, p.startX) + corners.BottomLeft + bar + corners.BottomRight)
}

type boxPosition struct {
	startY, startX int
}

func (b Box) position(size terminal.Size) boxPosition {
	p := b.Position
	ret := boxPosition{}
	switch {
	case p.Horizontal == Centre && p.Vertical == Middle:
		originX := size.Width / 2
		originY := size.Height / 2
		ret = boxPosition{
			startY: originY - b.height(size)/2,
			startX: originX - b.width(size)/2,
		}
	case p.Vertical == Middle && p.Horizontal == Right:
		originY := size.Height / 2
		ret = boxPosition{
			startY: originY - b.height(size)/2,
			startX: size.Width - b.width(size),
		}
	case p.Vertical == Top && p.Horizontal == Right:
		ret = boxPosition{
			startY: b.height(size),
			startX: size.Width - b.width(size),
		}
	default:
		panic(fmt.Sprintf("unhandled:box:position %+v", p))
	}
	if !p.Padding.Equal(NoPadding) {
		ret.startY = ret.startY - p.Padding.Top + p.Padding.Bottom
		ret.startX = ret.startX - p.Padding.Left + p.Padding.Right
	}
	return ret
}

func (b Box) height(size terminal.Size) int {
	return min(size.Height-1, len(b.BoxText))
}

func (b Box) width(size terminal.Size) int {
	return b.boxTextWidth(size) + b.widthFromStyle()
}

func (b Box) boxTextWidth(size terminal.Size) int {
	if b.height(size) == 0 {
		return b.Configuration.DefaultWidth
	}
	ret := 0
	for _, t := range b.BoxText {
		ret = max(ret, t.Len())
	}
	return ret
}

func (b Box) widthFromStyle() int {
	switch b.Style {
	case NoBorder:
		return 0
	case RoundedCorners, SharpCorners:
		return 2
	default:
		panic("unknown box style: " + b.Style.String())
	}
}

type corners struct {
	TopLeft, TopRight, BottomLeft, BottomRight string
}

func (s Style) getVertical() string {
	switch s {
	case RoundedCorners, SharpCorners:
		return themes.Primary("│")
	case NoBorder:
		return ""
	default:
		panic("unknown box style: " + s.String())
	}
}

func (s Style) getHorizontal() string {
	switch s {
	case RoundedCorners, SharpCorners:
		return themes.Primary("─")
	case NoBorder:
		return ""
	default:
		panic("unknown box style: " + s.String())
	}
}

func (s Style) getCorner() corners {
	switch s {
	case RoundedCorners:
		return corners{
			TopLeft:     themes.Primary("╭"),
			TopRight:    themes.Primary("╮"),
			BottomLeft:  themes.Primary("╰"),
			BottomRight: themes.Primary("╯"),
		}
	case SharpCorners:
		return corners{
			TopLeft:     themes.Primary("┌"),
			TopRight:    themes.Primary("┐"),
			BottomLeft:  themes.Primary("└"),
			BottomRight: themes.Primary("┘"),
		}
	case NoBorder:
		return corners{}
	default:
		panic("unknown box style: " + s.String())
	}
}
