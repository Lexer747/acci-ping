// Use of this source code is governed by a GPL-2 license that can be found in the LICENSE file.
//
// Copyright 2024-2025 Lexer747
//
// SPDX-License-Identifier: GPL-2.0-only

package ping

import (
	"context"
	"crypto/rand"
	"encoding/binary"
	"fmt"
	"math"
	"net"
	"sync"
	"time"

	"github.com/Lexer747/acci-ping/utils/errors"

	"golang.org/x/net/icmp"
)

type Ping struct {
	connect       *icmp.PacketConn
	addrType      addressType
	id            uint16
	currentURL    string
	timeout       time.Duration
	ratelimitTime time.Duration

	echoType, echoReply icmp.Type

	addresses *queryCache
}

func (p *Ping) LastIP() string {
	return p.addresses.GetLastIP()
}

// NewPing constructs a new Ping client which can perform accurate ping measurements. Either with
// [Ping.OneShot] or [Ping.CreateChannel], [Ping.OneShot] is best if a long lived client is undesired and will
// simply block while a single ping is sent to the given URL. [Ping.CreateChannel] creates a go channel which
// can be switched on to get consistent packets from the given URL.
//
// The [Ping.CreateChannel] handles things like:
//   - The URL changing IP address (load balancing)
//   - rate limiting this client
func NewPing() *Ping {
	// This is probably overkill, but we do expose this number to the public internet so might as-well use
	// this over math/rand. This is often to be recommended to be set as `pid` since this will allow more than
	// one client to work on the same machine. However pids can be larger than a u16 (see PID_MAX_LIMIT) which
	// means this doesn't hold due to overflow, furthermore this client isn't actually implemented to check
	// the return values, according to RFC the message body should be copied from the request into the
	// response payload, therefore it is our responsibility to match packets.
	b := [2]byte{}
	_, _ = rand.Read(b[:]) // Read from rand never returns an error
	seed := binary.LittleEndian.Uint16(b[:])
	return &Ping{
		id:        seed,
		addresses: &queryCache{m: &sync.Mutex{}, maxDrops: 3},
	}
}

// OneShot returns the time take for a ping to be replied too, or error if something went wrong.
func (p *Ping) OneShot(url string) (time.Duration, error) {
	// first we need to start listening this determines the underlying socket type we use and therefore
	// determine which DNS queries are valid so we need to do this first.

	// Create a listener for the IP we will use
	closer, err := p.startListening(url)
	defer closer()
	if err != nil {
		return 0, err
	}

	dnsTimeout, cancel := context.WithTimeoutCause(context.Background(), time.Second, pingTimeout{Duration: 100 * time.Millisecond})
	defer cancel()

	// Now find the IP address we will actually ping to
	err = p.addresses._DNSQuery(dnsTimeout, url, p.addrType)
	if err != nil {
		return 0, err
	}
	// Don't handle this [!ok] case in OneShot
	selectedIP, _ := p.addresses.Get()

	raw, err := p.makeOutgoingPacket(1)
	if err != nil {
		return 0, errors.Wrapf(err, "couldn't create outgoing %q packet", url)
	}

	// Actually write the echo request onto the connection:
	if err = p.writeEcho(selectedIP, raw); err != nil {
		return 0, err
	}
	begin := time.Now()

	// Now wait for the result
	buffer := make([]byte, 255)
	timeoutCtx, cancel := context.WithTimeoutCause(context.Background(), time.Second, pingTimeout{Duration: 100 * time.Millisecond})
	defer cancel()
	n, err := p.pingRead(timeoutCtx, buffer)
	duration := time.Since(begin)
	if err != nil {
		return duration, errors.Wrapf(err, "couldn't read packet from %q", url)
	}
	received, err := icmp.ParseMessage(protocolICMP, buffer[:n])
	if err != nil {
		return duration, errors.Wrapf(err, "couldn't parse raw packet from %q, %+v", url, received)
	}
	switch received.Type {
	case p.echoReply:
		return duration, nil
	default:
		return duration, errors.Errorf("Didn't receive a good message back from %q, got Code: %d", url, received.Code)
	}
}

// CreateChannel returns a channel of asynchronous ping results at the frequency of at most [rate], it may be
// slower is the network latency is larger than rate. Network error's are handled gracefully and will be on
// the channel as dropped packets, this function only returns an error if the internal state could not be
// configured or networking is completely disabled on this client.
func (p *Ping) CreateChannel(
	ctx context.Context,
	url string,
	rate PingsPerMinute,
	channelSize int,
) (<-chan PingResults, error) {
	result, _, err := p.CreateFlexibleChannel(ctx, url, rate, channelSize)
	return result, err
}

// PingsPerMinute represents how many pings per minute the [Ping] should target when [Ping.OneShot],
// [Ping.CreateChannel] or [Ping.CreateFlexibleChannel] is called. Use [AsFastAsPossible] or
// [NewPingsPerMinute] to create a [PingsPerMinute].
type PingsPerMinute struct {
	v float64
}

func AsFastAsPossible() PingsPerMinute {
	return PingsPerMinute{v: 0}
}

// NewPingsPerMinute less than 0 is treated as [AsFastAsPossible].
func NewPingsPerMinute(ppm float64) PingsPerMinute {
	if ppm < 0 {
		return PingsPerMinute{v: 0}
	}
	return PingsPerMinute{v: ppm}
}

// CreateFlexibleChannel is similar to [Ping.CreateChannel] but the speed of the channel can be updated by the
// second returned channel. Using [Speed].
func (p *Ping) CreateFlexibleChannel(
	ctx context.Context,
	url string,
	initialRate PingsPerMinute,
	channelSize int,
) (<-chan PingResults, chan<- Speed, error) {
	// Create a listener for the IP we will use
	closer, err := p.startListening(url)
	if err != nil {
		return nil, nil, err
	}

	initialRateLimit := p.buildRateLimiting(initialRate)

	dnsTimeout, cancel := context.WithTimeout(ctx, p.timeout)
	defer cancel()
	// Block the main thread to init this for the first time (most consumers will want to have a
	// [queryCache.GetLastIP] value as soon as this method returns), if we get an error let the main loop do
	// the retying.
	p.addresses.m.Lock()
	_ = p.addresses._DNSQuery(dnsTimeout, url, p.addrType)
	p.addresses.m.Unlock()

	client := make(chan PingResults, channelSize)
	speedChannel := make(chan Speed, channelSize)
	p.startChannel(ctx, client, closer, url, initialRateLimit, speedChannel)
	return client, speedChannel, nil
}

type PingResults struct {
	// Data is the data about this ping, containing the time taken for round trip or details if the packet was
	// dropped.
	Data PingDataPoint
	// IP is the address which this ping result was achieved from.
	IP net.IP
	// InternalErr represents some problem with [ping] package internal state which didn't gracefully handle
	// some network problem. Other network problems which are expected and represent dropped packets **should
	// be** handled gracefully and will be reported in the [PingDataPoint] field in the [Dropped].
	InternalErr error
}

type PingDataPoint struct {
	Duration   time.Duration
	Timestamp  time.Time
	DropReason Dropped
}

type Dropped byte

const (
	NotDropped Dropped = iota
	Timeout
	DNSFailure
	BadResponse

	TestDrop = 0xfe
)

type Speed byte

const (
	Faster Speed = iota
	Slower
	Fastest
)

func (s Speed) Delta(duration time.Duration) time.Duration {
	switch s {
	case Faster:
		return duration - (250 * time.Millisecond)
	case Slower:
		return duration + (250 * time.Millisecond)
	case Fastest:
		return 0
	default:
		panic("exhaustive:enforce")
	}
}

func (s Speed) String() string {
	switch s {
	case Faster:
		return "Faster"
	case Slower:
		return "Slower"
	case Fastest:
		return "Fastest"
	default:
		panic("exhaustive:enforce")
	}
}

func (p PingResults) String() string {
	switch {
	case p.IP == nil && p.InternalErr == nil:
		return "DNS Failure could not get IP"
	case p.InternalErr != nil:
		return "Internal API Error " + timestampString(p.Data) + " reason " + p.InternalErr.Error()
	default:
		return p.IP.String() + " | " + p.Data.String()
	}
}

func (p PingDataPoint) String() string {
	if p.Good() {
		return fmt.Sprintf("%s | %s", timestampString(p), p.Duration.String())
	}
	return fmt.Sprintf("%s | DROPPED, reason %q", timestampString(p), p.DropReason.String())
}

func (p PingDataPoint) Dropped() bool {
	return p.DropReason != NotDropped
}
func (p PingDataPoint) Good() bool {
	return p.DropReason == NotDropped
}
func (p PingDataPoint) Equal(other PingDataPoint) bool {
	return p.Duration == other.Duration && p.Timestamp.Equal(other.Timestamp) && p.DropReason == other.DropReason
}

func PingsPerMinuteToDuration(pingsPerMinute PingsPerMinute) time.Duration {
	if pingsPerMinute.v == 0 {
		return 0
	}
	gapBetweenPings := math.Round((60 * 1000) / pingsPerMinute.v)
	return time.Millisecond * time.Duration(gapBetweenPings)
}
