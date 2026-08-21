// Use of this source code is governed by a GPL-2 license that can be found in the LICENSE file.
//
// Copyright 2026 Lexer747
//
// SPDX-License-Identifier: GPL-2.0-only

package syncutils

import "sync"

// LockGuard provides a proof of ownership model for locks from [sync], a lock free API may be written such
// that each function in said API requires that a [LockGuard] is also given. This is why the type arg exists
// merely to be a compile time opt-in for a package specific guard. Runtime checks are opt-in and would
// prevent lock guard sharing from multiple lock guards from the same package.
//
// Example usage:
//
//	type myPackageLock struct{} // private to the package
//	type LockFree *syncutils.LockGuard[myPackageLock] // Alias for convenience
//
//	// normally dangerous lock free interface now hands out a this proof instead of void.
//	func (foo *Bar) Lock() LockFree {
//		foo.m.Lock()
//		return syncutils.New[myPackageLock](foo.m)
//	}
//	func (foo *Bar) Unlock(proof LockFree) { foo.m.Unlock() }
//	// The safe version for use when perf isn't a concern
//	func (foo *Bar) SafeAlgo() {
//		foo.m.Lock()
//		defer foo.m.Unlock()
//		foo.hardAlgo()
//	}
//	// Normally a flight risk as someone could call this without the lock, but since they
//	// have to pass proof we can verify if it's lock (and handle that accordingly)
//	func (foo *Bar) LockFreeHardAlgo(proof LockFree) {
//		if proof.IsMine(foo.M) {
//			foo.hardAlgo()
//			return
//		}
//		panic("wrong proof of locking!")
//	}
type LockGuard[T any] struct {
	l sync.Locker
}

// New constructs a new lock guard on the given locker, no operations are performed on the lock its merely
// held so that [IsMine] can confirm that this is in fact the same lock that was reference.
func New[T any](locker sync.Locker) *LockGuard[T] {
	return &LockGuard[T]{l: locker}
}

// IsMine returns true if and only if the passed lock is the same as the one which constructed the lock guard.
func (lg *LockGuard[T]) IsMine(locker sync.Locker) bool {
	return lg.l == locker
}

// Fault returns true if and only if the passed lock isn't the original one we used to construct the lock guard.
func (lg *LockGuard[T]) Fault(locker sync.Locker) bool {
	return lg.l != locker
}
