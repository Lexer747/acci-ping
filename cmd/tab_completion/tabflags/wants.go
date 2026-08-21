// Use of this source code is governed by a GPL-2 license that can be found in the LICENSE file.
//
// Copyright 2026 Lexer747
//
// SPDX-License-Identifier: GPL-2.0-only

package tabflags

type Wants byte

const (
	Nothing Wants = 0b00000000
	File    Wants = 0b00000001
	Folder  Wants = 0b00000010
)

func (b Wants) WantsFolderOrFile() bool {
	return b.IsSet(File) || b.IsSet(Folder)
}

func (b Wants) IsSet(desc Wants) bool {
	return b&desc == desc
}
