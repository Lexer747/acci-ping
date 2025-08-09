// Use of this source code is governed by a GPL-2 license that can be found in the LICENSE file.
//
// Copyright 2024-2025 Lexer747
//
// SPDX-License-Identifier: GPL-2.0-only

// th stands for "test helper"
package th

import (
	"fmt"
	"reflect"
	"strconv"
	"strings"
	"time"

	"github.com/Lexer747/acci-ping/terminal"
	"github.com/Lexer747/acci-ping/terminal/ansi"
	"github.com/Lexer747/acci-ping/utils/check"
	"github.com/Lexer747/acci-ping/utils/numeric"
	"github.com/google/go-cmp/cmp"
	"gotest.tools/v3/assert"
	is "gotest.tools/v3/assert/cmp"
	"pgregory.net/rapid"
)

// TestWithTimeout allows a test function to **always** run with a timeout similar to the `go test` built in
// `-timeout` flag.
func TestWithTimeout(t T, timeout time.Duration, test func()) {
	c := time.After(timeout)
	done := make(chan struct{})
	go func() {
		test()
		done <- struct{}{}
	}()
	select {
	case <-c:
		t.Fatalf("Test timed out after %s", timeout.String())
	case <-done:
	}
}

// AssertFloatEqual checks that the two floats are equal within the given significant figures.
func AssertFloatEqual(t T, expected float64, actual float64, sigFigs int, msgAndArgs ...any) {
	t.Helper()
	a := numeric.RoundToNearestSigFig(actual, sigFigs)
	e := numeric.RoundToNearestSigFig(expected, sigFigs)
	assert.Check(t, is.Equal(e, a), msgAndArgs...)
}

var AllowAllUnexported = cmp.Exporter(func(reflect.Type) bool { return true })

// T is the "current" acci-ping most generic test interface for the most use with all test frameworks and
// third party helpers. [*testing.T] is safer and easier to use if in doubt, you will know when you need this
// helper because certain cross dependency tests have stopped compiling.
type T interface {
	rapid.TB
}

func MakeBuffer(size terminal.Size) []string {
	output := make([]string, size.Height)
	for i := range output {
		output[i] = strings.Repeat(" ", size.Width)
	}
	return output
}

type TerminalWrapping int

const (
	// Strict, suitable for graph tests
	Panic TerminalWrapping = iota
	// Rotates the input buffer, essentially representing a scroll down of the terminal output
	WrapBuffer
	// Wraps where possible and if out of bounds occurs (i.e. scrolls up too much) will silently drop the
	// characters
	SilentlyDrop
)

// EmulateTerminal is essentially a terminal emulator shim. Unlike a raw string buffer this will apply ansi
// commands to move the cursor around in the output space. This allows any ansi text to be presented into a
// string buffer which can be inspected. This allows for in memory terminal.
//
// Since it's the target of tests it accidentally describes the minimal ansi commands that acci-ping needs to
// function in a sane way. Note that acci-ping uses more commands which this emulator doesn't support (such as
// colour).
//
// The api of this function is paniky by nature, this is because it's not designed to give good errors or be
// robust as it should only be used for tests.
func EmulateTerminal(ansiText string, buffer []string, size terminal.Size, terminalWrapping TerminalWrapping) []string {
	// this all the state we need to track to emulate a terminal, its fixed size and unchanging
	a := &ansiState{
		cursorRow:    1,
		cursorColumn: 1,
		buffer:       buffer,
		size:         size,
		ansiText:     ansiText,
		asRunes:      []rune(ansiText),
		head:         0,

		// debug is a bit nebulous, it's only kept in memory and never referenced outside this package, the
		// idea is that in a literal debugger you can inspect this memory. The debug will be filled with the
		// past ansi commands and runes which have already been consumed.
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

type ansiState struct {
	cursorRow, cursorColumn int
	// buffer is the in memory representation of the terminal, each entry of the slice is a row.
	buffer      []string
	size        terminal.Size
	ansiText    string
	asRunes     []rune
	head        int
	outOfBounds bool

	terminalWrapping TerminalWrapping

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

func (a *ansiState) EoF() bool {
	return a.head >= len(a.asRunes)
}

func (a *ansiState) write(c rune) {
	if a.outOfBounds {
		if a.terminalWrapping == Panic || a.terminalWrapping == WrapBuffer {
			panic(fmt.Sprintf(
				"row out of bounds writing (%c): row is %d, terminal height is %d. Debug: %s",
				c, a.cursorRow, a.size.Height, a.debug,
			))
		} else {
			// silently drop the rune
			return
		}
	}
	y := []rune(a.buffer[a.cursorRow-1])
	y[a.cursorColumn-1] = c
	a.buffer[a.cursorRow-1] = string(y)
}

// handleControl function describes the supported ansi commands of the emulator.
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
				a.buffer = MakeBuffer(a.size)
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
	if a.terminalWrapping == WrapBuffer && a.cursorRow > a.size.Height {
		// since we are allowed to wrap, we can print this newline and move the cursor back up to the start of
		// the now new line.
		a.newline()
		a.cursorRow--
		a.cursorColumn = 1
	}

	// The state is out of bounds if, the cursors row is above the maximum height of this emulator. OR the row
	// is negative.
	a.outOfBounds = a.cursorRow > a.size.Height || a.cursorRow < 0

	// Finally handle out of bounds cases:
	if a.terminalWrapping != SilentlyDrop {
		check.Checkf(a.cursorColumn != 0 && a.cursorRow != 0, "cursor should not be 0. Debug: %s", a.debug)
	} else {
		if a.cursorColumn == 0 {
			a.cursorColumn++
		}
		if a.cursorRow == 0 {
			a.cursorRow++
		}
	}
}

func (a *ansiState) newline() {
	// We create a "newline" by rotating all the string buffers (each entry is a row) up one position
	for i := range len(a.buffer) - 1 {
		a.buffer[i] = a.buffer[i+1]
	}
	// then inserting a new clean row to the last entry.
	a.buffer[len(a.buffer)-1] = strings.Repeat(" ", a.size.Width)
}
