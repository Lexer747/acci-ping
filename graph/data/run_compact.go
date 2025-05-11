// Use of this source code is governed by a GPL-2 license that can be found in the LICENSE file.
//
// Copyright 2024-2025 Lexer747
//
// SPDX-License-Identifier: GPL-2.0-only

package data

import "io"

func (r *Run) AsCompact(w io.Writer) error {
	ret := make([]byte, runLen)
	_ = r.write(ret)
	_, err := w.Write(ret)
	return err
}

func (r *Run) fromCompact(input []byte, version version) (int, error) {
	switch version {
	case noRuns:
		panic("should not be called")
	case runsWithNoIndex:
		i := readUint64(input, &r.Longest)
		i += readUint64(input[i:], &r.Current)
		return i, nil
	case currentDataVersion:
		i := readInt64(input, &r.LongestIndexEnd)
		i += readUint64(input[i:], &r.Longest)
		i += readUint64(input[i:], &r.Current)
		return i, nil
	}
	panic("exhaustive:enforce")
}

func (r *Run) FromCompact(input []byte) (int, error) {
	return r.fromCompact(input, currentDataVersion)
}

func (r *Run) write(ret []byte) int {
	i := writeInt64(ret, r.LongestIndexEnd)
	i += writeUint64(ret[i:], r.Longest)
	i += writeUint64(ret[i:], r.Current)
	return i
}

func (r *Run) byteLen() int {
	return runLen
}
