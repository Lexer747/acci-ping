// Use of this source code is governed by a GPL-2 license that can be found in the LICENSE file.
//
// Copyright 2024-2025 Lexer747
//
// SPDX-License-Identifier: GPL-2.0-only

package channels

import (
	"context"
)

// TeeBufferedChannel, duplicates the channel such that both returned channels receive values from [c], this
// duplication is unsynchronised. Both channels are closed when the [ctx] is done.
func TeeBufferedChannel[T any](ctx context.Context, c <-chan T, channelSize int) (
	chan T,
	chan T,
) {
	left := make(chan T, channelSize)
	right := make(chan T, channelSize)
	go func() {
		defer close(left)
		defer close(right)
		for {
			select {
			case <-ctx.Done():
			case v := <-c:
				go func() {
					left <- v
				}()
				go func() {
					right <- v
				}()
			}
		}
	}()
	return left, right
}

// FanInFanOut takes ownership of single channel of [T] and returns [fanOutCount] channels which all receive
// the duplicated output of the original input channel. The actually fan out will be sent over multiple
// go-routines so each returned channel will be updated in a random order. If the ctx is done then all
// channels will be closed apart from the original input channel [c].
func FanInFanOut[T any](ctx context.Context, c <-chan T, channelSize, fanOutCount int) (
	Out []<-chan T,
) {
	outChans := make([]chan T, fanOutCount)
	for i := range fanOutCount {
		outChans[i] = make(chan T, channelSize)
	}
	go func() {
		defer func() {
			for i := range fanOutCount {
				close(outChans[i])
			}
		}()
		for {
			select {
			case <-ctx.Done():
			case v := <-c:
				for i := range fanOutCount {
					go func() {
						outChans[i] <- v
					}()
				}
			}
		}
	}()
	result := make([]<-chan T, fanOutCount)
	for i := range fanOutCount {
		result[i] = outChans[i]
	}
	return result
}
