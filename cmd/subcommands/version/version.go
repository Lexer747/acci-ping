// Use of this source code is governed by a GPL-2 license that can be found in the LICENSE file.
//
// Copyright 2025 Lexer747
//
// SPDX-License-Identifier: GPL-2.0-only

package version

import (
	"flag"
	"fmt"
	"strings"

	"github.com/Lexer747/acci-ping/terminal/ansi"
	"github.com/Lexer747/acci-ping/utils/application"
)

type Config struct {
	*application.BuildInfo
	*flag.FlagSet
}

func GetFlags(info *application.BuildInfo) *Config {
	f := flag.NewFlagSet("", flag.ContinueOnError)
	ret := &Config{
		BuildInfo: info,
		FlagSet:   f,
	}
	return ret
}

func RunVersion(c *Config) {
	versionColour := ansi.Cyan
	detailsColour := ansi.Gray
	const header = "acci-ping version: %s\n"
	if c.BuildInfo == nil {
		fmt.Printf(
			header,
			versionColour("local build"),
		)
	} else {
		var b strings.Builder
		const details = "Details - Commit:%s Branch:%q GoVersion:%q BuildTimestamp:%s\n"
		fmt.Fprintf(&b, header, versionColour(c.Tag()))
		b.WriteString(detailsColour(
			fmt.Sprintf(details,
				c.Commit(),
				c.Branch(),
				c.GoVersion(),
				c.BuildTimestamp(),
			),
		))
		fmt.Print(b.String())
	}
}
