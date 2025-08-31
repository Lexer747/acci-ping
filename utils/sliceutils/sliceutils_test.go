// Use of this source code is governed by a GPL-2 license that can be found in the LICENSE file.
//
// Copyright 2025 Lexer747
//
// SPDX-License-Identifier: GPL-2.0-only

package sliceutils_test

import (
	"testing"

	"gotest.tools/v3/assert"
	is "gotest.tools/v3/assert/cmp"

	"github.com/Lexer747/acci-ping/utils/sliceutils"
)

func TestSplitN(t *testing.T) {
	t.Parallel()
	t.Run("Basic", testCase[int]{
		Input:  []int{1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14, 15, 16},
		SplitN: 4,
		Output: [][]int{
			{1, 2, 3, 4},
			{5, 6, 7, 8},
			{9, 10, 11, 12},
			{13, 14, 15, 16},
		},
	}.Run)
	t.Run("None 0-modulo", testCase[int]{
		Input:  []int{1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14, 15, 16},
		SplitN: 5,
		Output: [][]int{
			{1, 2, 3, 4, 5},
			{6, 7, 8, 9, 10},
			{11, 12, 13, 14, 15},
			{16},
		},
	}.Run)
	t.Run("Larger means single", testCase[int]{
		Input:  []int{1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14, 15, 16},
		SplitN: 20,
		Output: [][]int{
			{1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14, 15, 16},
		},
	}.Run)
	t.Run("String type as bytes", testCase[byte]{
		Input:  []byte("abcdefghijklmnopqrstuvwxyz_abcdefghijklmnopqrstuvwxyz_abcdefghijklmnopqrstuvwxyz"),
		SplitN: 26,
		Output: [][]byte{
			[]byte("abcdefghijklmnopqrstuvwxyz"),
			[]byte("_abcdefghijklmnopqrstuvwxy"),
			[]byte("z_abcdefghijklmnopqrstuvwx"),
			[]byte("yz"),
		},
	}.Run)
	t.Run("String type as runes", testCase[rune]{
		Input:  []rune("aa ⚠️ bb ⚠️ cc ⚠️ "),
		SplitN: 6,
		Output: [][]rune{
			[]rune("aa ⚠️ "),
			[]rune("bb ⚠️ "),
			[]rune("cc ⚠️ "),
		},
	}.Run)
}

type testCase[T comparable] struct {
	Input  []T
	Output [][]T
	SplitN int
}

func (tc testCase[T]) Run(t *testing.T) {
	t.Helper()
	t.Parallel()
	result := sliceutils.SplitN(tc.Input, tc.SplitN)
	assert.Check(t, is.Equal(len(tc.Output), len(result)))
	for i := range min(len(tc.Output), len(result)) {
		assert.Check(t, is.DeepEqual(tc.Output[i], result[i]), "split %d: tc.Output[i]:%s, result[i]:%s", i, tc.Output[i], result[i])
	}
}
