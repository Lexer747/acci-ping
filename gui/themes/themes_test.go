// Use of this source code is governed by a GPL-2 license that can be found in the LICENSE file.
//
// Copyright 2025 Lexer747
//
// SPDX-License-Identifier: GPL-2.0-only

package themes_test

import (
	"encoding/json"
	"testing"

	"github.com/Lexer747/acci-ping/gui/themes"
	"github.com/Lexer747/acci-ping/terminal/ansi"
	"github.com/Lexer747/acci-ping/utils/errors"
	"gotest.tools/v3/assert"
)

//nolint:lll
func TestThemeString(t *testing.T) {
	t.Parallel()
	assert.Equal(t, themes.DarkTheme.String(),
		`{primary: White secondary: Gray highlight: Yellow emphasis: Cyan titleHighlight: Magenta positive: Green darkPositive: DarkGreen negative: Red darkNegative: DarkRed}`)
	assert.Equal(t, themes.LightTheme.String(),
		`{primary: Black secondary: Gray highlight: DarkYellow emphasis: DarkCyan titleHighlight: DarkMagenta positive: Green darkPositive: DarkGreen negative: Red darkNegative: DarkRed}`)
	assert.Equal(t, themes.NoTheme.String(),
		`{primary: No Colour secondary: No Colour highlight: No Colour emphasis: No Colour titleHighlight: No Colour positive: No Colour darkPositive: No Colour negative: No Colour darkNegative: No Colour}`)
	assert.Equal(t, themes.ComplexTheme.String(),
		`{primary: 15 secondary: 243 highlight: 227 emphasis: 87 titleHighlight: 199 positive: #30FF02 darkPositive: #1E75B negative: Red darkNegative: DarkRed}`)
}

//nolint:lll
func TestAnsiThemeCompatibility(t *testing.T) {
	t.Parallel()
	tests := []AnsiThemeTest{
		{Input: `"Black"`, Output: ansi.Black},
		{Input: `"Gray"`, Output: ansi.Gray},
		{Input: `"LightGray"`, Output: ansi.LightGray},
		{Input: `"White"`, Output: ansi.White},
		{Input: `"DarkRed"`, Output: ansi.DarkRed},
		{Input: `"DarkGreen"`, Output: ansi.DarkGreen},
		{Input: `"DarkYellow"`, Output: ansi.DarkYellow},
		{Input: `"DarkBlue"`, Output: ansi.DarkBlue},
		{Input: `"DarkMagenta"`, Output: ansi.DarkMagenta},
		{Input: `"DarkCyan"`, Output: ansi.DarkCyan},
		{Input: `"Red"`, Output: ansi.Red},
		{Input: `"Green"`, Output: ansi.Green},
		{Input: `"Yellow"`, Output: ansi.Yellow},
		{Input: `"Blue"`, Output: ansi.Blue},
		{Input: `"Magenta"`, Output: ansi.Magenta},
		{Input: `"Cyan"`, Output: ansi.Cyan},

		{Input: `{"8-bit": 15}`, Output: CurryColour(15)},
		{Input: `{"8-bit": 255}`, Output: CurryColour(255)},
		{Input: `{"8-bit": 0}`, Output: CurryColour(0)},

		{Input: `{"24-bit": "112233"}`, Output: CurryTrueColour(0x11, 0x22, 0x33)},
		{Input: `{"24-bit": "FF00AA"}`, Output: CurryTrueColour(0xFF, 0x00, 0xAA)},
		{Input: `{"24-bit": "#FF00AA"}`, Output: CurryTrueColour(0xFF, 0x00, 0xAA)},

		{Input: `"CyanWithATypo"`, UnmarshalErr: errors.Errorf(`while parsing colourJson caused by: Couldn't parse colour from "CyanWithATypo"`)},

		{Input: `{"8-bit": -1}`, ImplErr: errors.Errorf(`8-bit colour component out of range -1, should be within 0 and 255`)},
		{Input: `{"8-bit": ff}`, UnmarshalErr: errors.Errorf(`invalid character 'f' in literal false (expecting 'a')`)},

		{Input: `{"24-bit": "##FF00AA"}`, UnmarshalErr: errors.Errorf(`while parsing colourJson caused by: while parsing 24bit colour caused by: Wrong number of digits for colour "##FF00AA", should be 6 hex digits`)},
		{Input: `{"24-bit": "-1"}`, UnmarshalErr: errors.Errorf(`while parsing colourJson caused by: while parsing 24bit colour caused by: Wrong number of digits for colour "-1", should be 6 hex digits`)},
		{Input: `{"29-bit": "-1"}`, UnmarshalErr: errors.Errorf(`Unknown how to parse colour from '{"29-bit": "-1"}'`)},
		{Input: `{"32-bit": "#FF00AA"}`, UnmarshalErr: errors.Errorf(`Unknown how to parse colour from '{"32-bit": "#FF00AA"}'`)},
		{Input: `{"24-bit": "ff"}`, UnmarshalErr: errors.Errorf(`while parsing colourJson caused by: while parsing 24bit colour caused by: Wrong number of digits for colour "ff", should be 6 hex digits`)},
	}
	for _, tc := range tests {
		t.Run(tc.Input, tc.Run)
	}
}

func CurryColour(i int) func(s string) string {
	return func(s string) string {
		return ansi.Colour(s, i)
	}
}
func CurryTrueColour(r, g, b uint8) func(s string) string {
	return func(s string) string {
		return ansi.TrueColour(s, r, g, b)
	}
}

type AnsiThemeTest struct {
	UnmarshalErr error
	ImplErr      error
	Output       func(s string) string
	Input        string
}

const testString = "abcdefghijklmnopqrstuvwxyz"

func (tc AnsiThemeTest) Run(t *testing.T) {
	t.Parallel()
	c := themes.ColourJSON{}
	err := json.Unmarshal([]byte(tc.Input), &c)
	if tc.UnmarshalErr == nil {
		assert.NilError(t, err, "Unmarshal")
	} else {
		assert.Error(t, err, tc.UnmarshalErr.Error())
		return
	}
	impl, err := themes.Impl(c)
	if tc.ImplErr == nil {
		assert.NilError(t, err, "Impl")
	} else {
		assert.Error(t, err, tc.ImplErr.Error())
		return
	}

	assert.Equal(t, tc.Output(testString), impl.Do(testString))
}
