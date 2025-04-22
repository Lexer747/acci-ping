// Use of this source code is governed by a GPL-2 license that can be found in the LICENSE file.
//
// Copyright 2024-2025 Lexer747
//
// SPDX-License-Identifier: GPL-2.0-only

package graph

import (
	"context"
	"fmt"
	"io"
	"log/slog"
	"strings"
	"sync"
	"time"

	"github.com/Lexer747/acci-ping/draw"
	"github.com/Lexer747/acci-ping/graph/data"
	"github.com/Lexer747/acci-ping/graph/graphdata"
	"github.com/Lexer747/acci-ping/gui"
	"github.com/Lexer747/acci-ping/ping"
	"github.com/Lexer747/acci-ping/terminal"
	"github.com/Lexer747/acci-ping/utils/check"
)

type Graph struct {
	Term *terminal.Terminal

	ui gui.GUI

	debugStrict bool

	sinkAlive   bool
	dataChannel <-chan ping.PingResults

	pingsPerMinute float64

	data *graphdata.GraphData

	frameMutex *sync.Mutex
	lastFrame  frame

	drawingBuffer *draw.Buffer

	currentControl *controlState
	controlPlane   <-chan Control
}

type Control struct {
	FollowLatestSpan bool
}

type controlState struct {
	Control
	m *sync.Mutex
}

type GraphConfiguration struct {
	// Input will be owned by the graph and represents the source of data for the graph to plot.
	Input <-chan ping.PingResults
	// Terminal is the underlying terminal the graph will draw to and perform all external I/O too. The graph
	// takes ownership of the terminal.
	Terminal *terminal.Terminal
	// Gui is the external GUI interface which the terminal will call [gui.Draw] on each frame if the gui
	// component requires it. This
	Gui            gui.GUI
	PingsPerMinute float64
	URL            string
	DrawingBuffer  *draw.Buffer
	InitialControl Control
	ControlPlane   <-chan Control
	DebugStrict    bool

	// Optionals:
	Data *data.Data
}

func StartUp() {
	xAxisStartup()
	yAxisStartup()
	drawWindowStartUp()
}

func NewGraph(ctx context.Context, cfg GraphConfiguration) *Graph {
	if cfg.Data == nil {
		cfg.Data = data.NewData(cfg.URL)
	}
	if cfg.Gui == nil {
		cfg.Gui = gui.NoGUI()
	}
	g := &Graph{
		Term:           cfg.Terminal,
		sinkAlive:      true,
		dataChannel:    cfg.Input,
		pingsPerMinute: cfg.PingsPerMinute,
		data:           graphdata.NewGraphData(cfg.Data),
		frameMutex:     &sync.Mutex{},
		lastFrame:      frame{},
		drawingBuffer:  cfg.DrawingBuffer,
		ui:             cfg.Gui,
		debugStrict:    cfg.DebugStrict,
		controlPlane:   cfg.ControlPlane,
		currentControl: &controlState{Control: cfg.InitialControl, m: &sync.Mutex{}},
	}
	if ctx != nil {
		// A nil context is valid: It means that no new data is expected and the input channel isn't active
		go g.sink(ctx)
	}
	return g
}

// Run holds the thread an listens on it's ping channel continuously, drawing a new graph every time a new
// packet comes in. It only returns a fatal error in which case it couldn't continue drawing (although it may
// still panic). It will return [terminal.UserControlCErr] if the thread was cancelled by the user.
//
// Since this runs in a concurrent sense any method is thread safe but therefore may also block if another
// thread is already holding the lock.
//
// Returns
//   - The graph main function
//   - the defer function which will restore the terminal to the correct state
//   - a channel containing all the terminal size updates
//   - an error if creating any of the above failed.
func (g *Graph) Run(
	ctx context.Context,
	stop context.CancelCauseFunc,
	fps int,
	listeners []terminal.ConditionalListener,
	fallbacks []terminal.Listener,
) (func() error, func(), <-chan terminal.Size, error) {
	timeBetweenFrames := getTimeBetweenFrames(fps, g.pingsPerMinute)
	frameRate := time.NewTicker(timeBetweenFrames)
	cleanup, err := g.Term.StartRaw(ctx, stop, listeners, fallbacks)
	if err != nil {
		return nil, cleanup, nil, err
	}
	terminalUpdates := make(chan terminal.Size)
	graph := func() error {
		size := g.Term.GetSize()
		defer close(terminalUpdates)
		for {
			select {
			case <-ctx.Done():
				return context.Cause(ctx)
			case <-frameRate.C:
				if err = g.Term.UpdateSize(); err != nil {
					return err
				}
				if size != g.Term.GetSize() {
					slog.Debug("sending size update", "size", size)
					terminalUpdates <- size
					size = g.Term.GetSize()
				}
				g.currentControl.m.Lock()
				toWrite := g.computeFrame(timeBetweenFrames, g.currentControl.FollowLatestSpan, true)
				g.currentControl.m.Unlock()
				err = toWrite(g.Term)
				if err != nil {
					return err
				}
			}
		}
	}
	if g.controlPlane != nil {
		go g.handleControl(ctx)
	}
	return graph, cleanup, terminalUpdates, err
}

// OneFrame doesn't run the graph but runs all the code to create and print a single frame to the terminal.
func (g *Graph) OneFrame() error {
	if err := g.Term.ClearScreen(terminal.MoveHome); err != nil {
		return err
	}
	if err := g.Term.UpdateSize(); err != nil {
		return err
	}
	toWrite := g.computeFrame(0, g.currentControl.FollowLatestSpan, false)
	return toWrite(g.Term)
}

// LastFrame will return the last graphical frame printed to the terminal.
func (g *Graph) LastFrame() string {
	g.frameMutex.Lock()
	defer g.frameMutex.Unlock()
	var b strings.Builder
	err := g.lastFrame.framePainterNoGui(&b)
	check.NoErr(err, "While painting frame to string buffer")
	return b.String()
}

// Summarise will summarise the graph's backed data according to the [*graphdata.GraphData.String] function.
func (g *Graph) Summarise() string {
	g.frameMutex.Lock()
	defer g.frameMutex.Unlock()
	return strings.ReplaceAll(g.data.String(), "| ", "\n\t")
}

func (g *Graph) sink(ctx context.Context) {
	for {
		select {
		case <-ctx.Done():
			g.sinkAlive = false
			return
		case p, ok := <-g.dataChannel:
			// TODO configure logging channels
			// slog.Debug("graph sink data received", "packet", p)
			if !ok {
				g.sinkAlive = false
				return
			}
			g.data.AddPoint(p)
		}
	}
}

func (g *Graph) handleControl(ctx context.Context) {
	for {
		select {
		case <-ctx.Done():
			return
		case p, ok := <-g.controlPlane:
			if !ok {
				return
			}
			// Note we don't need the mutex while reading the values.
			if p.FollowLatestSpan {
				// But we do need it while writing
				g.currentControl.m.Lock()
				g.currentControl.FollowLatestSpan = !g.currentControl.FollowLatestSpan
				g.currentControl.m.Unlock()
			}
			slog.Debug("switching to:", "FollowLatestSpan", g.currentControl.FollowLatestSpan)
		}
	}
}

func (g *Graph) checkf(shouldBeTrue bool, format string, a ...any) {
	if g.debugStrict {
		check.Checkf(shouldBeTrue, format, a...)
	} else if !shouldBeTrue {
		slog.Error("check failed: " + fmt.Sprintf(format, a...))
	}
}

type frame struct {
	PacketCount       int64
	yAxis             drawingYAxis
	xAxis             drawingXAxis
	framePainter      func(io.Writer) error
	framePainterNoGui func(io.Writer) error
	spinnerIndex      int
	followLatestSpan  bool
}

func (f frame) Match(s terminal.Size, followLatestSpan bool) bool {
	return f.xAxis.size == s.Width && f.yAxis.size == s.Height &&
		f.followLatestSpan == followLatestSpan
}

func (f frame) Size() terminal.Size {
	return terminal.Size{Height: f.xAxis.size, Width: f.yAxis.size}
}
