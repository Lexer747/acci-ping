// Use of this source code is governed by a GPL-2 license that can be found in the LICENSE file.
//
// Copyright 2024-2025 Lexer747
//
// SPDX-License-Identifier: GPL-2.0-only

package graph_test

import (
	"context"
	"fmt"
	"os"
	"path"
	"strings"
	"testing"
	"time"

	"github.com/Lexer747/acci-ping/draw"
	"github.com/Lexer747/acci-ping/graph"
	"github.com/Lexer747/acci-ping/graph/data"
	graphTh "github.com/Lexer747/acci-ping/graph/th"
	"github.com/Lexer747/acci-ping/gui/themes"
	"github.com/Lexer747/acci-ping/ping"
	"github.com/Lexer747/acci-ping/terminal"
	"github.com/Lexer747/acci-ping/utils/env"
	"github.com/Lexer747/acci-ping/utils/th"
	"gotest.tools/v3/assert"
	is "gotest.tools/v3/assert/cmp"
)

func init() {
	themes.LoadTheme(themes.NoTheme)
	graph.StartUp()
}

const (
	inputPath  = "data/testdata/input"
	outputPath = "data/testdata/output"
)

type FileTest struct {
	FileName         string
	Sizes            []terminal.Size
	OnlyDoLinear     bool
	TimeZoneOfFile   *time.Location
	TerminalWrapping th.TerminalWrapping
}

var StandardTestSizes = []terminal.Size{
	{Height: 40, Width: 80}, // Viewing height
	{Height: 25, Width: 80},
	{Height: 16, Width: 284}, // My small vscode window
	{Height: 30, Width: 300}, // My average vscode window
	{Height: 45, Width: 120}, // Tall vertical screen
	{Height: 74, Width: 354}, // Fullscreen
}

var winter = time.FixedZone("+0", 0)
var summer = time.FixedZone("+1", 3_600)

func TestFiles(t *testing.T) {
	t.Parallel()
	t.Run("Small", FileTest{
		FileName:       "small-2-02-08-2024",
		Sizes:          StandardTestSizes,
		TimeZoneOfFile: summer,
	}.Run)
	t.Run("Medium", FileTest{
		FileName:       "medium-395-02-08-2024",
		Sizes:          StandardTestSizes,
		TimeZoneOfFile: summer,
	}.Run)
	t.Run("Medium with Drops", FileTest{
		FileName:       "medium-309-with-induced-drops-02-08-2024",
		Sizes:          StandardTestSizes,
		TimeZoneOfFile: summer,
	}.Run)
	t.Run("Medium with minute Gaps", FileTest{
		FileName:       "medium-minute-gaps",
		Sizes:          StandardTestSizes,
		TimeZoneOfFile: summer,
	}.Run)
	t.Run("Medium with hour Gaps", FileTest{
		FileName:       "medium-hour-gaps",
		Sizes:          StandardTestSizes,
		TimeZoneOfFile: summer,
	}.Run)
	t.Run("Hotel", FileTest{
		FileName:       "medium-hotel",
		Sizes:          StandardTestSizes,
		TimeZoneOfFile: summer,
	}.Run)
	t.Run("Large Hotel", FileTest{
		FileName:       "large-hotel",
		Sizes:          StandardTestSizes,
		TimeZoneOfFile: summer,
	}.Run)
	t.Run("Gap", FileTest{
		FileName:       "long-gap",
		Sizes:          StandardTestSizes,
		TimeZoneOfFile: summer,
	}.Run)
	t.Run("Smoke Test", FileTest{
		FileName:       "smoke",
		Sizes:          StandardTestSizes,
		TimeZoneOfFile: winter,
	}.Run)
	t.Run("Span bugs", FileTest{
		FileName:       "huge-over-days",
		Sizes:          StandardTestSizes,
		TimeZoneOfFile: winter,
	}.Run)
	t.Run("verybad", FileTest{
		FileName:       "verybad-london",
		Sizes:          StandardTestSizes,
		TimeZoneOfFile: winter,
	}.Run)
}

func TestSmallWindows(t *testing.T) {
	t.Parallel()
	t.Run("Small Sizes", FileTest{
		FileName: "huge-over-days",
		Sizes: []terminal.Size{
			{Height: 1, Width: 1},
			{Height: 2, Width: 2},
			{Height: 3, Width: 3},
			{Height: 4, Width: 4},
			{Height: 5, Width: 5},
			{Height: 6, Width: 6},
			{Height: 7, Width: 7},
			{Height: 8, Width: 8},
		},
		TimeZoneOfFile:   winter,
		TerminalWrapping: th.WrapBuffer,
	}.Run)
}

func (ft FileTest) Run(t *testing.T) {
	t.Parallel()

	d := graphTh.GetFromFile(t, ft.getInputFileName())
	d = d.In(ft.TimeZoneOfFile)
	yAxis := []graph.YAxisScale{graph.Linear}
	if !ft.OnlyDoLinear {
		yAxis = append(yAxis, graph.Logarithmic)
	}
	for _, y := range yAxis {
		for _, size := range ft.Sizes {
			actualStrings := produceFrame(t, size, d, y, ft.TerminalWrapping)

			// ft.update(t, y, size, actualStrings)
			ft.assertEqual(t, y, size, actualStrings)
		}
	}
}

func (ft FileTest) assertEqual(t *testing.T, yAxis graph.YAxisScale, size terminal.Size, actualStrings []string) {
	t.Helper()
	outputFile := ft.getOutputFileName(yAxis, size)
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

func (ft FileTest) getInputFileName() string {
	return fmt.Sprintf("%s/%s.pings", inputPath, ft.FileName)
}
func (ft FileTest) getOutputFileName(yAxis graph.YAxisScale, size terminal.Size) string {
	if ft.OnlyDoLinear {
		return fmt.Sprintf("%s/%s/w%d-h%d.frame", outputPath, ft.FileName, size.Width, size.Height)
	}
	return fmt.Sprintf("%s/%s/%s-w%d-h%d.frame", outputPath, ft.FileName, yAxis, size.Width, size.Height)
}

//nolint:unused
func (ft FileTest) update(t *testing.T, yAxis graph.YAxisScale, size terminal.Size, actualStrings []string) {
	t.Helper()
	outputFile := ft.getOutputFileName(yAxis, size)
	err := os.MkdirAll(path.Dir(outputFile), 0o777)
	assert.NilError(t, err)
	err = os.WriteFile(outputFile, []byte(strings.Join(actualStrings, "\n")), 0o777)
	assert.NilError(t, err)
	t.Fail()
	t.Log("Only call update drawing once")
}

func produceFrame(
	t *testing.T,
	size terminal.Size,
	data *data.Data,
	yAxis graph.YAxisScale,
	terminalWrapping th.TerminalWrapping,
) []string {
	t.Helper()
	stdin, _, term, setTerm, err := th.NewTestTerminal()
	setTerm(size)
	ctx, cancel := context.WithCancel(t.Context())
	// cancel this, we don't want the graph collecting from the channel in the background
	cancel()
	assert.NilError(t, err)
	pingChannel := make(chan ping.PingResults)
	close(pingChannel)
	g := graph.NewGraph(ctx, graph.GraphConfiguration{
		Input:         pingChannel,
		Terminal:      term,
		DrawingBuffer: draw.NewPaintBuffer(),
		DebugStrict:   true,
		Data:          data,
		Presentation:  graph.Presentation{YAxisScale: yAxis},
	})
	defer func() { stdin.WriteCtrlC(t) }()
	output := th.MakeBuffer(size)
	return th.EmulateTerminal(g.ComputeFrame(), output, size, terminalWrapping)
}
