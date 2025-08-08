// Use of this source code is governed by a GPL-2 license that can be found in the LICENSE file.
//
// Copyright 2025 Lexer747
//
// SPDX-License-Identifier: GPL-2.0-only

package backoff

import (
	"math"
	"time"
)

type ExpFallOff struct {
	// base is the initial smallest duration to wait in milliseconds
	base     float64
	curCount int
}

// https://en.wikipedia.org/wiki/Exponential_backoff
func NewExponentialBackoff(backoffStart time.Duration) *ExpFallOff {
	return &ExpFallOff{
		base: float64(backoffStart.Milliseconds()),
	}
}

// Wait marks the current backoff as failing and then blocks the current thread the corresponding amount of
// time.
func (e *ExpFallOff) Wait() {
	e.Fail()
	backoff := e.Duration()
	<-time.After(backoff)
}

// Success resets the current back off back to its initial duration marking the backoff as completed.
func (e *ExpFallOff) Success() {
	e.curCount = 0
}

// Fail marks the exp as failing increasing the next duration it will wait by an exponential amount.
func (e *ExpFallOff) Fail() {
	e.curCount++
}

// Duration returns the time that the backoff **would** wait if failure occurs.
func (e *ExpFallOff) Duration() time.Duration {
	computedBackOff := math.Pow(e.base, float64(e.curCount)) * float64(time.Millisecond)
	// use min max to clamp the backoff between the base and the max duration
	cappedBackOff := max(min(computedBackOff, math.MaxInt64/2), e.base)
	return time.Duration(cappedBackOff)
}
