// Use of this source code is governed by a GPL-2 license that can be found in the LICENSE file.
//
// Copyright 2024-2025 Lexer747
//
// SPDX-License-Identifier: GPL-2.0-only

package ping_test

import (
	"cmp"
	"context"
	"testing"
	"time"

	"github.com/Lexer747/acci-ping/ping"
	"github.com/Lexer747/acci-ping/utils/env"
	"github.com/Lexer747/acci-ping/utils/th"
	"gotest.tools/v3/assert"
)

func TestOneShot_google_com(t *testing.T) {
	networkingEnvGuard(t)
	t.Parallel()
	p := ping.NewPing()
	duration, err := p.OneShot("www.google.com")
	assert.NilError(t, err)
	assert.Assert(t, cmp.Compare(duration, time.Millisecond) > 0)
}

func TestChannel_google_com(t *testing.T) {
	networkingEnvGuard(t)
	t.Parallel()
	p := ping.NewPing()
	ctx, cancelFunc := context.WithCancel(t.Context())
	const testSize = 5
	channel, err := p.CreateChannel(ctx, "www.google.com", ping.AsFastAsPossible(), testSize)
	assert.NilError(t, err)
	for range testSize {
		result := <-channel
		assert.Check(t, !result.Data.Dropped(), result.Data.String())
		assert.Assert(t, cmp.Compare(result.Data.Duration, time.Millisecond) > 0)
	}
	cancelFunc()
}

func TestMultiplePings(t *testing.T) {
	networkingEnvGuard(t)
	t.Parallel()
	p1 := ping.NewPing()
	p2 := ping.NewPing()
	ctx, cancelFunc := context.WithCancel(t.Context())
	const testSize = 5
	channel1, err := p1.CreateChannel(ctx, "www.google.com", ping.AsFastAsPossible(), testSize)
	assert.NilError(t, err)
	channel2, err := p2.CreateChannel(ctx, "www.bing.com", ping.AsFastAsPossible(), testSize)
	assert.NilError(t, err)
	for range testSize {
		result := <-channel1
		assert.Check(t, !result.Data.Dropped(), result.Data.String())
		assert.Assert(t, cmp.Compare(result.Data.Duration, time.Millisecond) > 0)
		result = <-channel2
		assert.Check(t, !result.Data.Dropped(), result.Data.String())
		assert.Assert(t, cmp.Compare(result.Data.Duration, time.Millisecond) > 0)
	}
	cancelFunc()
}

func TestUint16Wrapping(t *testing.T) {
	t.Parallel()
	var i uint16 = 1
	for i != 0 {
		i++
	}
	assert.Equal(t, uint16(1), i+1)
}

func TestContextCancel(t *testing.T) {
	networkingEnvGuard(t)
	t.Parallel()
	parent := t.Context()
	// Over all test timeout should exceed the happy path, but be reasonable for CI
	th.TestWithTimeout(t, 5*time.Second, func() {
		p := ping.NewPing()
		ctx, cancelFunc := context.WithTimeout(parent, time.Second)
		defer cancelFunc()
		// Too slow to ever complete
		channel, err := p.CreateChannel(ctx, "www.google.com", ping.NewPingsPerMinute(0.0000001), 0)
		assert.NilError(t, err)
		// swallow the first result, this won't be delayed by the ticker
		<-channel
		select {
		case result := <-channel:
			t.Fatalf("unexpected cancellable test: %s", result)
		case <-ctx.Done():
			// Success, return
		}
	})
}

func TestSpeedChange(t *testing.T) {
	networkingEnvGuard(t)
	t.Parallel()
	ctx := t.Context()
	// Over all test timeout should exceed the happy path, but be reasonable for CI
	th.TestWithTimeout(t, time.Minute+(30*time.Second), func() {
		p := ping.NewPing()
		ctx, cancelFunc := context.WithTimeout(ctx, time.Minute)
		defer cancelFunc()
		const testSize = 5
		// Too slow to ever complete
		channel, speedChannel, err := p.CreateFlexibleChannel(ctx, "www.google.com", ping.NewPingsPerMinute(0.0000001), testSize)
		assert.NilError(t, err)
		select {
		case result := <-channel:
			t.Fatalf("unexpected result in speed change test: %s", result)
		case <-time.After(10 * time.Millisecond):
			speedChannel <- ping.Fastest
		}
		for range testSize {
			result := <-channel
			assert.Check(t, !result.Data.Dropped(), result.Data.String())
			assert.Assert(t, cmp.Compare(result.Data.Duration, time.Millisecond) > 0)
		}
	})
}

func networkingEnvGuard(t *testing.T) {
	t.Helper()
	if !env.SHOULD_TEST_NETWORK() {
		t.Skip("SHOULD_TEST_NETWORK disabled")
	}
}
