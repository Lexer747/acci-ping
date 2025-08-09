// Use of this source code is governed by a GPL-2 license that can be found in the LICENSE file.
//
// Copyright 2024-2025 Lexer747
//
// SPDX-License-Identifier: GPL-2.0-only

package main

import (
	"flag"
	"fmt"
	"io"
	"os"

	acciping "github.com/Lexer747/acci-ping/cmd/subcommands/acci-ping"
	"github.com/Lexer747/acci-ping/cmd/subcommands/drawframe"
	"github.com/Lexer747/acci-ping/cmd/subcommands/ping"
	"github.com/Lexer747/acci-ping/cmd/subcommands/rawdata"
	"github.com/Lexer747/acci-ping/cmd/subcommands/version"
	"github.com/Lexer747/acci-ping/terminal/ansi"
	"github.com/Lexer747/acci-ping/utils/application"
	"github.com/Lexer747/acci-ping/utils/errors"
	"github.com/Lexer747/acci-ping/utils/exit"
	"github.com/Lexer747/acci-ping/utils/flags"
)

// these looking more bash variables helps clue me into where these
// actually come from which is build time linking, see tools/build.sh.
//
//nolint:staticcheck
var (
	COMMIT     string
	GO_VERSION string
	BRANCH     string
	TIMESTAMP  string
	TAG        string
)

var programName = ansi.Green("acci-ping")

const drawframeString = "drawframe"
const rawdataString = "rawdata"
const pingString = "ping"
const versionString = "version"

type subcommand struct {
	subcommandName string
	description    string
}

var commandsUsage = []subcommand{
	{
		subcommandName: ansi.Red(drawframeString),
		description: programName + " " + ansi.Red(drawframeString) +
			" [file|folder]\n    will draw a single frame of the graph for a given .pings file, or folder of .pings files.",
	},
	{
		subcommandName: ansi.Red(rawdataString),
		description: programName + " " + ansi.Red(rawdataString) +
			" will print the statistics and all raw packets found in a .pings file to stdout.",
	},
	{
		subcommandName: ansi.Red(pingString),
		description: programName + " " + ansi.Red(pingString) +
			" will run like any other ping command line tool and print the plain text packet statistics to stdout.",
	},
	{
		subcommandName: ansi.Red(versionString),
		description: programName + " " + ansi.Red(versionString) +
			" simply prints the version of this program.",
	},
}

var mainDescription = programName + " can run be with no arguments to start the graphing ping accumulator." +
	" To exit simply kill the program via the normal control-c."

func main() {
	info := application.MakeBuildInfo(COMMIT, GO_VERSION, BRANCH, TIMESTAMP, TAG)
	if len(os.Args) > 1 {
		switch os.Args[1] {
		case drawframeString:
			df := drawframe.GetFlags(info)
			FlagParseError(df.Parse(os.Args[2:]))
			PrintHelpDebugIfNeeded(df.HelpDebug(), df.FlagSet)
			drawframe.RunDrawFrame(df)
			exit.Success()
		case rawdataString:
			rd := rawdata.GetFlags()
			FlagParseError(rd.Parse(os.Args[2:]))
			rawdata.RunPrintData(rd)
			exit.Success()
		case pingString:
			p := ping.GetFlags()
			FlagParseError(p.Parse(os.Args[2:]))
			ping.RunPing(p)
			exit.Success()
		case versionString:
			p := version.GetFlags(info)
			FlagParseError(p.Parse(os.Args[2:]))
			version.RunVersion(p)
			exit.Success()
		default:
			// fallthrough
		}
	}
	a := acciping.GetFlags(info)
	a.Usage = func() {
		subCommandUsage(a.Output())
		if a.HelpDebug() {
			flags.PrintFlagsFilter(a.FlagSet, flags.NoFilter())
		} else {
			flags.PrintFlagsFilter(a.FlagSet, flags.ExcludePrefix("debug"))
		}
	}
	FlagParseError(a.Parse(os.Args[1:]))
	PrintHelpDebugIfNeeded(a.HelpDebug(), a.FlagSet)
	acciping.RunAcciPing(a)
	exit.Success()
}

func PrintHelpDebugIfNeeded(b bool, fs *flag.FlagSet) {
	if b {
		fs.Usage()
		exit.Silent()
	}
}

func subCommandUsage(w io.Writer) {
	fmt.Fprint(w, "  "+mainDescription+"\n\n")
	for _, cmd := range commandsUsage {
		fmt.Fprint(w, "  "+cmd.subcommandName+"\n")
		fmt.Fprint(w, "      "+cmd.description+"\n")
	}
	fmt.Fprintf(w, "call any of the above subcommands with --help for extra details on those commands.\n")
	fmt.Fprint(w, "\n"+programName+" arguments:\n")
}

func FlagParseError(err error) {
	if errors.Is(err, flag.ErrHelp) {
		exit.Silent()
	} else {
		exit.OnError(err)
	}
}
