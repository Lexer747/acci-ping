// Use of this source code is governed by a GPL-2 license that can be found in the LICENSE file.
//
// Copyright 2025 Lexer747
//
// SPDX-License-Identifier: GPL-2.0-only

package gui_test

import (
	"bytes"
	"crypto/sha256"
	"encoding/base64"
	"fmt"
	"os"
	"path"
	"strconv"
	"strings"
	"testing"

	"github.com/Lexer747/acci-ping/gui"
	"github.com/Lexer747/acci-ping/terminal"
	"github.com/Lexer747/acci-ping/utils/env"
	"github.com/Lexer747/acci-ping/utils/sliceutils"
	"github.com/Lexer747/acci-ping/utils/th"
	"gotest.tools/v3/assert"
)

// TestBox creates many permutations of boxes and asserts that they produce consistent outputs. Currently we
// don't store every expected result for a few reasons:
//
//   - Generate too many permutations
//   - Terminal wrapping makes some permutations consistent but pointless to manually expect.
func TestBox(t *testing.T) {
	t.Parallel()
	positions := generatePositions()

	styles := []gui.Style{
		gui.NoBorder,
		gui.RoundedCorners,
		gui.SharpCorners,
	}
	// NOTE: ensure that the first index doesn't contain an overlapping string as this is used to generate a
	// unique file name
	texts := [][]string{
		{"ABCDEFGHIJKLMNOPQRSTUVWXYZ"},
		{"1", "12", "123", "1234"},
		{"@ 1 @", "@ aaa @", "@ ... @", "@ www @"},
		{
			"The quick brown fox jumps over the lazy dog",
			"",
			"Waltz, bad nymph, for quick jigs vex.",
			"Sphinx of black quartz, judge my vow.",
			"How vexingly quick daft zebras jump!",
			"Pack my box with five dozen liquor jugs.",
			"#",
		},
	}
	textAlignments := []gui.HorizontalAlignment{
		gui.Centre,
		gui.Left,
		gui.Right,
	}
	sizes := []terminal.Size{
		{Height: 5, Width: 5},
		{Height: 30, Width: 30},
		{Height: 30, Width: 100},
	}

	// Now iterate over every permutation accumulating the hashes of the terminal windows produced by each box
	hashes := make([]hash, 0, len(sizes)*len(styles)*len(positions)*len(textAlignments)*len(texts))
	for _, size := range sizes {
		for _, style := range styles {
			for _, position := range positions {
				for _, align := range textAlignments {
					for _, text := range texts {
						tc := &testCase{
							size:     size,
							style:    style,
							position: position,
							align:    align,
							text:     text,
						}
						t.Run(tc.getOutputFileName(), tc.Run)
						hashes = append(hashes, tc.result)
					}
				}
			}
		}
	}
	// So we could assert on every single test file in [tc.Run] but this would mean committing all the test
	// golden's to the repo. Currently I see this as too many files (4000~), instead a comprise is to assert
	// that, as a whole, all files contain the same contents by doing a simple hash of the files and XOR'ing
	// these hashes together.
	//
	// This does make A-B comparisons harder since now the test will not directly inform which configuration
	// of parameters causes the error, but I'll cross that bridge when I come to it.
	assertCorrectHash(t, "box_test/hashes", hashes)
}

func assertCorrectHash(t *testing.T, file string, hashes []hash) {
	t.Helper()
	err := os.MkdirAll(path.Dir(file), 0o777)
	assert.NilError(t, err)
	fileBytes, err := os.ReadFile(file)

	computeHashFileContents := func(hashCount int, hash ...hash) []byte {
		return fmt.Appendf(nil,
			"Hash Count: %d\nHash XOR: %s",
			hashCount,
			computeHash(hash),
		)
	}

	if env.LOCAL_FRAME_DIFFS() {
		if err != nil {
			fileBytes = computeHashFileContents(0, sha256.Sum256([]byte(`Empty`)))
		}
	} else {
		assert.NilError(t, err)
	}
	lines := strings.Split(string(fileBytes), "\n")
	hashCount, err := strconv.Atoi(strings.Split(lines[0], "Hash Count: ")[1])
	assert.NilError(t, err)
	if env.LOCAL_FRAME_DIFFS() {
		writeHashFile := func(t *testing.T, hashes []hash, file string) {
			t.Helper()
			toWrite := computeHashFileContents(len(hashes), hashes...)
			actualFile := file + ".actual"
			err := os.WriteFile(actualFile, toWrite, 0o777)
			assert.NilError(t, err)
			t.Logf("Diff in hash, inspect sub tests %s", actualFile)
			t.FailNow()
		}
		if hashCount != len(hashes) {
			writeHashFile(t, hashes, file)
		}

		expectedHashBase64 := strings.Split(lines[1], "Hash XOR: ")[1]
		resultBase64 := computeHash(hashes)
		if resultBase64 != expectedHashBase64 {
			writeHashFile(t, hashes, file)
		}
	} else {
		assert.Equal(t, hashCount, len(hashes), "number of hashes created is different")

		expectedHashBase64 := strings.Split(lines[1], "Hash XOR: ")[1]
		resultBase64 := computeHash(hashes)
		assert.Equal(t, expectedHashBase64, resultBase64, "hashes are not equal, check individual test frames for possible reasons.")
	}
}

func computeHash(hashes []hash) string {
	var result hash
	for _, hash := range hashes {
		for i, b := range hash {
			// XOR is not cryptographically sane (something something linear algebra), however it is sane
			// against arbitrary collisions which is all we're using the hash for. It also allows arbitrary
			// re-ordering of the input hashes, as a bonus it's quick.
			result[i] = result[i] ^ b
		}
	}
	return base64.StdEncoding.EncodeToString(result[:])
}

func (tc *testCase) Run(t *testing.T) {
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
	result := th.EmulateTerminal(buffer.String(), b, tc.size, th.SilentlyDrop)
	tc.writeToDisk(t, result)
	tc.writeToResult(result)
}

func (tc *testCase) writeToResult(result []string) {
	tc.result = sha256.Sum256(
		[]byte(strings.Join(result, "\n")),
	)
}

func (tc *testCase) writeToDisk(t *testing.T, actualStrings []string) {
	t.Helper()
	if !env.LOCAL_FRAME_DIFFS() {
		return
	}
	outputFile := tc.getOutputFileName()
	err := os.MkdirAll(path.Dir(outputFile), 0o777)
	assert.NilError(t, err)
	actualJoined := strings.Join(actualStrings, "\n")
	actualOutput := outputFile + ".no-commit-actual"
	err = os.WriteFile(actualOutput, []byte(actualJoined), 0o777)
	assert.NilError(t, err)
	t.Logf("Diff in outputs see %s", actualOutput)
}

func (tc *testCase) getOutputFileName() string {
	ret := "box_test" +
		"/style-" + tc.style.String() +
		"/position-" + tc.position.String() +
		"/align-" + tc.align.String() +
		"/size-" + tc.size.String() +
		"/text-" + tc.text[0]
	return strings.ReplaceAll(ret, " ", "_")
}

type testCase struct {
	size     terminal.Size
	style    gui.Style
	position gui.Position
	align    gui.HorizontalAlignment
	text     []string
	result   hash
}

// hash is the internal hash used by the test cases, in theory this could be changed if collisions become an issue.
type hash [sha256.Size]byte

func generatePositions() []gui.Position {
	hozAlignments := []gui.HorizontalAlignment{gui.Centre, gui.Left, gui.Right}
	verAlignments := []gui.VerticalAlignment{gui.Top, gui.Middle, gui.Bottom}
	paddings := []gui.Padding{
		gui.NoPadding,
		{Top: 1, Bottom: 0, Left: 0, Right: 1},
		{Top: 2, Bottom: 0, Left: 0, Right: 2},
		{Top: 0, Bottom: 1, Left: 1, Right: 0},
		{Top: 0, Bottom: 2, Left: 2, Right: 0},
	}
	positions := []gui.Position{}
	for _, hoz := range hozAlignments {
		for _, ver := range verAlignments {
			for _, padding := range paddings {
				positions = append(positions, gui.Position{
					Vertical:   ver,
					Horizontal: hoz,
					Padding:    padding,
				})
			}
		}
	}
	return positions
}
