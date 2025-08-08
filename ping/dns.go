// Use of this source code is governed by a GPL-2 license that can be found in the LICENSE file.
//
// Copyright 2024-2025 Lexer747
//
// SPDX-License-Identifier: GPL-2.0-only

package ping

import (
	"context"
	"fmt"
	"log/slog"
	"net"
	"slices"
	"strings"
	"sync"
	"time"

	"github.com/Lexer747/acci-ping/utils/check"
	"github.com/Lexer747/acci-ping/utils/errors"
)

// queryCache provides an interface for Ping to consume in which we respect the wishes of the servers we are
// causing load on, if they provide more than one address we should pick one at "random". Given we will re-use
// addresses from an original query we do the easier job of just round-robin.
//
// Thread safe.
type queryCache struct {
	m        *sync.Mutex
	store    []queryCacheItem
	index    int
	maxDrops uint
}

func (q *queryCache) socketedLockFree(addrType addressType) {
	check.Check(addrType != _UNRESOLVED, "cannot socket query cache to _UNRESOLVED")

	results := make([]queryCacheItem, 0, len(q.store))
	for _, item := range q.store {
		switch addrType {
		case _IP4, _UDP4:
			if isIpv4(item.addr.ip) {
				results = append(results, queryCacheItem{addr: New(addrType, item.addr.ip)})
			}
		case _IP6, _UDP6:
			if isIpv6(item.addr.ip) {
				results = append(results, queryCacheItem{addr: New(addrType, item.addr.ip)})
			}
		case _UNRESOLVED:
			// already handled above
		default:
			panic("unknown socket type (exhaustive:enforce)")
		}
	}
	// What I mean by this is that currently [Ping.evalListeningOptions] is greedy and just picks the first
	// network type which it can succeeded at listening too, instead it should be informed by what IPs are
	// returned. E.g. we can listen to [_UDP4] and [_UDP6], since the listen succeeds we pick [_UDP4], if the
	// DNSQuery resolves to an IPv6 address which are pigeon holed to an incorrect configuration due to this
	// panic here:
	check.Check(len(results) > 0, "TODO evalListeningOptions should work with socketed"+
		"\nMore accurately determining what network configuration we operate under")
	q.store = results
}

// GetLastIP will return the last IP address this cache used, formatted according to [net.IP.String].
func (q *queryCache) GetLastIP() string {
	q.m.Lock()
	defer q.m.Unlock()
	if len(q.store) > 0 {
		return q.store[q.index].addr.String()
	}
	return "<no ip>"
}

// Get will return an IP for use which is not considered stale and true. If the cache is exhausted an all IPs
// are stale then it will return nil and false.
func (q *queryCache) Get() (*addr, bool) {
	q.m.Lock()
	defer q.m.Unlock()
	return q.getLockFree()
}

func (q *queryCache) getLockFree() (*addr, bool) {
	if len(q.store) == 0 {
		// store is empty
		return nil, false
	}

	// iterate the cache, returning the first IP which isn't stale.
	start := q.index
	for {
		if r := q.store[q.index]; !r.stale {
			return r.addr, true
		}
		// this index was stale, advance the internal index.
		q.advance()
		// exit the loop if we went back to the start
		if q.index == start {
			break
		}
	}
	// No non-stale IPs found
	return nil, false
}

// Dropped tells this cache that the passed IP dropped a packet. Once enough drops have occurred for a given
// IP in the cache then the cache will consider that IP stale. Panic's if the IP isn't in the cache.
func (q *queryCache) Dropped(addr *addr) {
	q.m.Lock()
	defer q.m.Unlock()
	// We could keep the cache sorted and use binary searches, but for now we consider this a cold path and so
	// do not optimise for it.
	index := slices.IndexFunc(q.store, func(q queryCacheItem) bool {
		return q.addr.Equal(addr.ip)
	})
	check.Check(index != -1, "Unknown IP")

	// Now perform the update
	cur := q.store[index]
	stale := cur.dropCount > q.maxDrops
	q.store[q.index] = queryCacheItem{
		addr:      cur.addr,
		stale:     stale,
		dropCount: cur.dropCount + 1,
	}
}

func (q *queryCache) advance() {
	q.index = (q.index + 1) % len(q.store)
}

type queryCacheItem struct {
	addr      *addr
	stale     bool
	dropCount uint
}

func (qci queryCacheItem) String() string {
	var b strings.Builder
	b.WriteString("[" + qci.addr.String())
	if qci.stale {
		b.WriteString(", stale")
	} else {
		b.WriteString(", fresh")
	}
	fmt.Fprintf(&b, ", %d]", qci.dropCount)
	return b.String()
}

// _DNSQuery builds a new [ping.queryCache] for a given URL. If no valid addresses are found then an error is
// returned. The max drops specifies to the cache how many dropped packets an address is allowed before we
// consider that address too un-reliable, services may rotate their addresses in which case this cache will
// clear itself of these now defunct addresses. If maxDrops is 0, then only a single dropped packet will mean
// the address is considered stale.
func (q *queryCache) _DNSQuery(ctx context.Context, url string, addrType addressType) error {
	resolver := &net.Resolver{}
	// This doesn't need a lock because it should only be called by things already holding the lock
	ips, err := resolver.LookupIP(ctx, "ip", url)
	if err != nil {
		return errors.Wrapf(err, "couldn't get IP for %q (DNS failure)", url)
	}
	if len(ips) == 0 {
		return errors.Errorf("Couldn't resolve %q to any address. Network down? (DNS failure)", url)
	}

	// Only use IPs which are of the socket type we're actually operating under. If unresolved we forward all IPs as successful.
	results := make([]queryCacheItem, 0, len(ips))
	for _, ip := range ips {
		switch addrType {
		case _IP4, _UDP4:
			if isIpv4(ip) {
				results = append(results, queryCacheItem{addr: New(addrType, ip)})
			}
		case _IP6, _UDP6:
			if isIpv6(ip) {
				results = append(results, queryCacheItem{addr: New(addrType, ip)})
			}
		case _UNRESOLVED:
			results = append(results, queryCacheItem{addr: New(addrType, ip)})
		default:
			panic("unknown socket type (exhaustive:enforce)")
		}
	}
	if len(results) == 0 {
		return errors.Errorf("Couldn't resolve %q to a valid IP address (DNS failure)", url)
	}
	q.store = results
	return nil
}

func (p *Ping) dnsRetry(
	ctx context.Context,
	url string,
	client chan<- PingResults,
	timestamp time.Time,
	rateLimit *time.Ticker,
	closer func(),
) (*addr, func()) {
	p.addresses.m.Lock()
	defer p.addresses.m.Unlock()
	// fast path, just get the ip from the cache if it exists
	if ip, ok := p.addresses.getLockFree(); ok {
		return ip, nil
	}

	var err error
	var newCloser func()
	slog.Debug("dns hard retry entry", "url", url, "cause", "p.addresses empty")
	// I know that a goto and label looks scary but I assure you dear reader that this is sane, since our
	// control flow and error handling path implies an infinite loop (Because we need to try listening for
	// packets forever unless cancelled by the parent). This infinite loop is coupled to the results of the
	// DNS query, therefore to truly retry against the network requires a loop around the two inner loops,
	// which if done with a `for` construct is much less legible.
HARD_RETRY:
	// Keeping doing a DNS query until we get a valid result, count each failure as a dropped packet
	for {
		// start again, do a new DNS query
		timeoutErr := pingTimeout{Duration: p.timeout}
		// FWIW I don't think this timeout actually makes a difference on most platforms, it seems like OS
		// controlled Dialer and Resolver Timeouts come into effect before this will. It may also be
		// https://github.com/golang/go/issues/36848
		dnsTimeout, cancel := context.WithTimeoutCause(ctx, p.timeout, timeoutErr)
		defer cancel()
		err = p.addresses._DNSQuery(dnsTimeout, url, _UNRESOLVED)
		if err != nil {
			client <- packetLoss(nil, timestamp, DNSFailure)
			if rateLimit != nil {
				<-rateLimit.C
			}
			timestamp = time.Now()
			clear(p.addresses.store)
		} else {
			break
		}

		// Now is also a sane point in the function to determine if the parent wants us to stop spinning our
		// hamster wheel trying to find a packet. We only check this so we gracefully exit instead of spamming
		// DNS queries as we die.
		if err := ctx.Err(); err != nil {
			return nil, nil
		}
	}
	// Reset our listening, it's a chance our NIC died in which case we need to restart this. I don't
	// think we can tell that the inner listener died. Don't use exp back off here, this can only be a
	// client issue.
	closer()
	for {
		newCloser, err = p.startListening(url)
		if err == nil {
			p.addresses.socketedLockFree(p.addrType)
			break
		}
		// Now is a sane point in the function to determine if the parent wants us to stop spinning our
		// hamster wheel trying to find a packet. Overly checking this value is wasteful and unhelpful, we
		// expect the ratelimited loop to do that the majority of the time.
		if err := ctx.Err(); err != nil {
			return nil, nil
		}
		goto HARD_RETRY
	}

	ip, ok := p.addresses.getLockFree()
	if !ok {
		clear(p.addresses.store)
		goto HARD_RETRY // Avoid recursion, if we made it here either we have a fresh restart the entire address pool is exhausted
	}
	slog.Debug("dns hard retry exit", "url", url, "ip", ip)
	return ip, newCloser
}
