// Use of this source code is governed by a GPL-2 license that can be found in the LICENSE file.
//
// Copyright 2024-2025 Lexer747
//
// SPDX-License-Identifier: GPL-2.0-only

package data

import (
	"encoding/binary"
	"io"
	"math"
	"time"

	"github.com/Lexer747/acci-ping/utils/errors"
)

// ReadData is the function you want when you have a file or byte stream and wish to de-serialise the result
// into [Data] (use a [bytes.Buffer]). This byte stream should've been encoded with [Data.AsCompact],
// otherwise an error will occur. Note no checksums are in this data format so data-integrity is not
// guaranteed and this may mean that an unlikely file might trick this decoder.
func ReadData(r io.Reader) (*Data, error) {
	toReadFrom, err := io.ReadAll(r)
	if err != nil {
		return nil, errors.Wrap(err, "While reading into Data{}")
	}
	d := &Data{}
	_, err = d.FromCompact(toReadFrom)
	if err != nil {
		return nil, errors.Wrap(err, "While reading into Data{}")
	}
	return d, nil
}

type Compact interface {
	// AsCompact convert a [Compact]ing thing into bytes
	AsCompact(w io.Writer) error

	// FromCompact converts raw bytes back into said thing
	FromCompact(input []byte) (int, error)

	lenCompact
}

// As a rule any new type which is stored to (new!) file must be added here, it's implementation in said file
// named in a comment appended to the declaration. Along with the corresponding version bump and migration
// implementation see [Data.Migrate]. Keep this list of variables sorted.
//
// truly re-usable (within the context of serialising) compacting functions should be here in this file.

var _ Compact = (&Block{})       // block_compact.go
var _ Compact = (&DataIndexes{}) // data_indexes_compact.go
var _ Compact = (&Data{})        // data_compact.go
var _ Compact = (&Header{})      // header_compact.go
var _ Compact = (&Network{})     // network_compact.go
var _ Compact = (&Runs{})        // runs_compact.go
var _ Compact = (&Run{})         // run_compact.go
var _ Compact = (&Stats{})       // stats_compact.go
var _ Compact = (&TimeSpan{})    // timespan_compact.go

// Compacting implementations and list ends.

type Identifier byte

const (
	_ Identifier = 0

	TimeSpanID Identifier = 1
	StatsID    Identifier = 2
	BlockID    Identifier = 3
	HeaderID   Identifier = 4
	DataID     Identifier = 5
	NetworkID  Identifier = 6
	RunsID     Identifier = 7

	_ Identifier = 0xff
)

// phasedWrite is generally used by a compacting implementor to indicate that the data must be written in two
// phases, each phase is of this type. This is useful for types which have dynamic sizes e.g.
// [Network.FromCompact] [Network.AsCompact] which will write all the sizes of it's slices in it's first
// phase, then it's second phase is to write the variable length data. This allows the reader to be more
// simple and efficient as it can read all the sizes before consuming all the bytes.
type phasedWrite = func(ret []byte) int

// Note version"2" here corresponds to the literal 2 of [version], every time a new version is added a
// corresponding function should be created.
func (d *Data) readVersion2(i int, input []byte) (int, error) {
	insertOrderLen := 0
	i += readLen(input[i:], &insertOrderLen)
	i += readInt64(input[i:], &d.TotalCount)
	networkHeaderReader, networkDataReader := d.Network.twoPhaseRead()
	var IPsLen, blockIndexesLen int
	n, err := networkHeaderReader(input[i:], &IPsLen, &blockIndexesLen)
	if err != nil {
		return i, errors.Wrap(err, "while reading compact Data")
	}
	i += n
	// drop block header len, we know it's fixed until new versions are introduced
	i += readInt(input[i:], &n)
	blockLen := 0
	i += readLen(input[i:], &blockLen)
	d.Blocks = make([]*Block, blockLen)
	blockSizes := make([]*int, blockLen)
	blockReads := make([]blockRead, blockLen)
	for index := range blockLen {
		d.Blocks[index] = &Block{}
		blockSizes[index] = new(int)
		header, data := d.Blocks[index].twoPhaseRead()
		n, err := header(input[i:], blockSizes[index])
		if err != nil {
			return i, errors.Wrap(err, "while reading compact Data")
		}
		i += n
		blockReads[index] = data
	}
	URLLen := 0
	i += readLen(input[i:], &URLLen)
	if d.Runs == nil {
		d.Runs = &Runs{}
	}
	n, err = d.Runs.fromCompact(input[i:], d.PingsMeta)
	if err != nil {
		return i, errors.Wrap(err, "while reading compact Data")
	}
	i += n
	n, err = d.Header.FromCompact(input[i:])
	if err != nil {
		return i, errors.Wrap(err, "while reading compact Data")
	}
	i += n

	// Phase 2 read the variable sized data
	d.InsertOrder = make([]DataIndexes, insertOrderLen)
	for index := range d.InsertOrder {
		insert := &d.InsertOrder[index]
		n, err := insert.FromCompact(input[i:])
		if err != nil {
			return i, errors.Wrap(err, "while reading compact Data")
		}
		i += n
	}
	i += networkDataReader(input[i:], IPsLen, blockIndexesLen)
	for index, blockData := range blockReads {
		i += blockData(input[i:], *blockSizes[index])
	}
	i += readString(input[i:], &d.URL, URLLen)
	return i, nil
}

// Note version"1" here corresponds to the literal 1 of [version], every time a new version is added a
// corresponding function should be created.
func (d *Data) readVersion1(i int, input []byte) (int, error) {
	insertOrderLen := 0
	i += readLen(input[i:], &insertOrderLen)
	i += readInt64(input[i:], &d.TotalCount)
	networkHeaderReader, networkDataReader := d.Network.twoPhaseRead()
	var IPsLen, blockIndexesLen int
	n, err := networkHeaderReader(input[i:], &IPsLen, &blockIndexesLen)
	if err != nil {
		return i, errors.Wrap(err, "while reading compact Data")
	}
	i += n
	// drop block header len, we know it's fixed until new versions are introduced
	i += readInt(input[i:], &n)
	blockLen := 0
	i += readLen(input[i:], &blockLen)
	d.Blocks = make([]*Block, blockLen)
	blockSizes := make([]*int, blockLen)
	blockReads := make([]blockRead, blockLen)
	for index := range blockLen {
		d.Blocks[index] = &Block{}
		blockSizes[index] = new(int)
		header, data := d.Blocks[index].twoPhaseRead()
		n, err := header(input[i:], blockSizes[index])
		if err != nil {
			return i, errors.Wrap(err, "while reading compact Data")
		}
		i += n
		blockReads[index] = data
	}
	URLLen := 0
	i += readLen(input[i:], &URLLen)
	n, err = d.Header.FromCompact(input[i:])
	if err != nil {
		return i, errors.Wrap(err, "while reading compact Data")
	}
	i += n

	// Phase 2 read the variable sized data
	d.InsertOrder = make([]DataIndexes, insertOrderLen)
	for index := range d.InsertOrder {
		insert := &d.InsertOrder[index]
		n, err := insert.FromCompact(input[i:])
		if err != nil {
			return i, errors.Wrap(err, "while reading compact Data")
		}
		i += n
	}
	i += networkDataReader(input[i:], IPsLen, blockIndexesLen)
	for index, blockData := range blockReads {
		i += blockData(input[i:], *blockSizes[index])
	}
	i += readString(input[i:], &d.URL, URLLen)
	return i, nil
}

// for internal details we want a few extra methods from all [Compact] things, which provide convenience in
// the [write] function which is better suited for a parent of a child [Compact] compared to
// [Compact.FromCompact].
//
// [byteLen] is used in two key ways:
//   - [Compact.AsCompact] parents will use it to pre-allocate the exact amount of bytes needed to de-serialise
//     the children.
//   - byteLen will be used by parents to compute their own size and potential offsets for variable width
//     data (slices).
type lenCompact interface {
	write(toWriteInto []byte) int
	byteLen() int
}

// Lens in Bytes
const (
	intLen          = int64Len
	int64Len        = 8
	uint64Len       = int64Len
	float64Len      = int64Len
	timeLen         = int64Len
	timeDurationLen = int64Len
	idLen           = 1
	netIPLen        = 16 // Always store in ipv6 form

	timeSpanLen      = idLen + 2*timeLen + timeDurationLen
	statsLen         = idLen + 2*timeDurationLen + 4*float64Len + 2*uint64Len
	headerLen        = idLen + timeSpanLen + statsLen
	pingDataPointLen = timeDurationLen + timeLen + 1
	dataIndexesLen   = intLen + intLen
	runLen           = int64Len + uint64Len + uint64Len
	runsLen          = idLen + runLen + runLen
)

// sliceLenCompact works out the dynamic size for all items in a slice.
func sliceLenCompact[S ~[]T, T lenCompact](slice S) int {
	i := int64Len // 1 int64 to encode the length
	for _, item := range slice {
		i += item.byteLen()
	}
	return i
}

// sliceLenFixed works out the size of a slice where all the items are of fixed size.
func sliceLenFixed[S ~[]T, T any](slice S, itemLen int) int {
	return int64Len + len(slice)*itemLen
}

func stringLen[S ~string](str S) int {
	return int64Len + // 1 int64 to encode the length
		len([]byte(str)) // The number of bytes in the string
}

func writeTime(b []byte, t time.Time) int {
	return writeInt64(b, t.UnixMilli())
}

func readTime(b []byte, t *time.Time) int {
	var i int64
	ret := readInt64(b, &i)
	*t = time.UnixMilli(i)
	return ret
}

func writeDuration(b []byte, d time.Duration) int {
	return writeInt64(b, int64(d))
}

func readDuration(b []byte, d *time.Duration) int {
	var i int64
	ret := readInt64(b, &i)
	*d = time.Duration(i)
	return ret
}

func readID(b []byte, id Identifier) (int, error) {
	if len(b) <= 0 {
		return 0, errors.Errorf("Cannot read id, not enough bytes")
	}
	if id != Identifier(b[0]) {
		return 0, errors.Errorf("Unexpected id %d != %d", b[0], id)
	}
	return 1, nil
}

func writeByte[b ~byte](buf []byte, toWrite b) int {
	buf[0] = byte(toWrite)
	return 1
}

func readByte[b ~byte](buf []byte, toRead *b) int {
	*toRead = b(buf[0])
	return 1
}

func writeStringLen[S ~string](b []byte, str S) int {
	return writeLen(b, []byte(str))
}

func writeString[S ~string](b []byte, str S) int {
	asBytes := []byte(str)
	copy(b, asBytes)
	return len(asBytes)
}

func readString[S ~string](b []byte, s *S, strLen int) int {
	*s = S(string(b[:strLen]))
	return strLen
}

func writeLen[S ~[]T, T any](b []byte, slice S) int {
	binary.LittleEndian.PutUint64(b, uint64(len(slice)))
	return int64Len
}

func readLen(b []byte, i *int) int {
	//nolint:gosec
	// G115 if this overflows it means the underlying file was written with a system supporting 64 bits (as it
	// reached that length of slice), but this current code reading the file is only 32 bits, in which case it
	// won't be able to store the result anyway.
	*i = int(binary.LittleEndian.Uint64(b))
	return int64Len
}

func writeInt64(b []byte, i int64) int {
	//nolint:gosec
	// G115 converting to a uint64 is an overflow but we are simply writing the raw bits to the buffer for later.
	binary.LittleEndian.PutUint64(b, uint64(i))
	return int64Len
}

func readInt64(b []byte, i *int64) int {
	//nolint:gosec
	// G115 converting to a int64 is an overflow but we are simply reading the raw bits to the buffer
	// which started life as a int64.
	*i = int64(binary.LittleEndian.Uint64(b))
	return int64Len
}

func writeInt(b []byte, i int) int {
	//nolint:gosec
	// G115 converting to a uint64 is an overflow but we are simply writing the raw bits to the buffer for later.
	binary.LittleEndian.PutUint64(b, uint64(i))
	return int64Len
}

func readInt(b []byte, i *int) int {
	//nolint:gosec
	// G115 converting to a int64 is an overflow but we are simply reading the raw bits to the buffer
	// which started life as a int.
	*i = int(binary.LittleEndian.Uint64(b))
	return int64Len
}

func writeUint64(b []byte, i uint64) int {
	binary.LittleEndian.PutUint64(b, i)
	return uint64Len
}

func readUint64(b []byte, i *uint64) int {
	*i = binary.LittleEndian.Uint64(b)
	return uint64Len
}

func writeFloat64(b []byte, i float64) int {
	binary.LittleEndian.PutUint64(b, math.Float64bits(i))
	return float64Len
}

func readFloat64(b []byte, i *float64) int {
	*i = math.Float64frombits(binary.LittleEndian.Uint64(b))
	return float64Len
}
