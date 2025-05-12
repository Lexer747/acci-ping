// Use of this source code is governed by a GPL-2 license that can be found in the LICENSE file.
//
// Copyright 2024-2025 Lexer747
//
// SPDX-License-Identifier: GPL-2.0-only

package data

import (
	"io"

	"github.com/Lexer747/acci-ping/ping"
	"github.com/Lexer747/acci-ping/utils/errors"
)

func (b *Block) AsCompact(w io.Writer) error {
	ret := make([]byte, b.byteLen())
	_ = b.write(ret)
	_, err := w.Write(ret)
	return err
}

func (b *Block) FromCompact(input []byte) (int, error) {
	header, data := b.twoPhaseRead()
	rawLen := 0
	i, err := header(input, &rawLen)
	if err != nil {
		return i, err
	}
	return data(input[i:], rawLen), nil
}

func (b *Block) write(ret []byte) int {
	header, data := b.twoPhaseWrite()
	i := header(ret)
	i += data(ret[i:])
	return i
}

func (b *Block) twoPhaseWrite() (phasedWrite, phasedWrite) {
	return func(ret []byte) int {
			i := writeByte(ret, BlockID)
			i += writeLen(ret[i:], b.Raw)
			i += b.Header.write(ret[i:])
			return i
		}, func(ret []byte) int {
			i := 0
			for _, raw := range b.Raw {
				i += writePingDataPoint(ret[i:], raw)
			}
			return i
		}
}

type blockRead = func(input []byte, rawLen int) int

func (b *Block) twoPhaseRead() (
	func(input []byte, rawLen *int) (int, error),
	blockRead,
) {
	if b.Header == nil {
		b.Header = &Header{}
	}
	return func(input []byte, blockLen *int) (int, error) {
			i, err := readID(input, BlockID)
			if err != nil {
				return i, errors.Wrap(err, "while reading compact Block")
			}
			i += readLen(input[i:], blockLen)
			n, err := b.Header.FromCompact(input[i:])
			if err != nil {
				return i, errors.Wrap(err, "while reading compact Block")
			}
			return i + n, err
		},
		func(input []byte, rawLen int) int {
			b.Raw = make([]ping.PingDataPoint, rawLen)
			i := 0
			for rawIndex := range b.Raw {
				i += readPingDataPoint(input[i:], &b.Raw[rawIndex])
			}
			return i
		}
}

func (b *Block) byteLen() int {
	return idLen + headerLen + sliceLenFixed(b.Raw, pingDataPointLen)
}

func blockHeaderLen() int {
	return idLen + headerLen + sliceLenFixed([]byte{}, 0)
}

func writePingDataPoint(b []byte, p ping.PingDataPoint) int {
	i := writeDuration(b, p.Duration)
	i += writeTime(b[i:], p.Timestamp)
	i += writeByte(b[i:], p.DropReason)
	return i
}

func readPingDataPoint(b []byte, p *ping.PingDataPoint) int {
	i := readDuration(b, &p.Duration)
	i += readTime(b[i:], &p.Timestamp)
	i += readByte(b[i:], &p.DropReason)
	return i
}
