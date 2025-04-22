// Use of this source code is governed by a GPL-2 license that can be found in the LICENSE file.
//
// Copyright 2025 Lexer747
//
// SPDX-License-Identifier: GPL-2.0-only

package themes

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/Lexer747/acci-ping/utils/errors"
)

type Colour int

// Deliberately match foreground numbers of 3-bit and 4-bit https://en.wikipedia.org/wiki/ANSI_escape_code for
// the most widely supported initial colour constants

const (
	Black     Colour = 30
	Gray      Colour = 90
	LightGray Colour = 37
	White     Colour = 97

	DarkRed     Colour = 31
	DarkGreen   Colour = 32
	DarkYellow  Colour = 33
	DarkBlue    Colour = 34
	DarkMagenta Colour = 35
	DarkCyan    Colour = 36

	Red     Colour = 91
	Green   Colour = 92
	Yellow  Colour = 93
	Blue    Colour = 94
	Magenta Colour = 95
	Cyan    Colour = 96
)

type Luminance float64

const (
	Dark  Luminance = 0.0
	Light Luminance = 1.0
)

func (l Luminance) IsDark() bool {
	return l < 0.5
}
func (l Luminance) IsLight() bool {
	return l >= 0.5
}

// ST1003 acronym soup
//
//nolint:staticcheck
func ParseRGB_48bit(red, green, blue int) Luminance {
	// 48bit means 16 bits per channel
	const max16Bit = 65535.0
	r := float64(red) / max16Bit
	g := float64(green) / max16Bit
	b := float64(blue) / max16Bit
	return unsafeCCIR601(r, g, b)
}

// ST1003 acronym soup
//
//nolint:staticcheck
func ParseRGB_CSSString(s string) (Luminance, error) {
	r, g, b, err := cssStrToRGB(s)
	if err != nil {
		return -1, err
	}
	const max8Bit = 255.0
	return unsafeCCIR601(float64(r)/max8Bit, float64(g)/max8Bit, float64(b)/max8Bit), nil
}

func cssStrToRGB(s string) (uint8, uint8, uint8, error) {
	trimmed := strings.TrimPrefix(s, "#")
	if len(trimmed) != 6 {
		return 0, 0, 0, errors.Errorf("Wrong number of digits for colour %q, should be 6 hex digits", s)
	}
	redStr := trimmed[0:2]
	greenStr := trimmed[2:4]
	blueStr := trimmed[4:6]
	r, rerr := strconv.ParseInt(redStr, 16, 16)
	g, gerr := strconv.ParseInt(greenStr, 16, 16)
	b, berr := strconv.ParseInt(blueStr, 16, 16)
	if r > 255 || r < 0 {
		rerr = errors.Join(rerr, errors.Errorf("red component out of range %d, should be within 0 and 255", r))
	}
	if g > 255 || g < 0 {
		gerr = errors.Join(gerr, errors.Errorf("green component out of range %d, should be within 0 and 255", g))
	}
	if b > 255 || b < 0 {
		berr = errors.Join(berr, errors.Errorf("green component out of range %d, should be within 0 and 255", b))
	}
	if err := errors.Join(rerr, gerr, berr); err != nil {
		return 0, 0, 0, errors.Wrapf(err, "Couldn't parse RGB values for %q", s)
	}
	// G115: These are not an integer overflow because we bounds check above ^

	return uint8(r), uint8(g), uint8(b), nil //nolint:gosec
}

// unsafeCCIR601 computes the luminance of a colour based on the Red, Green, and Blue values, where the r, g,
// b must be a range from 0 to 1.0. It's unsafe because it doesn't actually check the inputs or outputs for
// sanity.
//
// CCIR 601: https://en.wikipedia.org/wiki/Rec._601
func unsafeCCIR601(r float64, g float64, b float64) Luminance {
	luminance := (0.299 * r) + (0.587 * g) + (0.114 * b)
	return Luminance(luminance)
}

func ParseColour(c string) (Colour, error) {
	switch strings.ToLower(c) {
	case "black":
		return Black, nil
	case "gray":
		return Gray, nil
	case "white":
		return White, nil
	case "lightgray", "light-gray":
		return LightGray, nil
	case "darkred", "light-red":
		return DarkRed, nil
	case "darkgreen", "light-green":
		return DarkGreen, nil
	case "darkyellow", "light-yellow":
		return DarkYellow, nil
	case "darkblue", "light-blue":
		return DarkBlue, nil
	case "darkmagenta", "light-magenta":
		return DarkMagenta, nil
	case "darkcyan", "light-cyan":
		return DarkCyan, nil
	case "red":
		return Red, nil
	case "green":
		return Green, nil
	case "yellow":
		return Yellow, nil
	case "blue":
		return Blue, nil
	case "magenta":
		return Magenta, nil
	case "cyan":
		return Cyan, nil
	}
	return 0, errors.Errorf("Couldn't parse colour from %q", c)
}

type colourImpl struct {
	Prefix string
	Suffix string
	Origin colourJSON
}

func (c colourImpl) Do(s string) string {
	return c.Prefix + s + c.Suffix
}

func (c Colour) String() string {
	switch c {
	case Black:
		return "Black"
	case Gray:
		return "Gray"
	case White:
		return "White"
	case LightGray:
		return "LightGray"
	case DarkRed:
		return "DarkRed"
	case DarkGreen:
		return "DarkGreen"
	case DarkYellow:
		return "DarkYellow"
	case DarkBlue:
		return "DarkBlue"
	case DarkMagenta:
		return "DarkMagenta"
	case DarkCyan:
		return "DarkCyan"
	case Red:
		return "Red"
	case Green:
		return "Green"
	case Yellow:
		return "Yellow"
	case Blue:
		return "Blue"
	case Magenta:
		return "Magenta"
	case Cyan:
		return "Cyan"
	}
	return "unsupported colour"
}

func (c colourImpl) String() string {
	switch {
	case c.Origin.bit4 != nil:
		return c.Origin.bit4.String()
	case c.Origin.bit8 != nil:
		return strconv.Itoa(*c.Origin.bit8)
	case c.Origin.bit24 != nil:
		a := *c.Origin.bit24
		if a.string != nil {
			return *a.string
		} else {
			return fmt.Sprintf("#%00X%00X%00X", *a.R, *a.G, *a.B)
		}
	case c.Origin.noColour:
		return "No Colour"
	}
	return "unknown colour implementation"
}
