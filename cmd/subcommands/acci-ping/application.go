// Use of this source code is governed by a GPL-2 license that can be found in the LICENSE file.
//
// Copyright 2024-2025 Lexer747
//
// SPDX-License-Identifier: GPL-2.0-only

package acciping

import (
	"context"
	"fmt"
	"io"
	"maps"
	"os"
	"slices"
	"strconv"
	"time"

	"github.com/Lexer747/acci-ping/draw"
	"github.com/Lexer747/acci-ping/files"
	"github.com/Lexer747/acci-ping/graph"
	"github.com/Lexer747/acci-ping/graph/data"
	"github.com/Lexer747/acci-ping/gui"
	"github.com/Lexer747/acci-ping/gui/themes"
	"github.com/Lexer747/acci-ping/ping"
	"github.com/Lexer747/acci-ping/terminal"
	"github.com/Lexer747/acci-ping/utils/application"
	"github.com/Lexer747/acci-ping/utils/backoff"
	"github.com/Lexer747/acci-ping/utils/channels"
	"github.com/Lexer747/acci-ping/utils/errors"
	"github.com/Lexer747/acci-ping/utils/exit"
)

type Application struct {
	*GUI
	g    *graph.Graph
	term *terminal.Terminal

	toUpdate *os.File
	config   Config
	// this doesn't need a mutex because we ensure that no two threads have access to the same byte index (I
	// think this is fine when the slice doesn't grow).
	drawBuffer *draw.Buffer

	errorChannel chan error
	controlPlane chan graph.Control
}

func (app *Application) Run(
	ctx context.Context,
	cancelFunc context.CancelCauseFunc,
	channel <-chan ping.PingResults,
	existingData *data.Data,
) error {
	var fileData *data.Data
	var graphChannel, fileChannel <-chan ping.PingResults
	if app.toUpdate != nil {
		// The ping channel which is already running needs to be duplicated, providing one to the Graph and second
		// to a file writer. This de-couples the processes, we don't want the GUI to affect storing data and vice
		// versa.
		graphChannel, fileChannel = channels.TeeBufferedChannel(ctx, channel, *app.config.pingBufferingLimit)
		var err error
		fileData, err = duplicateData(app.toUpdate)
		exit.OnError(err)
	} else {
		// We don't need to duplicate the channel since we are not writing anything to a file
		graphChannel = channel
	}

	app.drawBuffer = draw.NewPaintBuffer()

	helpCh := make(chan rune)
	controlCh := make(chan graph.Control)
	app.addFallbackListener(helpAction(helpCh))

	// TODO make it a command line setting to populate at start
	control := graph.Presentation{
		Following:  false,
		YAxisScale: graph.Linear,
	}

	// The graph will take ownership of the data channel and data pointer.
	app.g = graph.NewGraph(
		ctx,
		graph.GraphConfiguration{
			Input:          graphChannel,
			Terminal:       app.term,
			Gui:            app.GUI,
			PingsPerMinute: *app.config.pingsPerMinute,
			DrawingBuffer:  app.drawBuffer,
			Presentation:   control,
			ControlPlane:   app.controlPlane,
			DebugStrict:    *app.config.debugStrict,
			Data:           existingData,
		},
	)
	_ = app.g.Term.ClearScreen(terminal.UpdateSizeAndMoveHome)

	if *app.config.testErrorListener {
		app.makeErrorGenerator()
	}
	app.addListener('f', func(rune) error {
		control.Following = !control.Following
		update := graph.Control{
			FollowLatestSpan: graph.Change[bool]{
				DidChange: true,
				Value:     control.Following,
			},
			YAxisScale: graph.Change[graph.YAxisScale]{DidChange: false},
		}
		app.controlPlane <- update
		controlCh <- update
		return nil
	})
	app.addListener('l', func(rune) error {
		switch control.YAxisScale {
		case graph.Linear:
			control.YAxisScale = graph.Logarithmic
		case graph.Logarithmic:
			control.YAxisScale = graph.Linear
		}
		update := graph.Control{
			FollowLatestSpan: graph.Change[bool]{DidChange: false},
			YAxisScale: graph.Change[graph.YAxisScale]{
				DidChange: true,
				Value:     control.YAxisScale,
			},
		}
		app.controlPlane <- update
		controlCh <- update
		return nil
	})

	defer close(app.errorChannel)
	defer close(app.controlPlane)
	defer close(helpCh)
	defer close(controlCh)
	// Very high FPS is good for responsiveness in the UI (since it's locked) and re-drawing on a re-size.
	// TODO add UI listeners, zooming, changing ping speed - etc
	graph, cleanup, terminalSizeUpdates, err := app.g.Run(ctx, cancelFunc, 120, app.listeners(), app.fallbacks)
	termRecover := func() {
		_ = app.term.ClearScreen(terminal.UpdateSize)
		cleanup()
		if err := recover(); err != nil {
			panic(err)
		}
	}
	terminalUpdates := channels.FanInFanOut(ctx, terminalSizeUpdates, 0, 3)

	// https://go.dev/ref/spec#Handling_panics
	// https://go.dev/blog/defer-panic-and-recover
	//
	// Each go routine needs to handle a panic in the same way.
	if fileData != nil {
		go func() {
			defer termRecover()
			app.writeToFile(ctx, fileData, fileChannel)
		}()
	}
	go func() {
		defer termRecover()
		app.toastNotifications(ctx, terminalUpdates[0])
	}()
	go func() {
		defer termRecover()
		app.help(ctx, !*app.config.hideHelpOnStart, helpCh, terminalUpdates[1])
	}()
	go func() {
		defer termRecover()
		app.showControls(ctx, control, controlCh, terminalUpdates[2])
	}()
	defer termRecover()
	exit.OnError(err)
	return graph()
}

func appThemeStartUp() {
	helpStartup()
	graph.StartUp()
}

func (app *Application) Init(ctx context.Context, c Config) (<-chan ping.PingResults, *data.Data) {
	app.config = c
	app.errorChannel = make(chan error)
	app.controlPlane = make(chan graph.Control)
	app.GUI = newGUIState()
	p := ping.NewPing()
	var err error
	app.term, err = makeTerminal(c.debuggingTermSize)
	exit.OnError(err) // If we can't open the terminal for any reason we reasonably can't do anything this program offers.

	var existingData *data.Data
	if *c.filePath != "" {
		existingData, app.toUpdate = loadFile(*c.filePath, *c.url)
	} else {
		existingData = data.NewData(*c.url)
	}

	channel, err := p.CreateChannel(ctx, existingData.URL, *c.pingsPerMinute, *c.pingBufferingLimit)
	// If Creating the channel has an error this means we cannot continue, the network errors are already
	// wrapped and retried by this channel, other errors imply some larger problem
	exit.OnError(err)
	err = application.LoadTheme(*c.theme, app.term)
	appThemeStartUp()
	go func() { app.errorChannel <- err }()

	return channel, existingData
}

func (app *Application) Finish() {
	_ = app.term.ClearScreen(terminal.UpdateSize)
	app.term.Print(app.g.LastFrame())
	if *app.config.filePath != "" {
		app.term.Print("\n\n# Summary\nData Successfully recorded in file '" + *app.config.filePath + "'\n\t" +
			app.g.Summarise() + "\n")
	} else {
		app.term.Print("\n\n# Summary\nData not saved, use `-file [FILE_NAME]` to save recordings in future.\n\t" +
			app.g.Summarise() + "\n")
	}
}

func (app *Application) writeToFile(ctx context.Context, ourData *data.Data, input <-chan ping.PingResults) {
	defer app.toUpdate.Close()
	exp := backoff.NewExponentialBackoff(500 * time.Millisecond)
	for {
		select {
		case <-ctx.Done():
			return
		case p, ok := <-input:
			if !ok {
				return
			}
			ourData.AddPoint(p)
			_, err := app.toUpdate.Seek(0, 0)
			if err != nil {
				app.errorChannel <- err
				exp.Wait()
				continue
			}
			err = ourData.AsCompact(app.toUpdate)
			if err != nil {
				app.errorChannel <- err
				exp.Wait()
				continue
			}
			exp.Success()
		}
	}
}

func (app *Application) makeErrorGenerator() {
	app.addListener('e', func(r rune) error {
		go func() { app.errorChannel <- errors.New("Test Error") }()
		return nil
	})
	helpCopy = append(helpCopy,
		gui.Typography{ToPrint: "Press " + themes.Positive("e") + " to generate a test error.", TextLen: 6 + 1 + 26, Alignment: gui.Left},
	)
}

func (app *Application) addListener(r rune, Action func(rune) error) {
	if _, found := app.listeningChars[r]; found {
		panic(fmt.Sprintf("Adding more than one listener for '%v'", r))
	}
	app.listeningChars[r] = terminal.ConditionalListener{
		Listener: terminal.Listener{
			Action: Action,
			Name:   "GUI Listener " + strconv.QuoteRune(r),
		},
		Applicable: func(in rune) bool {
			return in == r
		},
	}
}

func (app *Application) addFallbackListener(Action func(rune) error) {
	app.fallbacks = append(app.fallbacks, terminal.Listener{
		Action: Action,
		Name:   "GUI Fallback Listener",
	})
}

func (app *Application) listeners() []terminal.ConditionalListener {
	ret := make([]terminal.ConditionalListener, 0, len(app.listeningChars))
	return slices.AppendSeq(ret, maps.Values(app.listeningChars))
}

func duplicateData(f *os.File) (*data.Data, error) {
	d := &data.Data{}
	file, fileErr := io.ReadAll(f)
	_, readingErr := d.FromCompact(file)
	return d, errors.Join(fileErr, readingErr)
}

// TODO incremental read/writes, get the URL ASAP then start the channel, then incremental continuation.
func loadFile(file, url string) (*data.Data, *os.File) {
	// TODO this currently panics if the url's don't match we should do better
	d, f, err := files.LoadOrCreateFile(file, url)
	exit.OnError(err)
	return d, f
}

func makeTerminal(termSize *string) (*terminal.Terminal, error) {
	if termSize != nil && *termSize != "" {
		s, err := terminal.NewSize(*termSize)
		if err != nil {
			return nil, err
		}
		return terminal.NewDebuggingTerminal(s)
	} else {
		return terminal.NewTerminal()
	}
}
