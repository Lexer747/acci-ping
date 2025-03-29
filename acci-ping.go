// Use of this source code is governed by a GPL-2 license that can be found in the LICENSE file.
//
// Copyright 2024-2025 Lexer747
//
// SPDX-License-Identifier: GPL-2.0-only

package main

import (
	"flag"
	"fmt"
	"os"

	acciping "github.com/Lexer747/acci-ping/cmd/subcommands/acci-ping"
	"github.com/Lexer747/acci-ping/cmd/subcommands/drawframe"
	"github.com/Lexer747/acci-ping/cmd/subcommands/ping"
	"github.com/Lexer747/acci-ping/cmd/subcommands/rawdata"
	"github.com/Lexer747/acci-ping/graph/terminal/ansi"
	"github.com/Lexer747/acci-ping/utils/errors"
	"github.com/Lexer747/acci-ping/utils/exit"
)

var programName = ansi.Green("acci-ping")

const drawframeString = "drawframe"
const rawdataString = "rawdata"
const pingString = "ping"

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
}

var mainDescription = programName + " can run be with no arguments to start the graphing ping accumulator." +
	" To exit simply kill the program via the normal control-c."

func main() {
	if len(os.Args) > 1 {
		switch os.Args[1] {
		case drawframeString:
			df := drawframe.GetFlags()
			FlagParseError(df.Parse(os.Args[2:]))
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
		default:
			// fallthrough
		}
	}
	a := acciping.GetFlags()
	a.Usage = func() {
		fmt.Fprint(a.Output(), "  "+mainDescription+"\n\n")
		for _, cmd := range commandsUsage {
			fmt.Fprint(a.Output(), "  "+cmd.subcommandName+"\n")
			fmt.Fprint(a.Output(), "      "+cmd.description+"\n")
		}
		fmt.Fprintf(a.Output(), "call any of the above subcommands with --help for extra details on those commands.\n")
		fmt.Fprint(a.Output(), "\n"+programName+" arguments:\n")
		a.PrintDefaults()
	}
	FlagParseError(a.Parse(os.Args[1:]))
	acciping.RunAcciPing(a)
	exit.Success()
}

func FlagParseError(err error) {
	if errors.Is(err, flag.ErrHelp) {
		exit.Silent()
	} else {
		exit.OnError(err)
	}
}
