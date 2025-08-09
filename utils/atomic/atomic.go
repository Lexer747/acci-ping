// Use of this source code is governed by a GPL-2 license that can be found in the LICENSE file.
//
// Copyright 2025 Lexer747
//
// SPDX-License-Identifier: GPL-2.0-only

package atomic

import "sync"

// Atomic much like [sync.Mutex] must not be copied.
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

func (a *Of[T]) Get() T {
	a.m.RLock()
	defer a.m.RUnlock()
	return a.storage
}

func (a *Of[T]) Set(t T) {
	a.m.Lock()
	defer a.m.Unlock()
	a.storage = t
}
