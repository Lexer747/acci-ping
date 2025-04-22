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

	"github.com/Lexer747/acci-ping/draw"
	"github.com/Lexer747/acci-ping/files"
	"github.com/Lexer747/acci-ping/graph"
	"github.com/Lexer747/acci-ping/graph/data"
	"github.com/Lexer747/acci-ping/gui/themes"
	"github.com/Lexer747/acci-ping/terminal"
	"github.com/Lexer747/acci-ping/terminal/ansi"
	"github.com/Lexer747/acci-ping/utils/application"
	"github.com/Lexer747/acci-ping/utils/check"
	"github.com/Lexer747/acci-ping/utils/exit"
)

type Config struct {
	cpuprofile  *string
	debugStrict *bool
	debugFollow *bool
	logFile     *string
	memprofile  *string
	termSize    *string
	theme       *string

	*application.BuildInfo
	*flag.FlagSet
}

func GetFlags(info *application.BuildInfo) *Config {
	f := flag.NewFlagSet("", flag.ContinueOnError)
	ret := &Config{
		BuildInfo:   info,
		cpuprofile:  f.String("cpuprofile", "", "write cpu profile to `file`"),
		debugStrict: f.Bool("debug-strict", false, "enables more strict operation in which warnings turn into crashes."),
		debugFollow: f.Bool("debug-follow", false, "switches drawing to followLastSpan."),
		logFile:     f.String("l", "", "write logs to `file`. (default no logs written)"),
		memprofile:  f.String("memprofile", "", "write memory profile to `file`"),
		termSize: f.String("term-size", "", "controls the terminal size and fixes it to the input,"+
			" input is in the form \"<H>x<W>\" e.g. 20x80. H and W must be integers - where H == height, and W == width of the terminal."),
		theme: f.String("theme", "", "the colour theme (either a path or builtin theme name) to use for the program,\n"+
			"if empty this will try to get the background colour of the terminal and pick the\n"+
			"built in dark or light theme based on the colour found.\n"+
			"There's also the builtin list of themes:\n"+strings.Join(themes.DescribeBuiltins(), "\n")+
			"\nSee the docs "+ansi.Blue("https://github.com/Lexer747/acci-ping/blob/main/docs/themes.md")+
			" for how to create custom themes."),
		FlagSet: f,
	}
	f.Usage = func() {
		w := flag.CommandLine.Output()
		fmt.Fprintf(w, "Usage of %s: reads '.pings' files and outputs the final frame of the capture\n"+
			"\t drawframe [options] FILE\n\n"+
			"e.g. %s my_ping_capture.ping\n", os.Args[0], os.Args[0])
		flag.PrintDefaults()
	}
	return ret
}

func RunDrawFrame(c *Config) {
	check.Check(c.Parsed(), "flags not parsed")
	closeLogFile := application.InitLogging(*c.logFile, c.BuildInfo)
	defer closeLogFile()
	closeCPUProfile := application.InitCPUProfiling(*c.cpuprofile)
	defer closeCPUProfile()
	closeMemProfile := application.InitMemProfile(*c.memprofile)
	defer closeMemProfile()
	profiling := *c.cpuprofile != "" || *c.memprofile != ""

	toPrint := c.Args()
	if len(toPrint) == 0 {
		fmt.Fprint(os.Stderr, "No files found, exiting. Use -h/--help to print usage instructions.\n")
		exit.Success()
	}

	term, err := makeTerminal(c.termSize)
	exit.OnErrorMsg(err, "failed to open terminal to draw")

	err = application.LoadTheme(*c.theme, term)
	exit.OnErrorMsg(err, "failed to use theme")
	graph.StartUp()

	for _, path := range toPrint {
		run(term, path, profiling, *c.debugFollow, *c.debugStrict)
	}
	fmt.Println()
	fmt.Println()
	fmt.Println()
}

func run(term *terminal.Terminal, path string, profiling, debugFollow, debugStrict bool) {
	fs, err := os.Stat(path)
	exit.OnErrorMsgf(err, "Couldn't stat path %q, failed with", path)
	if fs.IsDir() {
		err := filepath.WalkDir(path, func(p string, d os.DirEntry, err error) error {
			if filepath.Ext(p) != ".pings" {
				return nil
			}
			do(p, term, profiling, debugFollow, debugStrict)
			return nil
		})
		exit.OnErrorMsgf(err, "Couldn't walk path %q, failed with", path)
	} else {
		do(path, term, profiling, debugFollow, debugStrict)
	}
}

func do(path string, term *terminal.Terminal, profiling, debugFollow, debugStrict bool) {
	d, f, err := files.LoadFile(path)
	exit.OnErrorMsg(err, "Couldn't open and read file, failed with")
	f.Close()
	if err != nil {
		panic(err.Error())
	}

	// TODO dont profile like this when iterating over a folder of inputs.
	if profiling {
		timer := time.NewTimer(time.Second * 60)
		running := true
		for running {
			printGraph(term, d, debugFollow, debugStrict)
			select {
			case <-timer.C:
				running = false
			default:
			}
		}
	} else {
		printGraph(term, d, debugFollow, debugStrict)
	}
}

func makeTerminal(termSize *string) (*terminal.Terminal, error) {
	if termSize != nil && *termSize != "" {
		return terminal.NewParsedFixedSizeTerminal(*termSize)
	} else {
		return terminal.NewTerminal()
	}
}

func printGraph(term *terminal.Terminal, d *data.Data, debugFollow, debugStrict bool) {
	g := graph.NewGraph(
		context.Background(),
		graph.GraphConfiguration{
			Terminal:       term,
			PingsPerMinute: 0,
			DrawingBuffer:  draw.NewPaintBuffer(),
			InitialControl: graph.Control{
				FollowLatestSpan: debugFollow,
			},
			DebugStrict: debugStrict,
			Data:        d,
		},
	)
	fmt.Println()
	err := g.OneFrame()
	if err != nil {
		panic(err.Error())
	}
}
