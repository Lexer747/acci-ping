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

func (s *Stats) AsCompact(w io.Writer) error {
	ret := make([]byte, statsLen)
	_ = s.write(ret)
	_, err := w.Write(ret)
	return err
}

func (s *Stats) write(ret []byte) int {
	i := writeByte(ret, StatsID)
	i += writeDuration(ret[i:], s.Min)
	i += writeDuration(ret[i:], s.Max)
	i += writeFloat64(ret[i:], s.Mean)
	i += writeUint64(ret[i:], s.GoodCount)
	i += writeFloat64(ret[i:], s.Variance)
	i += writeFloat64(ret[i:], s.StandardDeviation)
	i += writeUint64(ret[i:], s.PacketsDropped)
	i += writeFloat64(ret[i:], s.sumOfSquares)
	return i
}

func (s *Stats) FromCompact(input []byte) (int, error) {
	i, err := readID(input, StatsID)
	if err != nil {
		return i, errors.Wrap(err, "while reading compact Stats")
	}
	i += readDuration(input[i:], &s.Min)
	i += readDuration(input[i:], &s.Max)
	i += readFloat64(input[i:], &s.Mean)
	i += readUint64(input[i:], &s.GoodCount)
	i += readFloat64(input[i:], &s.Variance)
	i += readFloat64(input[i:], &s.StandardDeviation)
	i += readUint64(input[i:], &s.PacketsDropped)
	i += readFloat64(input[i:], &s.sumOfSquares)
	return i, nil
}

func (s *Stats) byteLen() int {
	return statsLen
}
