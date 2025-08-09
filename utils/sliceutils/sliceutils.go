// Use of this source code is governed by a GPL-2 license that can be found in the LICENSE file.
//
// Copyright 2024-2025 Lexer747
//
// SPDX-License-Identifier: GPL-2.0-only

package sliceutils

import (
	"fmt"
	"math/rand/v2"
	"slices"
	"strings"
)

// Map will create a new slice of type [OUT] from which every element comes from [slice] after the function
// [f] has been applied to it.
func Map[IN, OUT any, S ~[]IN](slice S, f func(IN) OUT) []OUT {
	ret := make([]OUT, len(slice))
	for i, in := range slice {
		ret[i] = f(in)
	}
	return ret
}

// OneOf as of go1.24 is simply a wrapper around [slices.ContainsFunc]. Reporting whether at least one element
// e of slice satisfies f(e).
func OneOf[S ~[]T, T any](slice S, f func(T) bool) bool {
	return slices.ContainsFunc(slice, f)
}

// AllOf reports whether all elements e of [slice] satisfy f(e).
func AllOf[S ~[]T, T any](slice S, f func(T) bool) bool {
	for _, item := range slice {
		if !f(item) {
			return false
		}
	}
	return true
}

// Fold is a Left fold over the [slice] of e, applying f(base, e) to every element and returning the final
// [OUT] accumulation.
func Fold[IN, OUT any, S ~[]IN](slice S, base OUT, f func(IN, OUT) OUT) OUT {
	ret := base
	for _, in := range slice {
		ret = f(in, ret)
	}
	return ret
}

// Shuffle uses [rand.Shuffle] to shuffle all the elements of the [slice] and return a shuffled [clone] of the
// input.
func Shuffle[S ~[]T, T any](slice S) S {
	ret := slices.Clone(slice)
	shuf := func(i, j int) {
		t := ret[i]
		ret[i] = ret[j]
		ret[j] = t
	}
	rand.Shuffle(len(ret), shuf)
	return ret
}

// Join is a stringifier over the slice of [fmt.Stringer] elements, calling String on each one and joining them
// all with the separator.
func Join[S ~[]T, T fmt.Stringer](slice S, sep string) string {
	return strings.Join(Map(slice, T.String), sep)
}

// JoinFunc will call [strings.Join] with the separator on the slice after applying [f] to every element,
// returning the result.
func JoinFunc[S ~[]T, T any](slice S, f func(T) string, sep string) string {
	return strings.Join(Map(slice, f), sep)
}

// Remove will remove the elements from the [slice], if no elements are found in the [slice] then a shallow
// copy is returned.
func Remove[S ~[]T, T comparable](slice S, elements ...T) S {
	toDelete := map[T]struct{}{}
	for _, t := range elements {
		toDelete[t] = struct{}{}
	}
	ret := S{}
	for _, s := range slice {
		if _, found := toDelete[s]; found {
			continue
		} else {
			ret = append(ret, s)
		}
	}
	return ret
}

// SplitN converts a slice in to slices of length at most [splitAfterN], where each return slice is cut from
// the original input slice.
func SplitN[S ~[]T, T any](slice S, splitAfterN int) []S {
	splits := len(slice) / splitAfterN
	if len(slice)%splitAfterN != 0 {
		splits++
	}
	ret := make([]S, splits)
	for i := range splits {
		end := min(len(slice), splitAfterN)
		ret[i] = slice[:end]
		slice = slice[end:]
	}
	return ret
}
