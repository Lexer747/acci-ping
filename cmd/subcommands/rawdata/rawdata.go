// Use of this source code is governed by a GPL-2 license that can be found in the LICENSE file.
//
// Copyright 2024-2025 Lexer747
//
// SPDX-License-Identifier: GPL-2.0-only

package rawdata

import (
	"flag"
	"fmt"
	"os"
	"time"

	"github.com/Lexer747/acci-ping/cmd/tab_completion/tabflags"
	"github.com/Lexer747/acci-ping/graph/data"
	"github.com/Lexer747/acci-ping/utils/check"
	"github.com/Lexer747/acci-ping/utils/exit"
)

type Config struct {
	*tabflags.FlagSet

	printAll *bool
	toCSV    *bool
}

func GetFlags() *Config {
	f := flag.NewFlagSet("", flag.ContinueOnError)
	tf := tabflags.NewAutoCompleteFlagSet(f, true, ".pings")
	ret := &Config{
		FlagSet:  tf,
		printAll: tf.Bool("all", false, "prints all raw values otherwise only summarises '.pings' files"),
		toCSV:    tf.Bool("csv", false, "writes '.pings' files as '.csv'"),
	}

	f.Usage = func() {
		w := flag.CommandLine.Output()
		fmt.Fprintf(w, "Usage of %s: reads '.pings' files and outputs the raw data to the stdout\n"+
			"\t rawdata [-all][-csv] FILES\n\n"+
			"e.g. %s rawdata my_ping_capture.ping\n", os.Args[0], os.Args[0])
		f.PrintDefaults()
	}
	return ret
}

func RunPrintData(c *Config) {
	check.Check(c.Parsed(), "flags not parsed")
	flag.Parse()
	toPrint := c.Args()
	if len(toPrint) == 0 {
		fmt.Fprintf(os.Stderr, "No files found, exiting. Use -h/--help to print usage instructions.\n")
		exit.Success()
	}
	for _, file := range toPrint {
		f, err := os.OpenFile(file, os.O_RDONLY, 0)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Failed to open %q, %s\n", file, err.Error())
			continue
		}
		d, err := data.ReadData(f)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Failed to parse %q, %s\n", file, err.Error())
			continue
		}
		defer f.Close()
		handle(*c.printAll, *c.toCSV, d)
	}
}

func handle(printAll, toCSV bool, d *data.Data) {
	// In precedence order of flags
	switch {
	case printAll:
		fmt.Fprintf(os.Stdout, "BEGIN %s: %s\n", d.URL, d.Header.String())
		for i := range d.TotalCount {
			p := d.GetFull(i)
			fmt.Fprintf(os.Stdout, "%d: %s\n", i, p.String())
		}
		fmt.Fprintf(os.Stdout, "END %s: %s\n", d.URL, d.Header.String())
	case toCSV:
		handleCSV(d)
	default:
		fmt.Fprintln(os.Stdout, d.Summary())
	}
}

func handleCSV(d *data.Data) {
	fmt.Fprintln(os.Stdout, "timestamp(RFC3339Nano),latency,dropped,ip,header")
	fmt.Fprintf(os.Stdout, ",,,,%q\n", d.String())
	for i := range d.TotalCount {
		p := d.GetFull(i)
		fmt.Fprintf(
			os.Stdout,
			"%q,%q,%q,%q,\n",
			p.Data.Timestamp.Format(time.RFC3339Nano),
			p.Data.Duration.String(),
			p.Data.DropReason.String(),
			p.IP.String(),
		)
	}
}
