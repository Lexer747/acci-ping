// Use of this source code is governed by a GPL-2 license that can be found in the LICENSE file.
//
// Copyright 2024-2025 Lexer747
//
// SPDX-License-Identifier: GPL-2.0-only

package ansi

import (
	"strconv"
)

// Helper section

var Clear = EraseInDisplay(CursorScreen)
var Home = CursorPosition(1, 1)

// Spec definitions

func CursorUp(n int) string {
	if n <= 0 {
		return ""
	}
	return CSI + i(n) + "A"
}
func CursorDown(n int) string {
	if n <= 0 {
		return ""
	}
	return CSI + i(n) + "B"
}

// CursorForward will emit a no-op for values of [n] less than 0
func CursorForward(n int) string {
	if n <= 0 {
		return ""
	}
	return CSI + i(n) + "C"
}

// CursorBack will emit a no-op for values of [n] less than 0
func CursorBack(n int) string {
	if n <= 0 {
		return ""
	}
	return CSI + i(n) + "D"
}
func CursorNextLine(n int) string           { return CSI + i(n) + "E" }
func CursorPreviousLine(n int) string       { return CSI + i(n) + "F" }
func CursorHorizontalAbsolute(n int) string { return CSI + i(n) + "G" }

type ED int // Erase in Display
type EL int // Erase in Line

const (
	// Control Sequence Introducer | Starts most of the useful sequences, terminated by a byte in the range
	// 0x40 through 0x7E.
	CSI = "\033["

	// Operating System Commands | Does other things not as well documented.
	OSC = "\x1b]"

	CursorToScreenEnd         ED = 0
	CursorToScreenBegin       ED = 1
	CursorScreen              ED = 2
	CursorScreenAndScrollBack ED = 3

	CursorToEndOfLine   EL = 0
	CursorToBeginOfLine EL = 1
	EntireLine          EL = 2

	// FormattingReset turns all attributes off, including colour, bold, etc.
	FormattingReset = CSI + "0m"
	HideCursor      = CSI + "?25l"
	ShowCursor      = CSI + "?25h"
)

// Compacted when defaults are passed, some chars may be elided:
//
// > The values are 1-based, and default to '1' (top left corner) if omitted. A sequence such as 'CSI ;5H' is
// > a synonym for 'CSI 1;5H' as well as 'CSI 17;H' is the same as 'CSI 17H' and 'CSI 17;1H'. [wikipedia]
//
// In the graph context [row] is the `Y` coordinate and the [column] is the `X` coordinate.
//
// [wikipedia]: https://en.wikipedia.org/wiki/ANSI_escape_code
func CursorPosition(row, column int) string {
	if row == 1 && column == 1 {
		return CSI + "H"
	} else if row == 1 {
		return CSI + ";" + i(column) + "H"
	} else if column == 1 {
		return CSI + i(row) + "H"
	}
	if row <= 0 && column <= 0 {
		return ""
	}
	return CSI + i(row) + ";" + i(column) + "H"
}

func EraseInDisplay(n ED) string { return CSI + i(int(n)) + "J" }
func EraseInLine(n EL) string    { return CSI + i(int(n)) + "K" }

// helpful short hands inside the package

var i = strconv.Itoa
var r = FormattingReset

// Colours Section:

func Black(s string) string       { return CSI + "30m" + s + r }
func Blue(s string) string        { return CSI + "94m" + s + r }
func Cyan(s string) string        { return CSI + "96m" + s + r }
func DarkBlue(s string) string    { return CSI + "34m" + s + r }
func DarkCyan(s string) string    { return CSI + "36m" + s + r }
func DarkGreen(s string) string   { return CSI + "32m" + s + r }
func DarkMagenta(s string) string { return CSI + "35m" + s + r }
func DarkRed(s string) string     { return CSI + "31m" + s + r }
func DarkYellow(s string) string  { return CSI + "33m" + s + r }
func Gray(s string) string        { return CSI + "90m" + s + r }
func Green(s string) string       { return CSI + "92m" + s + r }
func LightGray(s string) string   { return CSI + "37m" + s + r }
func Magenta(s string) string     { return CSI + "95m" + s + r }
func Red(s string) string         { return CSI + "91m" + s + r }
func White(s string) string       { return CSI + "97m" + s + r }
func Yellow(s string) string      { return CSI + "93m" + s + r }

// 256 Colours (8-bit)

// Colour is the 8-bit colour choice Ansi encoding, note this is not a specific palette and generally
// implementation defined. Highly recommended to pick colours by inspecting them locally.
//
// https://en.wikipedia.org/wiki/ANSI_escape_code#8-bit
func Colour(s string, colour int) string { return CSI + "38;5;" + i(colour) + "m" + s + r }

// 16,777,216 Colours (24-bit)

// TrueColour is the 24-bit colour choice this is the same as CSS colouring, note that not all terminals will
// support this.
//
// https://en.wikipedia.org/wiki/Color_depth#True_color_(24-bit)
func TrueColour(s string, red, green, blue uint8) string {
	return CSI + "38;2;" + i(int(red)) + ";" + i(int(green)) + ";" + i(int(blue)) + "m" + s + r
}

// Fonts:

func Bold(s string) string          { return CSI + "1m" + s + r }
func Light(s string) string         { return CSI + "2m" + s + r }
func Italic(s string) string        { return CSI + "3m" + s + r }
func Underline(s string) string     { return CSI + "4m" + s + r }
func Strikethrough(s string) string { return CSI + "9m" + s + r }
