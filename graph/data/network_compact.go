// Use of this source code is governed by a GPL-2 license that can be found in the LICENSE file.
//
// Copyright 2024-2025 Lexer747
//
// SPDX-License-Identifier: GPL-2.0-only

package data

import (
	"io"
	"net"

	"github.com/Lexer747/acci-ping/utils/errors"
)

func (n *Network) AsCompact(w io.Writer) error {
	thisLen := n.byteLen()
	ret := make([]byte, thisLen)
	n.write(ret)
	_, err := w.Write(ret)
	return err
}

func (n *Network) write(ret []byte) int {
	header, data := n.twoPhaseWrite()
	i := header(ret)
	i += data(ret[i:])
	return i
}

func (n *Network) twoPhaseWrite() (phasedWrite, phasedWrite) {
	return func(ret []byte) int {
			i := writeByte(ret, NetworkID)
			i += writeInt(ret[i:], n.curBlockIndex)
			i += writeLen(ret[i:], n.IPs)
			i += writeLen(ret[i:], n.BlockIndexes)
			return i
		}, func(ret []byte) int {
			i := 0
			for _, ip := range n.IPs {
				i += writeIP(ret[i:], ip)
			}
			for _, index := range n.BlockIndexes {
				i += writeInt(ret[i:], index)
			}
			return i
		}
}

func (n *Network) twoPhaseRead() (
	func(input []byte, IPsLen, blockIndexesLen *int) (int, error),
	func(input []byte, IPsLen, blockIndexesLen int) int) {
	return func(input []byte, IPsLen, blockIndexesLen *int) (int, error) {
			i, err := readID(input, NetworkID)
			if err != nil {
				return i, errors.Wrap(err, "while reading compact Network")
			}
			i += readInt(input[i:], &n.curBlockIndex)
			i += readLen(input[i:], IPsLen)
			i += readLen(input[i:], blockIndexesLen)
			return i, nil
		},
		func(input []byte, IPsLen, blockIndexesLen int) int {
			n.IPs = make([]net.IP, IPsLen)
			n.BlockIndexes = make([]int, blockIndexesLen)
			i := 0
			for ip := range n.IPs {
				n.IPs[ip] = make(net.IP, netIPLen)
				i += readIP(input[i:], n.IPs[ip])
			}
			for blockIndex := range n.BlockIndexes {
				i += readInt(input[i:], &n.BlockIndexes[blockIndex])
			}
			return i
		}
}

func (n *Network) byteLen() int {
	return sliceLenFixed(n.IPs, netIPLen) + sliceLenFixed(n.BlockIndexes, intLen) + intLen + idLen
}

func (n *Network) FromCompact(input []byte) (int, error) {
	header, data := n.twoPhaseRead()
	IPsLen := 0
	BlockIndexesLen := 0
	i, err := header(input, &IPsLen, &BlockIndexesLen)
	if err != nil {
		return i, err
	}
	return data(input[i:], IPsLen, BlockIndexesLen), nil
}

func writeIP(b []byte, ip net.IP) int {
	ensure16 := ip.To16()
	copy(b, ensure16)
	return netIPLen
}

func readIP(b []byte, ip net.IP) int {
	copy(ip, b[:netIPLen])
	return netIPLen
}
