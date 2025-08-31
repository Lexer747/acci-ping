// Use of this source code is governed by a GPL-2 license that can be found in the LICENSE file.
//
// Copyright 2025 Lexer747
//
// SPDX-License-Identifier: GPL-2.0-only

package themes_test

import (
	"testing"

	"gotest.tools/v3/assert"

	"github.com/Lexer747/acci-ping/gui/themes"
	"github.com/Lexer747/acci-ping/utils/errors"
)

func TestCSStoRGB(t *testing.T) {
	t.Parallel()
	happy := []CSSTest{
		{input: "#1e750b", R: 30, G: 117, B: 11},
		{input: "#00FF00", R: 0, G: 255, B: 0},
		{input: "#FF0000", R: 255, G: 0, B: 0},
		{input: "#0000FF", R: 0, G: 0, B: 255},
		{input: "0000FF", R: 0, G: 0, B: 255},
	}
	for _, tc := range happy {
		t.Run(tc.input, tc.Run)
	}
}

//nolint:lll
func TestCSStoRGB_Errors(t *testing.T) {
	t.Parallel()
	sad := []CSSTest{
		{input: "#1e750b000", Err: errors.Errorf(`Wrong number of digits for colour "#1e750b000", should be 6 hex digits`)},
		{input: "###112233", Err: errors.Errorf(`Wrong number of digits for colour "###112233", should be 6 hex digits`)},
		{input: "#GGGGGG", Err: errors.Errorf(`Couldn't parse RGB values for "#GGGGGG" caused by: strconv.ParseInt: parsing "GG": invalid syntax
strconv.ParseInt: parsing "GG": invalid syntax
strconv.ParseInt: parsing "GG": invalid syntax`)},
		{input: "#-1-2-3", Err: errors.Errorf(`Couldn't parse RGB values for "#-1-2-3" caused by: red component out of range -1, should be within 0 and 255
green component out of range -2, should be within 0 and 255
green component out of range -3, should be within 0 and 255`)},
	}
	for _, tc := range sad {
		t.Run(tc.input, tc.Run)
	}
}

type CSSTest struct {
	Err     error
	input   string
	R, G, B uint8
}

func (tc CSSTest) Run(t *testing.T) {
	t.Parallel()
	r, g, b, err := themes.CSSStrToRGB(tc.input)

	if tc.Err == nil {
		assert.NilError(t, err)
	} else {
		assert.Error(t, err, tc.Err.Error())
	}
	assert.Equal(t, tc.R, r)
	assert.Equal(t, tc.G, g)
	assert.Equal(t, tc.B, b)
}
