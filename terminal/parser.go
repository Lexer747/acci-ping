// Use of this source code is governed by a GPL-2 license that can be found in the LICENSE file.
//
// Copyright 2025 Lexer747
//
// SPDX-License-Identifier: GPL-2.0-only

package terminal

import (
	"context"
	"io"
	"log/slog"
	"strconv"
	"strings"
	"time"

	"github.com/Lexer747/acci-ping/gui/themes"
	"github.com/Lexer747/acci-ping/terminal/ansi"
	"github.com/Lexer747/acci-ping/utils/errors"
	"golang.org/x/term"
)

// Expected response of the form `\e]11;COLOR\a`, where COLOR == `rgb:RRRR/GGGG/BBBB` and `R`, `G`, and `B`
// are hex digits. (typically scaled up from the internal 24-bit representation), padding is because we use
// this to allocate a correctly sized buffer and we want some extra room incase we get more text than
// expected.
var expectedBGOutput = []byte(ansi.OSC + "11;rgb:RRRR/GGGG/BBBB<-----padding---->")

// tryGetBackgroundColour uses a somewhat standard xterm [ctlseqs] control sequences. To get the back ground
// colour of the terminal. In this case it's collapsed to a single luminance value through the [themes]
// package.
//
// In Chapter [OSC] there is special documentation about the `?` character which can be used as a query
// instead of setting a value.
//
//	If a "?" is given rather than a name or RGB specification,
//	xterm replies with a control sequence of the same form which
//	can be used to set the corresponding dynamic color.
//
// [ctlseqs]: https://invisible-island.net/xterm/ctlseqs/ctlseqs.html
// [OSC]: https://invisible-island.net/xterm/ctlseqs/ctlseqs.html#h3-Operating-System-Commands
func (t *Terminal) tryGetBackgroundColour() bool {
	if t.isTestTerminal {
		return false
	}
	// First put the terminal in raw mode, this is needed so that we can intercept the output of the query
	// from the terminal and it's not just printed unhelpfully for the user.
	oldState, _ := term.MakeRaw(t.stdinFd)
	// Always put the terminal back no matter if we succeeded or failed.
	defer func() {
		_ = term.Restore(t.stdinFd, oldState)
	}()
	// Since read may block forever (a unhelpful terminal for example) therefore we need a sane timeout here.
	// Note that this does leak the go-routine if the terminal does actually never return.
	ctx, cancel := context.WithTimeout(context.Background(), 20*time.Millisecond)
	defer cancel()

	// the parser helper will encapsulate the reads for us, we can't use an [io.ReadAll] because from testing
	// terminals don't return EOF at the end of the query (normally a null byte) so we need to do the reads
	// ourself.
	p := parser{
		Ctx:          ctx,
		Buffer:       make([]byte, len(expectedBGOutput)),
		ToStreamFrom: t.stdout,
	}
	fail := func(err error) {
		slog.Info("Failed to get background colour falling back to dark theme", "underlying reason", err.Error())
	}
	slog.Debug("Reading background colour query from terminal")

	_, err := t.stdin.Write([]byte(ansi.OSC + "11;?\x07"))
	if err != nil {
		fail(err)
		return false
	}

	// Since we are doing the read asynchronously and it may fail we get the results of the parsing from this
	// channel instead of synchronously.
	type rgb struct {
		red, green, blue int
	}
	result := make(chan rgb)

	go func() {
		// TODO also support rgb:RR/GG/BB responses
		// TODO also support rgba:RRRR/GGGG/BBBB/AAAA responses
		// TODO also support rgba:0/0/BBBB/ responses (right now we parse 4 fixed digits)

		// For now it's good enough to hard code only one valid parsing response.
		defer close(result)
		err := p.Consume([]byte("\x1b]11;rgb:"))
		if err != nil {
			fail(err)
			return
		}
		red, err := p.ParseHex(4)
		if err != nil {
			fail(err)
			return
		}
		err = p.Consume([]byte("/"))
		if err != nil {
			fail(err)
			return
		}
		green, err := p.ParseHex(4)
		if err != nil {
			fail(err)
			return
		}
		err = p.Consume([]byte("/"))
		if err != nil {
			fail(err)
			return
		}
		blue, err := p.ParseHex(4)
		if err != nil {
			fail(err)
			return
		}
		slog.Debug("done", "r", red, "g", green, "b", blue)
		result <- rgb{red: red, green: green, blue: blue}
	}()

	// Now finally conclude this saga by waiting on the timeout signal or the valid result from the terminal.
	select {
	case <-ctx.Done():
		fail(errors.Errorf("Timeout"))
	case res := <-result:
		t.backgroundColour = themes.ParseRGB_48bit(res.red, res.green, res.blue)
		return true
	}
	return false
}

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
