// Use of this source code is governed by a GPL-2 license that can be found in the LICENSE file.
//
// Copyright 2024-2025 Lexer747
//
// SPDX-License-Identifier: GPL-2.0-only

package acciping

import (
	"context"
	"flag"
	"strings"

	"github.com/Lexer747/acci-ping/graph"
	"github.com/Lexer747/acci-ping/gui/themes"
	"github.com/Lexer747/acci-ping/terminal"
	"github.com/Lexer747/acci-ping/terminal/ansi"
	"github.com/Lexer747/acci-ping/utils/application"
	"github.com/Lexer747/acci-ping/utils/check"
	"github.com/Lexer747/acci-ping/utils/errors"
	"github.com/Lexer747/acci-ping/utils/exit"
)

type Config struct {
	*application.BuildInfo
	*application.SharedFlags
	*flag.FlagSet

	debugFps           *int
	debuggingTermSize  *string
	filePath           *string
	followingOnStart   *bool
	hideHelpOnStart    *bool
	logarithmicOnStart *bool
	pingBufferingLimit *int
	pingsPerMinute     *float64
	testErrorListener  *bool
	theme              *string
	url                *string
}

func GetFlags(info *application.BuildInfo) *Config {
	f := flag.NewFlagSet("", flag.ContinueOnError)
	sf := application.NewSharedFlags(f)
	ret := &Config{
		BuildInfo:   info,
		SharedFlags: sf,
		FlagSet:     f,

		filePath:           f.String("file", "", "the file to write the pings into. (default data not saved)"),
		hideHelpOnStart:    f.Bool("hide-help", false, "if this flag is used the help box will be hidden by default"),
		pingBufferingLimit: new(int),
		pingsPerMinute: f.Float64("pings-per-minute", 60.0,
			"sets the speed at which the program will try to get new ping results, 0 represents no limit.\n"+
				"Negative values are an error."),
		testErrorListener: f.Bool("debug-error-creator", false,
			"binds the ["+ansi.Blue("e")+"] key to create errors for GUI verification"),
		url: f.String("url", "www.google.com", "the url to target for ping testing"),
		theme: f.String("theme", "", "the colour theme (either a path or builtin theme name) to use for the program,\n"+
			"if empty this will try to get the background colour of the terminal and pick the\n"+
			"built in dark or light theme based on the colour found.\n"+
			"There's also the builtin themes:\n"+strings.Join(themes.DescribeBuiltins(), "\n")+
			"\nSee the docs "+ansi.Blue("https://github.com/Lexer747/acci-ping/blob/main/docs/themes.md")+
			" for how to create custom themes."),
		debuggingTermSize:  f.String("debug-term-size", "", "switches the terminal to fixed mode and no iteractivity"),
		followingOnStart:   f.Bool("follow", false, "if this flag is used the graph will be shown in following mode immediately"),
		debugFps:           f.Int("debug-fps", 240, "configures the internal tickrate for the graph re-paint look (in FPS)"),
		logarithmicOnStart: f.Bool("logarithmic", false, "if this flag is used the graph will be shown in logarithmic mode immediately"),
	}
	*ret.pingBufferingLimit = 10
	return ret
}

func RunAcciPing(c *Config) {
	check.Check(c.Parsed(), "flags not parsed")
	closeLogFile := c.InitLogging(c.BuildInfo)
	defer closeLogFile()
	closeCPUProfile := c.InitCPUProfiling()
	defer closeCPUProfile()

	app := Application{}
	closeMemProfile := c.InitMemProfile()
	defer closeMemProfile()
	ctx, cancelFunc := context.WithCancelCause(context.Background())
	defer cancelFunc(nil)
	ch, d := app.Init(ctx, *c)
	err := app.Run(ctx, cancelFunc, ch, d)
	if err != nil && !errors.Is(err, terminal.UserCancelled) {
		exit.OnError(err)
	} else {
		app.Finish()
	}
}

func (c *Config) YScale() graph.YAxisScale {
	if c.logarithmicOnStart == nil || !*c.logarithmicOnStart {
		return graph.Linear
	}
	return graph.Logarithmic
}
