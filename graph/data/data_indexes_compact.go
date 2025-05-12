// Use of this source code is governed by a GPL-2 license that can be found in the LICENSE file.
//
// Copyright 2024-2025 Lexer747
//
// SPDX-License-Identifier: GPL-2.0-only

package data

import "io"

func (di *DataIndexes) AsCompact(w io.Writer) error {
	ret := make([]byte, di.byteLen())
	_ = di.write(ret)
	_, err := w.Write(ret)
	return err
}

func (di *DataIndexes) FromCompact(input []byte) (int, error) {
	i := readInt(input, &di.BlockIndex)
	i += readInt(input[i:], &di.RawIndex)
	return i, nil
}

func (di *DataIndexes) write(toWriteInto []byte) int {
	i := writeInt(toWriteInto, di.BlockIndex)
	i += writeInt(toWriteInto[i:], di.RawIndex)
	return i
}

func (di *DataIndexes) byteLen() int {
	return dataIndexesLen
}
