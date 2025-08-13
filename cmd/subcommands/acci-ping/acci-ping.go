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
	*flag.FlagSet

	cpuprofile         *string
	debugStrict        *bool
	filePath           *string
	hideHelpOnStart    *bool
	helpDebug          *bool
	logFile            *string
	memprofile         *string
	pingBufferingLimit *int
	pingsPerMinute     *float64
	debuggingTermSize  *string
	testErrorListener  *bool
	theme              *string
	url                *string
}

func GetFlags(info *application.BuildInfo) *Config {
	f := flag.NewFlagSet("", flag.ContinueOnError)
	ret := &Config{
		BuildInfo:          info,
		cpuprofile:         f.String("debug-cpuprofile", "", "write cpu profile to `file`"),
		debugStrict:        f.Bool("debug-strict", false, "enables more strict operation in which warnings turn into crashes."),
		filePath:           f.String("file", "", "the file to write the pings into. (default data not saved)"),
		hideHelpOnStart:    f.Bool("hide-help", false, "if this flag is used the help box will be hidden by default"),
		logFile:            f.String("debug-log", "", "write logs to `file`. (default no logs written)"),
		memprofile:         f.String("debug-memprofile", "", "write memory profile to `file`"),
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
		debuggingTermSize: f.String("debug-term-size", "", "switches the terminal to fixed mode and no iteractivity"),
		helpDebug:         f.Bool("help-debug", false, "prints all additional debug arguments"),
		FlagSet:           f,
	}
	*ret.pingBufferingLimit = 10
	return ret
}

func RunAcciPing(c *Config) {
	check.Check(c.Parsed(), "flags not parsed")
	closeLogFile := application.InitLogging(*c.logFile, c.BuildInfo)
	defer closeLogFile()
	closeCPUProfile := application.InitCPUProfiling(*c.cpuprofile)
	defer closeCPUProfile()

	app := Application{}
	closeMemProfile := application.InitMemProfile(*c.memprofile)
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

func (c *Config) HelpDebug() bool {
	return *c.helpDebug
}
