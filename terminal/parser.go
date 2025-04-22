// Use of this source code is governed by a GPL-2 license that can be found in the LICENSE file.
//
// Copyright 2025 Lexer747
//
// SPDX-License-Identifier: GPL-2.0-only

package terminal

import (
	"context"
	"io"
	"strconv"
	"strings"

	"github.com/Lexer747/acci-ping/utils/errors"
)

// parser is a helper type for consuming content from a non-EoF based stream. It's API is the public methods
// and the private methods while usable from this package are probably not what you need.
type parser struct {
	// intialise these before usage:

	//nolint:containedctx
	Ctx          context.Context
	ToStreamFrom io.Reader
	Buffer       []byte

	parserHead  int
	pointer     int
	bufferSlice []byte
}

// Consume the exact bytes passed, errors if the stream produced different bytes.
func (p *parser) Consume(bytes []byte) error {
	for _, toConsume := range bytes {
		next, err := p.next()
		if err != nil {
			return err
		}
		if next != toConsume {
			return errors.Errorf("Failed to consume %s, got %s instead", string(next), string(toConsume))
		}
	}
	return nil
}

// ParseHex consume n bytes and then parses the resulting string in base 16 and returns it. Will return an
// error if the parse failed or underlying stream read failed.
func (p *parser) ParseHex(numBytes int) (int, error) {
	var b strings.Builder
	for range numBytes {
		next, err := p.next()
		if err != nil {
			return -1, err
		}
		b.WriteByte(next)
	}
	i, err := strconv.ParseInt(b.String(), 16, 64)
	return int(i), err
}

// next is the streaming API for the internal usage of the parser, it yields exactly one byte at a time and
// cannot be peeked.
func (p *parser) next() (byte, error) {
	if p.pointer <= p.parserHead {
		err := p.waitForMore()
		if err != nil {
			return 0, err
		}
	}
	result := p.Buffer[p.parserHead]
	p.parserHead++
	if result == 0 {
		return p.next()
	}
	return result, nil
}

// since we are treating the underlying reader as "stream" this function does the actual non-stream like
// reading from the [io.Reader] and stores the chunk read into the parser buffer and stores the offset.
func (p *parser) waitForMore() error {
	select {
	case <-p.Ctx.Done():
		return errors.Errorf("closed")
	default:
	}
	if p.bufferSlice == nil {
		p.bufferSlice = p.Buffer[:]
	}
	n, err := p.ToStreamFrom.Read(p.bufferSlice)
	p.bufferSlice = p.bufferSlice[:n]
	if err != nil {
		// We treat EoF as an actual error
		return err
	}
	p.pointer += n
	return nil
}
