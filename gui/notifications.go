// Use of this source code is governed by a GPL-2 license that can be found in the LICENSE file.
//
// Copyright 2025 Lexer747
//
// SPDX-License-Identifier: GPL-2.0-only

package gui

import (
	"slices"
	"sync"
	"time"

	"github.com/Lexer747/acci-ping/terminal"
	"github.com/Lexer747/acci-ping/utils/bytes"
	"github.com/Lexer747/acci-ping/utils/sliceutils"
)

// Notification is a thread safe gui primitive that represents a collection of [T], which are transient in
// nature. As in they may only be presented on the GUI for some fixed duration before being removed from the
// collection.
//
// The [Drawable] function callback provides a way for the user of this struct to customise how the collection
// of [T] should actually be presented.
type Notification[T any] struct {
	m        *sync.Mutex
	storage  map[int]stored[T]
	key      int
	drawFunc func(terminal.Size, []T) Draw
	lastSize terminal.Size
}

// NewNotification builds a new thread safe collection of [T]. The function passed will be called with a
// time-ordered (oldest first, newest last) list of T to be drawn. The returned [Draw] from the callback will
// be applied to the buffer during [Notification.NewValue] and [Notification.NewSize].
func NewNotification[T any](initial terminal.Size, Drawable func(terminal.Size, []T) Draw) *Notification[T] {
	return &Notification[T]{
		m:        &sync.Mutex{},
		storage:  map[int]stored[T]{},
		key:      0,
		drawFunc: Drawable,
		lastSize: initial,
	}
}

// NewValue adds a [T] to the collection for the given timeout, once timeout has passed the [T] is
// automatically removed and a re-render performed.
//
// g is the GUI in which the paint update should be sent too.
func (n *Notification[T]) NewValue(g GUI, toShow T, buffer *bytes.SafeBuffer, timeout time.Duration) {
	n.m.Lock()
	// First generate a unique id for this value and add it to our map.
	key := n.insert(toShow)
	// Now refresh the window size and write the value to the window
	g.Paint(n.locklessRender(n.lastSize, buffer))
	n.m.Unlock()
	// Now after some timeout, remove the value and re-render
	go func() {
		<-time.After(timeout)
		n.m.Lock()
		delete(n.storage, key)
		// don't use the size from this call, it may have changed.
		g.Paint(n.locklessRender(n.lastSize, buffer))
		n.m.Unlock()
	}()
}

// NewSize should be called when only a size update has occurred. This will cause all future re-paints to use
// this size and cause all currently stored items to be re-drawn.
//
// g is the GUI in which the paint update should be sent too.
func (n *Notification[T]) NewSize(g GUI, size terminal.Size, buffer *bytes.SafeBuffer) {
	n.m.Lock()
	defer n.m.Unlock()
	n.lastSize = size
	g.Paint(n.locklessRender(size, buffer))
}

// stored wraps any T with a timestamp
type stored[T any] struct {
	timestamp time.Time
	value     T
}

func newStored[T any](toShow T) stored[T] {
	return stored[T]{timestamp: time.Now(), value: toShow}
}

// insert generates a unique key for the value that's inserted.
func (n *Notification[T]) insert(toShow T) int {
	n.key++
	newVar := newStored(toShow)
	n.storage[n.key] = newVar
	return n.key
}

// locklessRender writes the drawable to the buffer [b].
func (n *Notification[T]) locklessRender(size terminal.Size, b *bytes.SafeBuffer) PaintUpdate {
	ret := None
	hasData := b.Len() != 0
	b.Reset()
	if len(n.storage) == 0 {
		if hasData {
			ret = ret | Invalidate
		}
		return ret
	}
	toasts := n.orderToasts()
	drawable := n.drawFunc(size, toasts)
	drawable.Draw(size, b)
	return ret | Paint
}

// orderToasts will return a slice of ordered [T] where they're sorted by the timestamp in which they were
// added to the storage, should only be called while the lock is held.
func (n *Notification[T]) orderToasts() []T {
	order := make([]stored[T], 0, len(n.storage))
	for _, t := range n.storage {
		idx, _ := slices.BinarySearchFunc(order, t, func(a, b stored[T]) int { return a.timestamp.Compare(b.timestamp) })
		order = slices.Insert(order, idx, t)
	}
	return sliceutils.Map(order, func(s stored[T]) T { return s.value })
}
