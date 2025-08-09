// Use of this source code is governed by a GPL-2 license that can be found in the LICENSE file.
//
// Copyright 2025 Lexer747
//
// SPDX-License-Identifier: GPL-2.0-only

package iterutils

import "iter"

func Map[IN, OUT any](in iter.Seq[IN], f func(IN) OUT) iter.Seq[OUT] {
	return func(yield func(OUT) bool) {
		in(func(v IN) bool {
			return yield(f(v))
		})
	}
}
