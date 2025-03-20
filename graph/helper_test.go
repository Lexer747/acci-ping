// Use of this source code is governed by a GPL-2 license that can be found in the LICENSE file.
//
// Copyright 2024-2025 Lexer747
//
// SPDX-License-Identifier: GPL-2.0-only

package graph_test

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/Lexer747/acci-ping/graph/terminal"
	"github.com/Lexer747/acci-ping/graph/terminal/ansi"
	"github.com/Lexer747/acci-ping/utils/check"
)

func makeBuffer(size terminal.Size) []string {
	output := make([]string, size.Height)
	for i := range output {
		output[i] = strings.Repeat(" ", size.Width)
	}
	return output
}

type ansiState struct {
	cursorRow, cursorColumn int
	// buffer is the in memory representation of the terminal, each entry of the slice is a row.
	buffer      []string
	size        terminal.Size
	ansiText    string
	asRunes     []rune
	head        int
	outOfBounds bool

	// terminalWrapping will allow the helper to line wrap the output text when the end of the terminal is
	// reached, instead of panicking.
	terminalWrapping bool

	// for errors in tests we accumulate this debug object which will store the runes processed to create the
	// current terminal.
	debug debug
}

func (a *ansiState) peekN(n int) rune     { return a.asRunes[a.head+n] }
func (a *ansiState) peek() rune           { return a.peekN(1) }
func (a *ansiState) isNext(r rune) bool   { return a.peek() == r }
func (a *ansiState) consume()             { a.head++ }
func (a *ansiState) isDigit() bool        { return a.peek() >= '0' && a.peek() <= '9' }
func (a *ansiState) isNegativeSign() bool { return a.peek() == '-' }
func (a *ansiState) consumeIfNext(r rune) bool {
	if ok := a.isNext(r); ok {
		a.consume()
		a.debug.consume(a, r)
		return true
	}
	return false
}
func (a *ansiState) consumeExact(s string) {
	start := a.head - 1
	for _, r := range s {
		check.Checkf(a.consumeIfNext(r), "consumeExact Expected %q got %q. Debug: %s", s, string(a.asRunes[start:a.head]), a.debug)
	}
}
func (a *ansiState) consumeOneOf(s string) bool {
	for _, r := range s {
		if a.consumeIfNext(r) {
			return true
		}
	}
	return false
}
func (a *ansiState) consumeDigits() int {
	digits := []rune{}
	if a.isNegativeSign() {
		digits = append(digits, '-')
		a.consume()
		a.debug.consume(a, '-')
	}
	for a.isDigit() {
		digit := a.peek()
		digits = append(digits, digit)
		a.consume()
		a.debug.consume(a, digit)
	}
	parsed, _ := strconv.ParseInt(string(digits), 10, 0)
	return int(parsed)
}

func playAnsiOntoStringBuffer(ansiText string, buffer []string, size terminal.Size, terminalWrapping bool) []string {
	a := &ansiState{
		cursorRow:        1,
		cursorColumn:     1,
		buffer:           buffer,
		size:             size,
		ansiText:         ansiText,
		asRunes:          []rune(ansiText),
		head:             0,
		debug:            debug{data: []debugItem{}},
		terminalWrapping: terminalWrapping,
	}

	for {
		c := a.peekN(0)
		a.debug.reset(a, c)
		switch c {
		case '\033':
			start := a.head
			if a.consumeIfNext('[') {
				a.handleControl(start)
			}
		default:
			a.write(c)
			a.changeCursor(a.cursorColumn+1, a.cursorRow)
		}
		a.consume()
		if a.EoF() {
			break
		}
	}
	return a.buffer
}

func (a *ansiState) EoF() bool {
	return a.head >= len(a.asRunes)
}

func (a *ansiState) write(c rune) {
	if a.outOfBounds {
		panic(fmt.Sprintf(
			"row out of bounds writing (%c): row is %d, terminal height is %d. Debug: %s",
			c, a.cursorRow, a.size.Height, a.debug,
		))
	}
	y := []rune(a.buffer[a.cursorRow-1])
	y[a.cursorColumn-1] = c
	a.buffer[a.cursorRow-1] = string(y)
}

func (a *ansiState) handleControl(start int) {
	switch {
	case a.isNext('?'):
		a.debug.setCommand(a, "show|hide cursor")
		// show hide cursor control bytes
		a.consumeExact("25")
		if !a.consumeOneOf("lh") {
			panic(fmt.Sprintf("unexpected control byte sequence %q", string(a.asRunes[start:a.head])))
		}
	case a.isNext('H'): // CursorPosition
		a.debug.setCommand(a, "CursorPosition 1, 1")
		// Shortest possible hand for 'CSI1;1H'
		a.changeCursor(1, 1)
		a.consume()
	case a.isNext(';'): // CursorPosition
		// The first row param has been omitted (meaning it's one)
		a.consume()
		d := a.consumeDigits()
		a.debug.setCommand(a, "CursorPosition %d, 1", d)
		a.consumeExact("H")
		a.changeCursor(d, 1)
	case a.isDigit() || a.isNegativeSign():
		d := a.consumeDigits()
		peeked := a.peek()
		switch peeked {
		case 'm':
			a.consume()
			a.debug.setCommand(a, "Colour")
		case ';': // CursorPosition
			// Both params present
			a.consume()
			col := a.consumeDigits()
			a.debug.setCommand(a, "CursorPosition %d, %d", col, d)
			a.consumeExact("H")
			a.changeCursor(col, d)
		case 'H': // CursorPosition
			// The second column param has been omitted (meaning it's one)
			a.debug.setCommand(a, "CursorPosition 1, %d", d)
			a.changeCursor(1, d)
			a.consume()
		case 'A': // CursorUp
			a.debug.setCommand(a, "CursorUp")
			a.changeCursor(a.cursorColumn, a.cursorRow-d)
			a.consume()
		case 'B': // CursorDown
			a.debug.setCommand(a, "CursorDown")
			a.changeCursor(a.cursorColumn, a.cursorRow+d)
			a.consume()
		case 'C': // CursorForward
			a.debug.setCommand(a, "CursorForward")
			a.changeCursor(a.cursorColumn+d, a.cursorRow)
			a.consume()
		case 'D': // CursorBack
			a.debug.setCommand(a, "CursorBack")
			a.changeCursor(a.cursorColumn-d, a.cursorRow)
			a.consume()
		case 'E': // CursorNextLine
			panic("todo CursorNextLine")
		case 'F': // CursorPreviousLine
			panic("todo CursorPreviousLine")
		case 'G': // CursorHorizontalAbsolute
			panic("todo CursorHorizontalAbsolute")
		case 'J': // EraseInDisplay
			a.debug.setCommand(a, "EraseInDisplay %d", d)
			switch ansi.ED(d) {
			case ansi.CursorToScreenEnd:
			case ansi.CursorToScreenBegin:
			case ansi.CursorScreen:
				a.buffer = makeBuffer(a.size)
			case ansi.CursorScreenAndScrollBack:
			default:
				panic("unknown EraseInDisplay enum")
			}
			a.consume()
		case 'K': // EraseInLine
			panic("todo EraseInLine")
		}
		a.debug.consume(a, peeked)
	default:
	}
}

func (a *ansiState) changeCursor(newC, newR int) {
	a.cursorColumn = newC
	a.cursorRow = newR
	// Positive wrapping, go to the next line
	if a.cursorColumn > a.size.Width {
		a.cursorColumn = 1
		a.cursorRow++
	}
	// Negative wrapping, go back to the last line, last col
	if a.cursorColumn <= 0 {
		a.cursorColumn = a.size.Width - 1
		a.cursorRow--
	}
	if a.terminalWrapping && a.cursorRow > a.size.Height {
		// since we are allowed to wrap, we can print this newline and move the cursor back up to the start of
		// the now new line.
		a.newline()
		a.cursorRow--
		a.cursorColumn = 1
	}
	if a.cursorRow > a.size.Height || a.cursorRow < 0 {
		a.outOfBounds = true
	} else {
		a.outOfBounds = false
	}
	check.Checkf(a.cursorColumn != 0 && a.cursorRow != 0, "cursor should not be 0. Debug: %s", a.debug)
}

func (a *ansiState) newline() {
	// We create a "newline" by rotating all the string buffers (each entry is a row) up one position
	for i := range len(a.buffer) - 1 {
		a.buffer[i] = a.buffer[i+1]
	}
	// then inserting a new clean row to the last entry.
	a.buffer[len(a.buffer)-1] = strings.Repeat(" ", a.size.Width)
}
