// Use of this source code is governed by a GPL-2 license that can be found in the LICENSE file.
//
// Copyright 2025 Lexer747
//
// SPDX-License-Identifier: GPL-2.0-only

package iterutils

import "iter"

// Map returns a new sequence where every value from the original sequence is transformed by f.
func Map[In any, Out any](s iter.Seq[In], f func(In) Out) iter.Seq[Out] {
	return func(yield func(Out) bool) {
		for in := range s {
			if !yield(f(in)) {
				return
			}
		}
	}
}

// Filter returns a new sequence only containing values from the original sequence for which the predicate f
// returns true.
func Filter[T any](s iter.Seq[T], f func(T) bool) iter.Seq[T] {
	return func(yield func(T) bool) {
		for in := range s {
			if f(in) {
				if !yield(in) {
					return
				}
			}
		}
	}
}
