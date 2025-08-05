// Use of this source code is governed by a GPL-2 license that can be found in the LICENSE file.
//
// Copyright 2024-2025 Lexer747
//
// SPDX-License-Identifier: GPL-2.0-only

package ping

import (
	"context"
	"net"
	"strings"
	"time"

	"github.com/Lexer747/acci-ping/utils/bytes"
	"github.com/Lexer747/acci-ping/utils/errors"
	"github.com/Lexer747/acci-ping/utils/sliceutils"
	"golang.org/x/net/icmp"
	"golang.org/x/net/ipv4"
	"golang.org/x/net/ipv6"
)

func timestampString(p PingDataPoint) string {
	return p.Timestamp.Format(time.RFC3339Nano)
}

func (p *Ping) startChannel(ctx context.Context, client chan<- PingResults, closer func(), url string, rateLimit *time.Ticker) {
	// TODO we actually need two goroutines, one for reading one for writing, then we should have a timer per
	// sent packet, and match them up according to the ID and sequence number. This would ensure more than one
	// client can work and that we maintain a matched tick rate should packets timeout and slow down the
	// sending routine.

	run := func() {
		defer close(client)
		defer closer()
		var seq uint16
		buffer := make([]byte, 255)
		var errorDuringLoop bool
		for {
			timestamp := time.Now()

			ip, newCloser := p.dnsRetry(ctx, url, client, timestamp, rateLimit, closer)
			if newCloser != nil {
				defer newCloser()
				closer = newCloser
				// Reset the timestamp, we were stuck in DNS for too long
				timestamp = time.Now()
			}

			if errorDuringLoop = p.pingOnChannel(ctx, timestamp, ip, seq, client, buffer); errorDuringLoop {
				// Keep track of this address as maybe being unreliable
				p.addresses.Dropped(ip)
			}
			seq++ // Deliberate wrap-around
			select {
			case <-ctx.Done():
				return
			default:
				if rateLimit != nil {
					// This throttles us if required, it will also drop ticks if we are pinging something very slow
					<-rateLimit.C
				}
			}
		}
	}
	go run()
}

func (p *Ping) buildRateLimiting(pingsPerMinute float64) *time.Ticker {
	p.timeout = 500 * time.Millisecond
	var rateLimit *time.Ticker
	// Zero is the sentinel, go as fast as possible
	if pingsPerMinute != 0 {
		maxPingDuration := PingsPerMinuteToDuration(pingsPerMinute)
		rateLimit = time.NewTicker(maxPingDuration)
		p.timeout = max(min(p.timeout, maxPingDuration), 500*time.Millisecond)
	}
	return rateLimit
}

func internalErr(IP net.IP, Timestamp time.Time, err error) PingResults {
	return PingResults{
		Data:        PingDataPoint{Timestamp: Timestamp},
		IP:          IP,
		InternalErr: err,
	}
}

func packetLoss(IP net.IP, Timestamp time.Time, Reason Dropped) PingResults {
	return PingResults{
		Data: PingDataPoint{
			Timestamp:  Timestamp,
			DropReason: Reason,
		},
		IP: IP,
	}
}

func goodPacket(IP net.IP, Duration time.Duration, Timestamp time.Time) PingResults {
	return PingResults{
		Data: PingDataPoint{
			Duration:   Duration,
			Timestamp:  Timestamp,
			DropReason: NotDropped,
		},
		IP: IP,
	}
}

// pingOnChannel performs a single ping to the already discovered IP, using the buffer as a scratch buffer,
// and writes ALL results to the channel (including errors). It self limits it's execution if it was called
// too recently compared to the desired rate.
func (p *Ping) pingOnChannel(
	ctx context.Context,
	timestamp time.Time,
	selected *addr,
	seq uint16,
	client chan<- PingResults,
	buffer []byte,
) bool {
	// Can gain some speed here by not remaking this each time, only to change the sequence number.
	raw, err := p.makeOutgoingPacket(seq)
	if err != nil {
		client <- internalErr(selected.ip, timestamp, err)
		return true
	}

	// Actually write the echo request onto the connection:
	if err = p.writeEcho(selected, raw); err != nil {
		client <- internalErr(selected.ip, timestamp, err)
		return true
	}
	begin := time.Now()
	timeoutErr := pingTimeout{Duration: p.timeout}
	timeoutCtx, cancel := context.WithTimeoutCause(ctx, p.timeout, timeoutErr)
	defer cancel()
	n, err := p.pingRead(timeoutCtx, buffer)
	duration := time.Since(begin)
	if err != nil && errors.Is(err, timeoutErr) {
		client <- packetLoss(selected.ip, timestamp, Timeout)
		return true
	} else if err != nil {
		client <- internalErr(selected.ip, timestamp, errors.Wrapf(err, "couldn't read packet from %q", p.currentURL))
		return true
	}
	received, err := icmp.ParseMessage(protocolICMP, buffer[:n])
	if err != nil {
		client <- internalErr(selected.ip, timestamp, errors.Wrapf(err, "couldn't parse raw packet from %q, %+v", p.currentURL, received))
		return true
	}
	body, ok := received.Body.(*icmp.Echo)
	if !ok {
		client <- internalErr(selected.ip, timestamp, errors.Wrapf(err, "couldn't parse body from %q, %+v", p.currentURL, received))
		return true
	}
	// Clear the buffer for next packet
	defer bytes.Clear(buffer, n)
	if body.Seq == int(seq) && received.Type == p.echoReply {
		client <- goodPacket(selected.ip, duration, timestamp)
		return false
	} else {
		client <- packetLoss(selected.ip, timestamp, BadResponse)
		return true
	}
}

type pingTimeout struct {
	time.Duration
}

func (pt pingTimeout) Error() string { return "PingTimeout {" + pt.String() + "}" }

func (p *Ping) pingRead(ctx context.Context, buffer []byte) (int, error) {
	type read struct {
		n   int
		err error
	}
	c := make(chan read)
	go func() {
		n, _, err := p.connect.ReadFrom(buffer)
		c <- read{n: n, err: err}
	}()
	select {
	case <-ctx.Done():
		err := context.Cause(ctx)
		return 0, err
	case success := <-c:
		return success.n, success.err
	}
}

func (p *Ping) makeOutgoingPacket(seq uint16) ([]byte, error) {
	outGoingPacket := icmp.Message{
		Type: p.echoType,
		Body: &icmp.Echo{
			// This identifier is purely to help distinguish other ongoing echos since we are listening on the
			// broad cast. Its a u16 in the spec, as is the Seq.
			ID:   int(p.id),
			Seq:  int(seq),
			Data: []byte("# acci-ping #"), // Something small but identifiable should someone want to block this traffic
		},
	}
	raw, err := outGoingPacket.Marshal(nil)
	if err != nil {
		return nil, err
	}
	return raw, nil
}

func (p *Ping) writeEcho(selectedIP *addr, raw []byte) error {
	if _, err := p.connect.WriteTo(raw, selectedIP.Get()); err != nil {
		return errors.Wrapf(err, "couldn't write packet to connection %q", p.currentURL)
	}
	return nil
}

func (p *Ping) startListening(url string) (closer func(), err error) {
	p.connect, p.addrType, err = p.evalListeningOptions()
	p.currentURL = url
	if err != nil {
		return nil, errors.Wrapf(err, "couldn't listen")
	}
	p.determineEchoType()
	return func() {
		p.connect.Close()
		p.currentURL = ""
	}, nil
}

// determineEchoType should only be called once [p.addrType] is set.
func (p *Ping) determineEchoType() {
	switch p.addrType {
	case _IP4, _UDP4:
		p.echoType = ipv4.ICMPTypeEcho
		p.echoReply = ipv4.ICMPTypeEchoReply
	case _IP6, _UDP6:
		p.echoType = ipv6.ICMPTypeEchoRequest
		p.echoReply = ipv6.ICMPTypeEchoReply
	case _UNRESOLVED:
		panic(" _UNRESOLVED, bug in startListening, did not set listening type")
	default:
		panic("determineEchoType, exhaustive:enforce")
	}
}

func (p *Ping) evalListeningOptions() (*icmp.PacketConn, addressType, error) {
	errs := []error{}
	for _, listenCfg := range listenList {
		conn, err := icmp.ListenPacket(listenCfg.network, listenCfg.address)
		if conn != nil && err == nil {
			return conn, listenCfg.addressType, nil
		}
	}
	strs := sliceutils.Map(errs, func(e error) string {
		return e.Error() + "\n"
	})
	return nil, 0, errors.New("couldn't listen for ping packets:\n" + strings.Join(strs, "- "))
}

var ipv4ListenAddr = net.IPv4zero
var ipv6ListenAddr = net.IPv6zero

var listenList = []listenerConfig{
	{network: "udp4", address: ipv4ListenAddr.String(), addressType: _UDP4},
	{network: "udp6", address: ipv6ListenAddr.String(), addressType: _UDP6},
	{network: "ip4:1", address: ipv4ListenAddr.String(), addressType: _IP4},
	{network: "ip6:ipv6-icmp", address: ipv6ListenAddr.String(), addressType: _IP6},

	// TODO add and support:
	//	- ip4:icmp
	//	- ip6:58
	//	- udp4 + custom addr
	//	- udp6 + custom addr
}

type listenerConfig struct {
	network, address string
	addressType
}

type addressType int

// We use underscores here because we don't want to export these constants, but uppercase makes for better
// acronyms - staticcheck doesn't agree with us.
//
//nolint:staticcheck
const (
	_UNRESOLVED addressType = 0

	_IP4  addressType = 1
	_UDP4 addressType = 2
	_IP6  addressType = 3
	_UDP6 addressType = 4
)
