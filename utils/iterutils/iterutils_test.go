// Use of this source code is governed by a GPL-2 license that can be found in the LICENSE file.
//
// Copyright 2025 Lexer747
//
// SPDX-License-Identifier: GPL-2.0-only

package iterutils_test

import (
	"slices"
	"testing"

	"github.com/Lexer747/acci-ping/utils/iterutils"
	"gotest.tools/v3/assert"
)

func TestMap(t *testing.T) {
	t.Parallel()
	input := slices.Values([]string{"a", "b", "c"})
	expected := []string{"@a", "@b", "@c"}
	actual := slices.Collect(iterutils.Map(input, func(s string) string { return "@" + s }))
	assert.DeepEqual(t, expected, actual)
}
func TestFilter(t *testing.T) {
	t.Parallel()
	input := slices.Values([]string{"a", "b", "c"})
	expected := []string{"b"}
	actual := slices.Collect(iterutils.Filter(input, func(s string) bool { return s == "b" }))
	assert.DeepEqual(t, expected, actual)
}
