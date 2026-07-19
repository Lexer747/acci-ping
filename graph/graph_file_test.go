// Use of this source code is governed by a GPL-2 license that can be found in the LICENSE file.
//
// Copyright 2024-2026 Lexer747
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
	TimeZoneOfFile   *time.Location
	FileName         string
	Sizes            []terminal.Size
	TerminalWrapping th.TerminalWrapping
	OnlyDoLinear     bool
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
	for _, following := range []bool{false, true} {
		for _, y := range yAxis {
			for _, size := range ft.Sizes {
				actualStrings := produceFrame(t, size, d, y, following, ft.TerminalWrapping)

				// ft.update(t, y, size, following, actualStrings)
				ft.assertEqual(t, y, size, following, actualStrings)
			}
		}
	}
}

func (ft FileTest) assertEqual(t *testing.T, yAxis graph.YAxisScale, size terminal.Size, following bool, actualStrings []string) {
	t.Helper()
	outputFile := ft.getOutputFileName(yAxis, size, following)
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
func (ft FileTest) getOutputFileName(yAxis graph.YAxisScale, size terminal.Size, following bool) string {
	var b strings.Builder
	fmt.Fprintf(&b, "%s/%s/", outputPath, ft.FileName)
	if following {
		b.WriteString("following/")
	}
	if !ft.OnlyDoLinear {
		fmt.Fprintf(&b, "%s-", yAxis)
	}
	fmt.Fprintf(&b, "w%d-h%d.frame", size.Width, size.Height)
	return b.String()
}

//nolint:unused
func (ft FileTest) update(t *testing.T, yAxis graph.YAxisScale, size terminal.Size, following bool, actualStrings []string) {
	t.Helper()
	outputFile := ft.getOutputFileName(yAxis, size, following)
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
	following bool,
	terminalWrapping th.TerminalWrapping,
) []string {
	t.Helper()
	g := newTestGraph(t, size, data, yAxis, following)
	output := th.MakeBuffer(size)
	return th.EmulateTerminal(g.ComputeFrame(), output, size, terminalWrapping)
}

func newTestGraph(t *testing.T, size terminal.Size, d *data.Data, yAxis graph.YAxisScale, following bool) *graph.Graph {
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
		Data:          d,
		Presentation:  graph.Presentation{YAxisScale: yAxis, Following: following},
	})
	t.Cleanup(func() { stdin.WriteCtrlC(t) })
	return g
}

// TestXAxisFillsDrawableArea pins the invariant behind #18: the computed x-axis spans must be contiguous and
// together fill the whole drawable area (last endX == terminal width), across a range of widths.
func TestXAxisFillsDrawableArea(t *testing.T) {
	t.Parallel()
	// long-gap has multiple recording sessions, so the axis is split into several spans.
	d := graphTh.GetFromFile(t, inputPath+"/long-gap.pings")
	widths := []int{80, 120, 200, 300, 400, 500}
	for _, following := range []bool{false, true} {
		for _, w := range widths {
			size := terminal.Size{Height: 40, Width: w}
			g := newTestGraph(t, size, d, graph.Linear, following)
			bounds := g.ComputeXAxisBounds(size, following)
			assert.Assert(t, len(bounds) > 0)
			if !following {
				assert.Equal(t, bounds[0].StartX, 6, "first span starts after the y-axis labels (w=%d)", w)
			}
			last := bounds[len(bounds)-1]
			assert.Equal(t, last.EndX, size.Width, "last span fills to the terminal width (w=%d following=%t)", w, following)
			for i := 1; i < len(bounds); i++ {
				assert.Equal(t, bounds[i-1].EndX, bounds[i].StartX,
					"spans are contiguous at index %d (w=%d following=%t)", i, w, following)
			}
		}
	}
}

func TestEqualDurations(t *testing.T) {
	t.Parallel()
	d := data.NewData("example.com")
	base := time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)
	// Several good packets, all identical duration -> stats.Min == stats.Max
	for i := range 20 {
		d.AddPoint(ping.PingResults{
			Data: ping.PingDataPoint{
				Duration:  13 * time.Millisecond,
				Timestamp: base.Add(time.Duration(i) * time.Second),
			},
		})
	}
	// test is simply not to panic:
	size := terminal.Size{Height: 32, Width: 216}
	_ = produceFrame(t, size, d, graph.Linear, false, th.TerminalWrapping(0))
	_ = produceFrame(t, size, d, graph.Logarithmic, false, th.TerminalWrapping(0))
}
