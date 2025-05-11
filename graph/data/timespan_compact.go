// Use of this source code is governed by a GPL-2 license that can be found in the LICENSE file.
//
// Copyright 2024-2025 Lexer747
//
// SPDX-License-Identifier: GPL-2.0-only

package data

import (
	"io"

	"github.com/Lexer747/acci-ping/utils/errors"
)

func (ts *TimeSpan) AsCompact(w io.Writer) error {
	ret := make([]byte, timeSpanLen)
	_ = ts.write(ret)
	_, err := w.Write(ret)
	return err
}

func (ts *TimeSpan) write(ret []byte) int {
	i := writeByte(ret, TimeSpanID)
	i += writeTime(ret[i:], ts.Begin)
	i += writeTime(ret[i:], ts.End)
	i += writeDuration(ret[i:], ts.Duration)
	return i
}

func (ts *TimeSpan) FromCompact(input []byte) (int, error) {
	i, err := readID(input, TimeSpanID)
	if err != nil {
		return i, errors.Wrap(err, "while reading compact TimeSpan")
	}
	i += readTime(input[i:], &ts.Begin)
	i += readTime(input[i:], &ts.End)
	i += readDuration(input[i:], &ts.Duration)
	return i, nil
}

func (ts *TimeSpan) byteLen() int {
	return timeSpanLen
}
