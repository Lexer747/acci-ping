// Use of this source code is governed by a GPL-2 license that can be found in the LICENSE file.
//
// Copyright 2025-2026 Lexer747
//
// SPDX-License-Identifier: GPL-2.0-only

package atomic

import "sync"

// Of is an Atomic of [T] much like [sync.Mutex] must not be copied.
//
// Here the word "Atomic" is used as an idea not an implementation, the actual storage is guarded in a
// concurrent way with a mutex, it merely provides a convenient way (when critical sections are too
// cumbersome) to ensure thread safe of a given value or struct which normally would not be safe. There are
// obviously ways to misuse this such as taking pointers, of the returned values.
type Of[T any] struct {
	m       *sync.RWMutex
	storage T
}

func New[T any]() Of[T] {
	return Of[T]{m: &sync.RWMutex{}}
}

func Init[T any](t T) Of[T] {
	return Of[T]{m: &sync.RWMutex{}, storage: t}
}

// Get accesses the storage and returns it locking it from other writers.
func (a *Of[T]) Get() T {
	a.m.RLock()
	defer a.m.RUnlock()
	return a.storage
}

// Set accesses the storage and sets the value locking it from all other parties.
func (a *Of[T]) Set(t T) {
	a.m.Lock()
	defer a.m.Unlock()
	a.storage = t
}
