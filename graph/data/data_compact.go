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

func (d *Data) AsCompact(w io.Writer) error {
	ret := make([]byte, d.byteLen())
	_ = d.write(ret)
	_, err := w.Write(ret)
	return err
}

// this version of write is the current version and the only one supported. Note not all acci-programs (sub
// commands) actually do this "migration", some will only do a migration in place (in memory) and do not write
// the result back to file. This allows some data to remain very old, the main program itself is repeatably
// writing data back to a file and therefore will always be automatically migrated.

func (d *Data) write(ret []byte) int {
	networkHeader, networkData := d.Network.twoPhaseWrite()
	i := writeByte(ret, DataID)
	// We explicitly do not preserve the version in this data, we have migrated and the write code only ever
	// supports the latest version.
	i += writeByte(ret[i:], currentDataVersion)
	i += writeLen(ret[i:], d.InsertOrder)
	i += writeInt64(ret[i:], d.TotalCount)
	i += networkHeader(ret[i:])
	i += writeInt(ret[i:], blockHeaderLen())
	i += writeLen(ret[i:], d.Blocks)
	deferredData := make([]phasedWrite, len(d.Blocks))
	for blockIndex, block := range d.Blocks {
		header, data := block.twoPhaseWrite()
		deferredData[blockIndex] = data
		i += header(ret[i:])
	}
	i += writeStringLen(ret[i:], d.URL)
	i += d.Runs.write(ret[i:])
	i += d.Header.write(ret[i:])

	// Phase 2 the variable length data
	for _, insert := range d.InsertOrder {
		i += insert.write(ret[i:])
	}
	i += networkData(ret[i:])
	for _, blockData := range deferredData {
		i += blockData(ret[i:])
	}
	i += writeString(ret[i:], d.URL)
	return i
}

// FromCompact, see the top level interface [Compact].
//
// Note: this function does automatically migrate the bytes from one serialization format to the latest. And
// in general the format will be guaranteed (by my promise here) that it will always be forward compatible and
// never backward compatible. As in, a file created with acci-ping v1, can be opened and read by acci-ping v2.
// But a file created by acci-ping v2 **may** not be opened in acci-ping v1 (no promises). Important caveat
// that the main program version doesn't correlate to the data serialisation version.
//
// The version of serialisation can be found at [version] and the constants below.
func (d *Data) FromCompact(input []byte) (int, error) {
	if d.Network == nil {
		d.Network = &Network{}
	}
	if d.Header == nil {
		d.Header = &Header{}
	}
	i, err := readID(input, DataID)
	if err != nil {
		return i, errors.Wrap(err, "while reading compact Data")
	}
	i += readByte(input[i:], &d.PingsMeta)
	switch d.PingsMeta {
	case noRuns:
		n, err := d.readVersion1(i, input)
		if err != nil {
			return i, errors.Wrap(err, "while reading compact Data")
		}
		i += n
		d.migrate()
		return i, nil
	case runsWithNoIndex, currentDataVersion:
		n, err := d.readVersion2(i, input)
		if err != nil {
			return i, errors.Wrap(err, "while reading compact Data")
		}
		i += n
		d.migrate()
		return i, nil
	default:
		panic("exhaustive:enforce")
	}
}

func (d *Data) byteLen() int {
	return idLen + // Identifier
		1 + // Version
		int64Len + // TotalCount
		d.Runs.byteLen() +
		d.Header.byteLen() +
		d.Network.byteLen() +
		intLen + // blockHeaderLen
		// Begin Variable sized items:
		sliceLenCompact(d.Blocks) +
		sliceLenFixed(d.InsertOrder, dataIndexesLen) +
		stringLen(d.URL)
}
