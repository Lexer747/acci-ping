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

func (r *Runs) AsCompact(w io.Writer) error {
	ret := make([]byte, runsLen)
	_ = r.write(ret)
	_, err := w.Write(ret)
	return err
}

func (r *Runs) FromCompact(input []byte) (int, error) {
	return r.fromCompact(input, currentDataVersion)
}
func (r *Runs) fromCompact(input []byte, version version) (int, error) {
	i, err := readID(input, RunsID)
	if err != nil {
		return i, errors.Wrap(err, "while reading compact Runs")
	}
	if r.DroppedPackets == nil {
		r.DroppedPackets = &Run{}
	}
	if r.GoodPackets == nil {
		r.GoodPackets = &Run{}
	}
	n, err := r.GoodPackets.fromCompact(input[i:], version)
	if err != nil {
		return i, errors.Wrap(err, "while reading compact Runs")
	}
	i += n
	n, err = r.DroppedPackets.fromCompact(input[i:], version)
	if err != nil {
		return i, errors.Wrap(err, "while reading compact Runs")
	}
	i += n
	return i, nil
}

func (r *Runs) write(ret []byte) int {
	i := writeByte(ret, RunsID)
	i += r.GoodPackets.write(ret[i:])
	i += r.DroppedPackets.write(ret[i:])
	return i
}

func (r *Runs) byteLen() int {
	return runsLen
}
