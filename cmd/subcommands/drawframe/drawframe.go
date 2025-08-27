// Use of this source code is governed by a GPL-2 license that can be found in the LICENSE file.
//
// Copyright 2024-2025 Lexer747
//
// SPDX-License-Identifier: GPL-2.0-only

package drawframe

import (
	"context"
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/Lexer747/acci-ping/cmd/tab_completion/tabflags"
	"github.com/Lexer747/acci-ping/draw"
	"github.com/Lexer747/acci-ping/files"
	"github.com/Lexer747/acci-ping/graph"
	"github.com/Lexer747/acci-ping/graph/data"
	"github.com/Lexer747/acci-ping/gui/themes"
	"github.com/Lexer747/acci-ping/terminal"
	"github.com/Lexer747/acci-ping/terminal/ansi"
	"github.com/Lexer747/acci-ping/utils/application"
	"github.com/Lexer747/acci-ping/utils/check"
	"github.com/Lexer747/acci-ping/utils/errors"
	"github.com/Lexer747/acci-ping/utils/exit"
	"github.com/Lexer747/acci-ping/utils/flags"
)

type Config struct {
	*application.BuildInfo
	*application.SharedFlags
	*tabflags.FlagSet

	debugFollow *bool
	termSize    *string
	theme       *string
	yAxisScale  *bool
}

func GetFlags(info *application.BuildInfo) *Config {
	f := flag.NewFlagSet("", flag.ContinueOnError)
	tf := tabflags.NewAutoCompleteFlagSet(f, true, ".pings")
	sf := application.NewSharedFlags(tf)
	ret := &Config{
		BuildInfo:   info,
		SharedFlags: sf,
		FlagSet:     tf,

		debugFollow: tf.Bool("debug-follow", false, "switches drawing to followLastSpan."),
		termSize: tf.String("term-size", "", "controls the terminal size and fixes it to the input,"+
			" input is in the form \"<H>x<W>\" e.g. 20x80. H and W must be integers - where H == height, and W == width of the terminal.",
			tabflags.AutoComplete{Choices: []string{"15x80", "20x85", "HxW"}}),
		theme: tf.String("theme", "", "the colour theme (either a path or builtin theme name) to use for the program,\n"+
			"if empty this will try to get the background colour of the terminal and pick the\n"+
			"built in dark or light theme based on the colour found.\n"+
			"There's also the builtin themes:\n"+strings.Join(themes.DescribeBuiltins(), "\n")+
			"\nSee the docs "+ansi.Blue("https://github.com/Lexer747/acci-ping/blob/main/docs/themes.md")+
			" for how to create custom themes.",
			tabflags.AutoComplete{Choices: themes.GetBuiltInNames(), WantsFile: true, FileExt: ".json"}),
		yAxisScale: tf.Bool("log-scale", false, "switches the y-axis to be in logarithmic scaling instead of linear"),
	}
	f.Usage = func() {
		var programName = "acci-ping " + ansi.Green("drawframe")

		w := flag.CommandLine.Output()
		fmt.Fprintf(w, "Usage of %s: reads '.pings' files and outputs the final frame of the capture\n"+
			"\t drawframe [options] FILE\n\n"+
			"e.g. '%s my_ping_capture.ping'\n", programName, programName)
		if ret.HelpDebug() {
			flags.PrintFlagsFilter(ret.FlagSet.FlagSet, flags.NoFilter())
		} else {
			flags.PrintFlagsFilter(ret.FlagSet.FlagSet, flags.ExcludePrefix("debug"))
		}
	}
	return ret
}

func RunDrawFrame(c *Config) {
	check.Check(c.Parsed(), "flags not parsed")
	closeLogFile := c.InitLogging(c.BuildInfo)
	defer closeLogFile()
	closeCPUProfile := c.InitCPUProfiling()
	defer closeCPUProfile()
	closeMemProfile := c.InitMemProfile()
	defer closeMemProfile()

	toPrint := c.Args()
	if len(toPrint) == 0 {
		fmt.Fprint(os.Stderr, "No files found, exiting. Use -h/--help to print usage instructions.\n")
		exit.Success()
	}

	term, err := makeTerminal(c.termSize, c.DebugStrict())
	exit.OnErrorMsg(err, "failed to open terminal to draw")

	err = application.LoadTheme(*c.theme, term)
	exit.OnErrorMsg(err, "failed to use theme")
	graph.StartUp()

	for _, path := range toPrint {
		run(term, path, c.Profiling(), *c.yAxisScale, *c.debugFollow, c.DebugStrict())
	}
	fmt.Println()
	fmt.Println()
	fmt.Println()
}

func run(term *terminal.Terminal, path string, profiling, logScale, debugFollow, debugStrict bool) {
	fs, err := os.Stat(path)
	exit.OnErrorMsgf(err, "Couldn't stat path %q, failed with", path)
	if fs.IsDir() {
		err := filepath.WalkDir(path, func(p string, d os.DirEntry, err error) error {
			if filepath.Ext(p) != ".pings" {
				return nil
			}
			do(p, term, profiling, logScale, debugFollow, debugStrict)
			return nil
		})
		exit.OnErrorMsgf(err, "Couldn't walk path %q, failed with", path)
	} else {
		do(path, term, profiling, logScale, debugFollow, debugStrict)
	}
}

func do(path string, term *terminal.Terminal, profiling, logScale, debugFollow, debugStrict bool) {
	d, f, err := files.LoadFile(path)
	exit.OnErrorMsg(err, "Couldn't open and read file, failed with")
	f.Close()
	if err != nil {
		panic(err.Error())
	}

	scale := graph.Linear
	if logScale {
		scale = graph.Logarithmic
	}
	g := makeGraph(term, debugFollow, scale, debugStrict, d)

	// TODO don't profile like this when iterating over a folder of inputs.
	if profiling {
		timer := time.NewTimer(time.Second * 30)
		running := true
		for running {
			printGraph(g)
			g.ClearForPerfTest()
			select {
			case <-timer.C:
				running = false
			default:
			}
		}
	} else {
		printGraph(g)
	}
}

func makeTerminal(termSize *string, debugStrict bool) (*terminal.Terminal, error) {
	fallback := func(t *terminal.Terminal, err error) (*terminal.Terminal, error) {
		if err == nil || debugStrict {
			return t, err
		}

		// In the case that we're not in strict mode so we should retry with a sane default terminal size.
		if errors.Is(err, terminal.TermSizeError) {
			return terminal.NewDebuggingTerminal(terminal.Size{Height: 20, Width: 100})
		}
		return t, err
	}
	if termSize != nil && *termSize != "" {
		t, err := terminal.NewParsedFixedSizeTerminal(*termSize)
		return fallback(t, err)
	} else {
		t, err := terminal.NewTerminal()
		return fallback(t, err)
	}
}

func printGraph(g *graph.Graph) {
	err := g.OneFrame()
	if err != nil {
		panic(err.Error())
	}
}

func makeGraph(term *terminal.Terminal, debugFollow bool, scale graph.YAxisScale, debugStrict bool, d *data.Data) *graph.Graph {
	g := graph.NewGraph(
		context.Background(),
		graph.GraphConfiguration{
			Terminal:      term,
			DrawingBuffer: draw.NewPaintBuffer(),
			Presentation: graph.Presentation{
				Following:  debugFollow,
				YAxisScale: scale,
			},
			DebugStrict: debugStrict,
			Data:        d,
		},
	)
	return g
}
