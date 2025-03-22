// Use of this source code is governed by a GPL-2 license that can be found in the LICENSE file.
//
// Copyright 2024-2025 Lexer747
//
// SPDX-License-Identifier: GPL-2.0-only

package terminal

import (
	"io"
	"os"
)

// this file contains the stub wrapper types that allow the main terminal to abstract over if it's running
// against a real file, this looks like a pointless wrapper when we could just be storing a [io.Writer] and
// [io.Reader] in the terminal instead, but there are some operations which still need access to the
// underlying [os.File] for some operations so instead we store these as a concrete union of the two values.

type stdout struct {
	realFile       *os.File
	stubFileWriter io.Writer
}

func (s *stdout) write(b []byte) (int, error) {
	if s.realFile != nil {
		return s.realFile.Write(b)
	} else {
		return s.stubFileWriter.Write(b)
	}
}

type stdin struct {
	realFile       *os.File
	stubFileReader io.Reader
}

func (s *stdin) read(b []byte) (int, error) {
	if s.realFile != nil {
		return s.realFile.Read(b)
	} else {
		return s.stubFileReader.Read(b)
	}
}
