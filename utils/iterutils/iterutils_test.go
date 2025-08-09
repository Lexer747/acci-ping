// Use of this source code is governed by a GPL-2 license that can be found in the LICENSE file.
//
// Copyright 2025 Lexer747
//
// SPDX-License-Identifier: GPL-2.0-only

package iterutils_test

import (
	"slices"
	"strconv"
	"testing"

	"github.com/Lexer747/acci-ping/utils/iterutils"

	"gotest.tools/v3/assert"
	is "gotest.tools/v3/assert/cmp"
)

func TestMap(t *testing.T) {
	t.Parallel()
	values := []int{1, 2, 3, 4, 5}
	seq := slices.Values(values)
	expected := []string{"1", "2", "3", "4", "5"}
	actual := slices.Collect(iterutils.Map(seq, strconv.Itoa))
	assert.Check(t, is.DeepEqual(expected, actual))
}
