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

func (h *Header) AsCompact(w io.Writer) error {
	ret := make([]byte, h.byteLen())
	_ = h.write(ret)
	_, err := w.Write(ret)
	return err
}

func (h *Header) write(ret []byte) int {
	i := writeByte(ret, HeaderID)
	i += h.Stats.write(ret[i:])
	i += h.TimeSpan.write(ret[i:])
	return i
}

func (h *Header) FromCompact(input []byte) (int, error) {
	i, err := readID(input, HeaderID)
	if err != nil {
		return i, errors.Wrap(err, "while reading compact Header")
	}
	if h.Stats == nil {
		h.Stats = &Stats{}
	}
	n, err := h.Stats.FromCompact(input[i:])
	if err != nil {
		return i, errors.Wrap(err, "while reading compact Header")
	}
	i += n
	if h.TimeSpan == nil {
		h.TimeSpan = &TimeSpan{}
	}
	n, err = h.TimeSpan.FromCompact(input[i:])
	if err != nil {
		return i, errors.Wrap(err, "while reading compact Header")
	}
	i += n
	return i, nil
}

func (h *Header) byteLen() int {
	return headerLen
}
