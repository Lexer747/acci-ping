// Use of this source code is governed by a GPL-2 license that can be found in the LICENSE file.
//
// Copyright 2025 Lexer747
//
// SPDX-License-Identifier: GPL-2.0-only

package gui

// NoGUI is a gui implementation which never does any work or asks for any work to be done. Useful for tests.
func NoGUI() GUI {
	return &noGUI{}
}

func (ng *noGUI) GetState() Token          { return ng }
func (ng *noGUI) ShouldDraw() bool         { return false }
func (ng *noGUI) ShouldInvalidate() bool   { return false }
func (ng *noGUI) Drawn(Token)              {}
func (ng *noGUI) Paint(update PaintUpdate) {}

var _ Token = (&noGUI{})
var _ GUI = (&noGUI{})

type noGUI struct {
}
