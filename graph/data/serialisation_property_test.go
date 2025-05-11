// Use of this source code is governed by a GPL-2 license that can be found in the LICENSE file.
//
// Copyright 2025 Lexer747
//
// SPDX-License-Identifier: GPL-2.0-only

package data_test

import (
	"fmt"
	"net"
	"strconv"
	"testing"
	"time"

	"github.com/Lexer747/acci-ping/graph/data"
	"github.com/Lexer747/acci-ping/ping"
	"pgregory.net/rapid"
)

const (
	maxSliceSize = 1024 << 1
	maxStatsSize = 1024 << 3
)

func TestCompactTimeSpan_Property(t *testing.T) {
	t.Parallel()
	rapid.Check(t, func(t *rapid.T) {
		var (
			begin = rapid.Int64().Draw(t, "begin")
			end   = rapid.Int64().Draw(t, "end")
		)
		testSpan := makeTestTimeSpan(begin, end)
		testCompacter(t, testSpan, &data.TimeSpan{})
	})
}

func TestCompactStats_Property(t *testing.T) {
	t.Parallel()
	rapid.Check(t, func(t *rapid.T) {
		var (
			size         = rapid.IntRange(1, maxStatsSize).Draw(t, "size")
			droppedCount = rapid.IntRange(0, size).Draw(t, "dropped count")
			goodCount    = rapid.IntRange(droppedCount, size).Draw(t, "good count")
			droppedFirst = rapid.Bool().Draw(t, "dropped first")
		)
		testStats := &data.Stats{}

		addGood := func() {
			for i := range goodCount {
				nextPoint := rapid.Int().Draw(t, fmt.Sprintf("point: %d", i))
				testStats.AddPoint(time.Duration(nextPoint) * time.Millisecond)
			}
		}
		addDropped := func() {
			for range droppedCount {
				testStats.AddDroppedPacket()
			}
		}
		if droppedFirst {
			addDropped()
			addGood()
		} else {
			addGood()
			addDropped()
		}
		testCompacter(t, testStats, &data.Stats{})
	})
}

func TestCompactHeader_Property(t *testing.T) {
	t.Parallel()
	rapid.Check(t, func(t *rapid.T) {
		var (
			size         = rapid.IntRange(1, maxStatsSize).Draw(t, "size")
			droppedCount = rapid.IntRange(0, size).Draw(t, "dropped count")
			goodCount    = rapid.IntRange(droppedCount, size).Draw(t, "good count")
			droppedFirst = rapid.Bool().Draw(t, "dropped first")
		)
		testHeader := &data.Header{Stats: &data.Stats{}, TimeSpan: &data.TimeSpan{}}
		addGood := func() {
			for i := range goodCount {
				testHeader.AddPoint(drawPingDataPoint(t, i))
			}
		}
		addDropped := func() {
			for i := range droppedCount {
				testHeader.AddPoint(drawDroppedDataPoint(t, i))
			}
		}
		if droppedFirst {
			addDropped()
			addGood()
		} else {
			addGood()
			addDropped()
		}
		testCompacter(t, testHeader, &data.Header{})
	})
}

func TestCompactNetwork_Property(t *testing.T) {
	t.Parallel()
	rapid.Check(t, func(t *rapid.T) {
		var (
			size = rapid.IntRange(1, maxSliceSize).Draw(t, "size")
		)
		testNetwork := &data.Network{IPs: []net.IP{}}
		for range size {
			testNetwork.AddPoint(drawIP(t))
		}
		testCompacter(t, testNetwork, &data.Network{})
	})
}

func TestCompactBlock_Property(t *testing.T) {
	t.Parallel()
	rapid.Check(t, func(t *rapid.T) {
		var (
			size         = rapid.IntRange(1, maxSliceSize).Draw(t, "size")
			droppedCount = rapid.IntRange(0, size).Draw(t, "dropped count")
			goodCount    = rapid.IntRange(droppedCount, size).Draw(t, "good count")
			droppedFirst = rapid.Bool().Draw(t, "dropped first")
		)
		testBlock := &data.Block{
			Header: &data.Header{Stats: &data.Stats{}, TimeSpan: &data.TimeSpan{}},
			Raw:    []ping.PingDataPoint{},
		}
		addGood := func() {
			for i := range goodCount {
				testBlock.AddPoint(drawPingDataPoint(t, i))
			}
		}
		addDropped := func() {
			for i := range droppedCount {
				testBlock.AddPoint(drawDroppedDataPoint(t, i))
			}
		}
		if droppedFirst {
			addDropped()
			addGood()
		} else {
			addGood()
			addDropped()
		}
		testCompacter(t, testBlock, &data.Block{})
	})
}

func TestCompactDataIndexes_Property(t *testing.T) {
	t.Parallel()
	rapid.Check(t, func(t *rapid.T) {
		var (
			block = rapid.Int().Draw(t, "block")
			raw   = rapid.Int().Draw(t, "raw")
		)
		testDataIndexes := &data.DataIndexes{BlockIndex: block, RawIndex: raw}
		testCompacter(t, testDataIndexes, &data.DataIndexes{})
	})
}

func TestCompactRun_Property(t *testing.T) {
	t.Parallel()
	rapid.Check(t, func(t *rapid.T) {
		var (
			size = rapid.IntRange(1, maxStatsSize).Draw(t, "size")
		)
		testRun := &data.Run{}
		for i := range size {
			if rapid.Bool().Draw(t, strconv.Itoa(i)) {
				testRun.Inc(int64(i))
			} else {
				testRun.Reset()
			}
		}
		testCompacter(t, testRun, &data.Run{})
	})
}

func TestCompactRuns_Property(t *testing.T) {
	t.Parallel()
	drawPingDropped := func(t *rapid.T, label string) ping.Dropped {
		t.Helper()
		if rapid.Bool().Draw(t, label) {
			return ping.NotDropped
		} else {
			return ping.TestDrop
		}
	}

	rapid.Check(t, func(t *rapid.T) {
		var (
			size = rapid.IntRange(1, maxStatsSize).Draw(t, "size")
		)
		testRuns := &data.Runs{GoodPackets: &data.Run{}, DroppedPackets: &data.Run{}}
		for i := range size {
			testRuns.AddPoint(int64(i), ping.PingDataPoint{DropReason: drawPingDropped(t, strconv.Itoa(i))})
		}
		testCompacter(t, testRuns, &data.Runs{})
	})
}

func TestCompactEmptyData_Property(t *testing.T) {
	t.Parallel()
	rapid.Check(t, func(t *rapid.T) {
		var (
			url = rapid.String().Draw(t, "URL")
		)
		testData := data.NewData(url)
		testCompacter(t, testData, &data.Data{})
	})
}

func TestCompactData_Property(t *testing.T) {
	t.Parallel()
	rapid.Check(t, func(t *rapid.T) {
		var (
			url  = rapid.String().Draw(t, "URL")
			size = rapid.IntRange(1, maxSliceSize).Draw(t, "size")
		)
		testData := data.NewData(url)
		for i := range size {
			testData.AddPoint(drawPingResults(t, i))
		}
		testCompacter(t, testData, &data.Data{})
	})
}

func drawPingResults(t *rapid.T, i int) ping.PingResults {
	t.Helper()
	ip := drawIP(t)
	if rapid.Bool().Draw(t, fmt.Sprintf("decision bool: %d", i)) {
		ts := rapid.Int64Range(int64(i)*10, int64(i)*20).Draw(t, fmt.Sprintf("dropped timestamp: %d", i))
		return ping.PingResults{
			Data: ping.PingDataPoint{DropReason: ping.TestDrop, Timestamp: time.UnixMilli(ts)},
			IP:   ip,
		}
	} else {
		d := rapid.Int().Draw(t, fmt.Sprintf("point duration: %d", i))
		ts := rapid.Int64Range(int64(i)*10, int64(i)*20).Draw(t, fmt.Sprintf("point timestamp: %d", i))
		return ping.PingResults{
			Data: ping.PingDataPoint{Duration: time.Duration(d), Timestamp: time.UnixMilli(ts)},
			IP:   ip,
		}
	}
}

func drawDroppedDataPoint(t *rapid.T, i int) ping.PingDataPoint {
	t.Helper()
	ts := rapid.Int64().Draw(t, fmt.Sprintf("dropped timestamp: %d", i))
	point := ping.PingDataPoint{DropReason: ping.TestDrop, Timestamp: time.UnixMilli(ts)}
	return point
}

func drawPingDataPoint(t *rapid.T, i int) ping.PingDataPoint {
	t.Helper()
	d := rapid.Int().Draw(t, fmt.Sprintf("point duration: %d", i))
	ts := rapid.Int64().Draw(t, fmt.Sprintf("point timestamp: %d", i))
	point := ping.PingDataPoint{Duration: time.Duration(d), Timestamp: time.UnixMilli(ts)}
	return point
}

func drawIP(t *rapid.T) net.IP {
	t.Helper()
	var (
		size = rapid.IntRange(4, 16).Draw(t, "IP size")
	)
	ret := make(net.IP, size)
	for i := range ret {
		ret[i] = rapid.Byte().Draw(t, fmt.Sprintf("ip: %d", i))
	}
	return ret
}

func makeTestTimeSpan(begin int64, end int64) *data.TimeSpan {
	testSpan := &data.TimeSpan{
		Begin: time.UnixMilli(begin),
		End:   time.UnixMilli(end),
	}
	testSpan.Duration = testSpan.End.Sub(testSpan.Begin)
	return testSpan
}
