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
	"gotest.tools/v3/assert"
	is "gotest.tools/v3/assert/cmp"
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
	_, err := p.CreateChannel(ctx, "www.google.com", -1, 0)
	assert.Assert(t, is.ErrorContains(err, "Invalid pings per minute"))
	const testSize = 5
	channel, err := p.CreateChannel(ctx, "www.google.com", 0, testSize)
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
	channel1, err := p1.CreateChannel(ctx, "www.google.com", 0, testSize)
	assert.NilError(t, err)
	channel2, err := p2.CreateChannel(ctx, "www.bing.com", 0, testSize)
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

func networkingEnvGuard(t *testing.T) {
	t.Helper()
	if !env.SHOULD_TEST_NETWORK() {
		t.Skip("SHOULD_TEST_NETWORK disabled")
	}
}
