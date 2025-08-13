// Use of this source code is governed by a GPL-2 license that can be found in the LICENSE file.
//
// Copyright 2025 Lexer747
//
// SPDX-License-Identifier: GPL-2.0-only

package th

import (
	"fmt"
	"strings"
)

type debug struct {
	data []debugItem
	cur  debugItem
}

type debugItem struct {
	sequence   []rune
	command    string
	startIndex int
	endIndex   int
}

func (d debug) String() string {
	b := &strings.Builder{}
	b.WriteString("[")
	for i, item := range d.data {
		b.WriteString(item.String())
		if i < len(d.data)-1 {
			b.WriteString(", ")
		}
	}
	b.WriteString("\nCur: " + d.cur.String() + "]")
	return b.String()
}
func (d debugItem) String() string {
	if d.command == "" {
		return fmt.Sprintf("{sequence:[%s] startIndex:%d endIndex:%d}",
			string(d.sequence), d.startIndex, d.endIndex,
		)
	}
	return fmt.Sprintf("{command:%q sequence:[%s] startIndex:%d endIndex:%d}",
		d.command, string(d.sequence), d.startIndex, d.endIndex,
	)
}

func (d debug) reset(a *ansiState, consumed rune) {
	a.debug = debug{
		data: append(d.data, d.cur),
		cur: debugItem{
			sequence:   []rune{consumed},
			command:    "",
			startIndex: a.head,
			endIndex:   a.head,
		},
	}
}
func (d debug) consume(a *ansiState, consumed rune) {
	a.debug = debug{
		data: d.data,
		cur: debugItem{
			sequence:   append(a.debug.cur.sequence, consumed),
			startIndex: a.debug.cur.startIndex,
			command:    a.debug.cur.command,
			endIndex:   a.head,
		},
	}
}
func (d debug) setCommand(a *ansiState, cmd string, args ...any) {
	a.debug = debug{
		data: d.data,
		cur: debugItem{
			sequence:   a.debug.cur.sequence,
			command:    fmt.Sprintf(cmd, args...),
			startIndex: a.debug.cur.startIndex,
			endIndex:   a.head,
		},
	}
}
