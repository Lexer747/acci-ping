// Use of this source code is governed by a GPL-2 license that can be found in the LICENSE file.
//
// Copyright 2024-2025 Lexer747
//
// SPDX-License-Identifier: GPL-2.0-only

package bytes

import (
	"bytes"
	"fmt"
	"io"
	"strings"
	"sync"
)

// Clear will zero all bytes upto [n]. Helpful when you want to reset a buffer after a [io.Writer] call
// succeeds.
func Clear(buffer []byte, n int) {
	for i := range n {
		buffer[i] = 0
	}
}

// HexPrint will print a buffer as hexadecimal values.
func HexPrint(buffer []byte) string {
	var b strings.Builder
	b.WriteString("[")
	for i, bite := range buffer {
		fmt.Fprintf(&b, "0x%x", bite)
		if i < len(buffer)-1 {
			b.WriteString(", ")
		}
	}
	b.WriteString("]")
	return b.String()
}

// SafeBuffer is a wrapped [bytes.Buffer] with a read-write-mutex, this guards the underlying mutex from
// invalid concurrent access making it thread safe. The zero value is not safe use [NewSafeBuffer] instead.
//
// This is not safe in regards to out of bounds writes and running out of memory or memory safety in general.
type SafeBuffer struct {
	b *bytes.Buffer
	m *sync.RWMutex
}

func NewSafeBuffer() *SafeBuffer {
	return &SafeBuffer{
		b: &bytes.Buffer{},
		m: &sync.RWMutex{},
	}
}
func NewExistingSafeBuffer(buf []byte) *SafeBuffer {
	return &SafeBuffer{
		b: bytes.NewBuffer(buf),
		m: &sync.RWMutex{},
	}
}
func NewSafeBufferString(s string) *SafeBuffer {
	return &SafeBuffer{
		b: bytes.NewBufferString(s),
		m: &sync.RWMutex{},
	}
}

// Bytes returns a slice of length b.Len() holding the unread portion of the buffer. The slice is valid for
// use only until the next buffer modification (that is, only until the next call to a method like
// [Buffer.Read], [Buffer.Write], [Buffer.Reset], or [Buffer.Truncate]). The slice aliases the buffer content
// at least until the next buffer modification, so immediate changes to the slice will affect the result of
// future reads.
//
// Holds the ReadLock.
func (b *SafeBuffer) Bytes() []byte {
	b.m.RLock()
	defer b.m.RUnlock()
	return b.b.Bytes()
}

// AvailableBuffer returns an empty buffer with b.Available() capacity.
// This buffer is intended to be appended to and
// passed to an immediately succeeding [Buffer.Write] call.
// The buffer is only valid until the next write operation on b.
//
// Holds the ReadLock.
func (b *SafeBuffer) AvailableBuffer() []byte {
	b.m.RLock()
	defer b.m.RUnlock()
	return b.b.AvailableBuffer()
}

// String returns the contents of the unread portion of the buffer
// as a string. If the [Buffer] is a nil pointer, it returns "<nil>".
//
// To build strings more efficiently, see the [strings.Builder] type.
//
// Holds the ReadLock.
func (b *SafeBuffer) String() string {
	b.m.RLock()
	defer b.m.RUnlock()
	return b.b.String()
}

// Len returns the number of bytes of the unread portion of the buffer;
// b.Len() == len(b.Bytes()).
//
// Holds the ReadLock.
func (b *SafeBuffer) Len() int {
	b.m.RLock()
	defer b.m.RUnlock()
	return b.b.Len()
}

// Cap returns the capacity of the buffer's underlying byte slice, that is, the
// total space allocated for the buffer's data.
//
// Holds the ReadLock.
func (b *SafeBuffer) Cap() int {
	b.m.RLock()
	defer b.m.RUnlock()
	return b.b.Cap()
}

// Available returns how many bytes are unused in the buffer.
//
// Holds the ReadLock.
func (b *SafeBuffer) Available() int {
	b.m.RLock()
	defer b.m.RUnlock()
	return b.b.Available()
}

// Truncate discards all but the first n unread bytes from the buffer
// but continues to use the same allocated storage.
// It panics if n is negative or greater than the length of the buffer.
//
// Holds the WriteLock.
func (b *SafeBuffer) Truncate(n int) {
	b.m.Lock()
	defer b.m.Unlock()
	b.b.Truncate(n)
}

// Reset resets the buffer to be empty,
// but it retains the underlying storage for use by future writes.
// Reset is the same as [Buffer.Truncate](0).
//
// Holds the WriteLock.
func (b *SafeBuffer) Reset() {
	b.m.Lock()
	defer b.m.Unlock()
	b.b.Reset()
}

// Grow grows the buffer's capacity, if necessary, to guarantee space for
// another n bytes. After Grow(n), at least n bytes can be written to the
// buffer without another allocation.
// If n is negative, Grow will panic.
// If the buffer can't grow it will panic with [ErrTooLarge].
//
// Holds the WriteLock.
func (b *SafeBuffer) Grow(n int) {
	b.m.Lock()
	defer b.m.Unlock()
	b.b.Grow(n)
}

// Write appends the contents of p to the buffer, growing the buffer as
// needed. The return value n is the length of p; err is always nil. If the
// buffer becomes too large, Write will panic with [ErrTooLarge].
//
// Holds the WriteLock.
func (b *SafeBuffer) Write(p []byte) (n int, err error) {
	b.m.Lock()
	defer b.m.Unlock()
	return b.b.Write(p)
}

// WriteString appends the contents of s to the buffer, growing the buffer as
// needed. The return value n is the length of s; err is always nil. If the
// buffer becomes too large, WriteString will panic with [ErrTooLarge].
//
// Holds the WriteLock.
func (b *SafeBuffer) WriteString(s string) (n int) {
	b.m.Lock()
	defer b.m.Unlock()
	n, _ = b.b.WriteString(s)
	return n
}

// ReadFrom reads data from r until EOF and appends it to the buffer, growing
// the buffer as needed. The return value n is the number of bytes read. Any
// error except io.EOF encountered during the read is also returned. If the
// buffer becomes too large, ReadFrom will panic with [ErrTooLarge].
//
// Holds the ReadLock.
func (b *SafeBuffer) ReadFrom(r io.Reader) (n int64, err error) {
	b.m.RLock()
	defer b.m.RUnlock()
	return b.b.ReadFrom(r)
}

// WriteTo writes data to w until the buffer is drained or an error occurs.
// The return value n is the number of bytes written; it always fits into an
// int, but it is int64 to match the [io.WriterTo] interface. Any error
// encountered during the write is also returned.
//
// Holds the WriteLock.
func (b *SafeBuffer) WriteTo(w io.Writer) (n int64, err error) {
	b.m.Lock()
	defer b.m.Unlock()
	return b.b.WriteTo(w)
}

// WriteByte appends the byte c to the buffer, growing the buffer as needed.
// The returned error is always nil, but is included to match [bufio.Writer]'s
// WriteByte. If the buffer becomes too large, WriteByte will panic with
// [ErrTooLarge].
//
// Holds the WriteLock.
func (b *SafeBuffer) WriteByte(c byte) error {
	b.m.Lock()
	defer b.m.Unlock()
	return b.b.WriteByte(c)
}

// WriteRune appends the UTF-8 encoding of Unicode code point r to the
// buffer, returning its length and an error, which is always nil but is
// included to match [bufio.Writer]'s WriteRune. The buffer is grown as needed;
// if it becomes too large, WriteRune will panic with [ErrTooLarge].
//
// Holds the WriteLock.
func (b *SafeBuffer) WriteRune(r rune) (n int, err error) {
	b.m.Lock()
	defer b.m.Unlock()
	return b.b.WriteRune(r)
}

// Read reads the next len(p) bytes from the buffer or until the buffer
// is drained. The return value n is the number of bytes read. If the
// buffer has no data to return, err is [io.EOF] (unless len(p) is zero);
// otherwise it is nil.
//
// Holds the ReadLock.
func (b *SafeBuffer) Read(p []byte) (n int, err error) {
	b.m.RLock()
	defer b.m.RUnlock()
	return b.b.Read(p)
}

// Next returns a slice containing the next n bytes from the buffer,
// advancing the buffer as if the bytes had been returned by [Buffer.Read].
// If there are fewer than n bytes in the buffer, Next returns the entire buffer.
// The slice is only valid until the next call to a read or write method.
//
// Holds the ReadLock.
func (b *SafeBuffer) Next(n int) []byte {
	b.m.RLock()
	defer b.m.RUnlock()
	return b.b.Next(n)
}

// ReadByte reads and returns the next byte from the buffer.
// If no byte is available, it returns error [io.EOF].
//
// Holds the ReadLock.
func (b *SafeBuffer) ReadByte() (byte, error) {
	b.m.Lock()
	defer b.m.Unlock()
	return b.b.ReadByte()
}

// ReadRune reads and returns the next UTF-8-encoded
// Unicode code point from the buffer.
// If no bytes are available, the error returned is io.EOF.
// If the bytes are an erroneous UTF-8 encoding, it
// consumes one byte and returns U+FFFD, 1.
//
// Holds the ReadLock.
func (b *SafeBuffer) ReadRune() (r rune, size int, err error) {
	b.m.RLock()
	defer b.m.RUnlock()
	return b.b.ReadRune()
}

// UnreadRune unreads the last rune returned by [Buffer.ReadRune].
// If the most recent read or write operation on the buffer was
// not a successful [Buffer.ReadRune], UnreadRune returns an error.  (In this regard
// it is stricter than [Buffer.UnreadByte], which will unread the last byte
// from any read operation.)
//
// Holds the WriteLock.
func (b *SafeBuffer) UnreadRune() error {
	b.m.Lock()
	defer b.m.Unlock()
	return b.b.UnreadRune()
}

// UnreadByte unreads the last byte returned by the most recent successful read operation that read at least
// one byte. If a write has happened since the last read, if the last read returned an error, or if the 'read'
// read zero bytes, UnreadByte returns an error.
//
// Holds the WriteLock.
func (b *SafeBuffer) UnreadByte() error {
	b.m.Lock()
	defer b.m.Unlock()
	return b.b.UnreadByte()
}

// ReadBytes reads until the first occurrence of delim in the input,
// returning a slice containing the data up to and including the delimiter.
// If ReadBytes encounters an error before finding a delimiter,
// it returns the data read before the error and the error itself (often [io.EOF]).
// ReadBytes returns err != nil if and only if the returned data does not end in
// delim.
//
// Holds the ReadLock.
func (b *SafeBuffer) ReadBytes(delim byte) (line []byte, err error) {
	b.m.RLock()
	defer b.m.RUnlock()
	return b.b.ReadBytes(delim)
}

// ReadString reads until the first occurrence of delim in the input,
// returning a string containing the data up to and including the delimiter.
// If ReadString encounters an error before finding a delimiter,
// it returns the data read before the error and the error itself (often [io.EOF]).
// ReadString returns err != nil if and only if the returned data does not end
// in delim.
//
// Holds the ReadLock.
func (b *SafeBuffer) ReadString(delim byte) (line string, err error) {
	b.m.RLock()
	defer b.m.RUnlock()
	return b.b.ReadString(delim)
}
