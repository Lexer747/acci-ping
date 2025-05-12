// Use of this source code is governed by a GPL-2 license that can be found in the LICENSE file.
//
// Copyright 2025 Lexer747
//
// SPDX-License-Identifier: GPL-2.0-only

package gui_test

import (
	"bytes"
	"os"
	"strings"
	"testing"

	"github.com/Lexer747/acci-ping/gui"
	"github.com/Lexer747/acci-ping/terminal"
	"github.com/Lexer747/acci-ping/utils/env"
	"github.com/Lexer747/acci-ping/utils/sliceutils"
	"github.com/Lexer747/acci-ping/utils/th"

	"gotest.tools/v3/assert"
	is "gotest.tools/v3/assert/cmp"
)

func TestBox(t *testing.T) {
	t.Parallel()
	positions := newFunction()

	styles := []gui.Style{gui.NoBorder, gui.RoundedCorners, gui.SharpCorners}
	texts := [][]string{
		{"ABCDEFGHIJKLMNOPQRSTUVWXYZ"},
		{"1", "12", "123", "1234"},
		{"@ 1 @", "@ aaa @", "@ ... @", "@ www @"},
	}
	textAlignments := []gui.HorizontalAlignment{gui.Centre, gui.Left, gui.Right}
	sizes := []terminal.Size{
		{Height: 5, Width: 5},
		{Height: 30, Width: 30},
		{Height: 30, Width: 300},
	}

	for _, size := range sizes {
		for _, style := range styles {
			for _, position := range positions {
				for _, align := range textAlignments {
					for _, text := range texts {
						tc := testCase{
							size:     size,
							style:    style,
							position: position,
							align:    align,
							text:     text,
						}
						t.Run(tc.getOutputFileName(), tc.Run)
					}
				}
			}
		}
	}

}

func (tc testCase) Run(t *testing.T) {
	t.Parallel()
	testBox := gui.Box{
		BoxText: sliceutils.Map(tc.text, func(s string) gui.Typography {
			return gui.Typography{
				ToPrint:        s,
				LenFromToPrint: true,
				Alignment:      tc.align,
			}
		}),
		Position: tc.position,
		Style:    tc.style,
	}
	buffer := &bytes.Buffer{}
	testBox.Draw(tc.size, buffer)
	b := th.MakeBuffer(tc.size)
	result := th.PlayAnsiOntoStringBuffer(buffer.String(), b, tc.size, false)
	tc.assertEqual(t, result)
}

func (tc testCase) assertEqual(t *testing.T, actualStrings []string) {
	t.Helper()
	outputFile := tc.getOutputFileName()
	expectedBytes, err := os.ReadFile(outputFile)
	assert.NilError(t, err)
	actualJoined := strings.Join(actualStrings, "\n")
	expected := string(expectedBytes)
	if env.LOCAL_FRAME_DIFFS() {
		actualOutput := outputFile + ".actual"
		if expected != actualJoined {
			err := os.WriteFile(actualOutput, []byte(actualJoined), 0o777)
			assert.NilError(t, err)
			t.Logf("Diff in outputs see %s", actualOutput)
			t.Fail()
		} else {
			os.Remove(actualOutput)
		}
	} else {
		assert.Check(t, is.Equal(expected, actualJoined), outputFile)
	}
}

func (tc testCase) getOutputFileName() string {
	ret := "size-" + tc.size.String() +
		"#style-" + tc.style.String() +
		"#position-" + tc.position.String() +
		"#align-" + tc.align.String() +
		"#text-" + strings.Join(tc.text, ",")
	return strings.ReplaceAll(ret, " ", "_")
}

type testCase struct {
	size     terminal.Size
	style    gui.Style
	position gui.Position
	align    gui.HorizontalAlignment
	text     []string
}

func newFunction() []gui.Position {
	hozAlignments := []gui.HorizontalAlignment{gui.Centre, gui.Left, gui.Right}
	verAlignments := []gui.VerticalAlignment{gui.Top, gui.Middle, gui.Bottom}
	paddings := []gui.Padding{
		gui.NoPadding,
		{Top: 1, Bottom: 1, Left: 1, Right: 1},
		{Top: 2, Bottom: 2, Left: 2, Right: 2},
	}
	positions := []gui.Position{}
	for _, hoz := range hozAlignments {
		for _, ver := range verAlignments {
			for _, p := range paddings {
				positions = append(positions, gui.Position{
					Vertical:   ver,
					Horizontal: hoz,
					Padding:    p,
				})
			}
		}
	}
	return positions
}
