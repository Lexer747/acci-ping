// Use of this source code is governed by a GPL-2 license that can be found in the LICENSE file.
//
// Copyright 2024-2025 Lexer747
//
// SPDX-License-Identifier: GPL-2.0-only

package acciping

import (
	"github.com/Lexer747/acci-ping/gui"
	"github.com/Lexer747/acci-ping/terminal"
)

type GUI struct {
	listeningChars map[rune]terminal.ConditionalListener
	GUIState       *gui.GUIState
	fallbacks      []terminal.Listener
}

func newGUIState() *GUI {
	return &GUI{
		listeningChars: map[rune]terminal.ConditionalListener{},
		fallbacks:      []terminal.Listener{},
		GUIState:       gui.NewGUIState(),
	}
}
