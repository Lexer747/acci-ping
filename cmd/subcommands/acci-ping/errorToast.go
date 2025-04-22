// Use of this source code is governed by a GPL-2 license that can be found in the LICENSE file.
//
// Copyright 2024-2025 Lexer747
//
// SPDX-License-Identifier: GPL-2.0-only

package acciping

import (
	"bytes"
	"context"
	"log/slog"
	"math/rand/v2"
	"slices"
	"strings"
	"sync"
	"time"

	"github.com/Lexer747/acci-ping/draw"
	"github.com/Lexer747/acci-ping/gui"
	"github.com/Lexer747/acci-ping/gui/themes"
	"github.com/Lexer747/acci-ping/terminal"
	"github.com/Lexer747/acci-ping/utils/sliceutils"
)

// toastNotifications which should only be called once the paint buffer is initialised.
func (app *Application) toastNotifications(ctx context.Context, terminalSizeUpdates <-chan terminal.Size) {
	store := toastStore{
		Mutex:  &sync.Mutex{},
		toasts: map[int]toast{},
	}
	toastBuffer := app.drawBuffer.Get(draw.ToastIndex)
	for {
		select {
		case <-ctx.Done():
			return
		case newSize := <-terminalSizeUpdates:
			store.Lock()
			app.paint(store.render(newSize, toastBuffer))
			store.Unlock()
		case toShow := <-app.errorChannel:
			if toShow == nil {
				continue
			}
			slog.Info("New Error being shown", "err", toShow)

			// A new error has been surfaced:
			store.Lock()
			// First generate a unique id for this error and add it to our map.
			key := store.insertToast(toShow)
			// Now refresh the window size and write the toast notification to the window
			app.paint(store.render(app.term.GetSize(), toastBuffer))
			store.Unlock()
			// Now after some timeout, remove the notification and re-render
			go func() {
				<-time.After(20 * time.Second)
				store.Lock()
				delete(store.toasts, key)
				app.paint(store.render(app.term.GetSize(), toastBuffer))
				store.Unlock()
			}()
		}
	}
}

type toast struct {
	timestamp time.Time
	err       string
}

type toastStore struct {
	*sync.Mutex
	toasts map[int]toast
}

// insertToast should only be called while the lock is held
func (ts toastStore) insertToast(toShow error) int {
	var key int
	for {
		key = rand.Int() //nolint:gosec
		_, ok := ts.toasts[key]
		if !ok {
			ts.toasts[key] = toast{
				timestamp: time.Now(),
				err:       toShow.Error(),
			}
			break
		}
	}
	return key
}

// render should only be called while the lock is held
func (ts toastStore) render(size terminal.Size, b *bytes.Buffer) paintUpdate {
	ret := None
	hasData := b.Len() != 0
	b.Reset()
	if len(ts.toasts) == 0 {
		if hasData {
			ret = ret | Invalidate
		}
		return ret
	}
	toasts := ts.orderToasts()
	box := makeBox(size, toasts)
	box.Draw(size, b)
	return ret | Paint
}

// orderToasts will return a slice of ordered toasts where they're sorted by the timestamp in which they were
// added to the storage, it also returns the longest error string. Should only be called while the lock is
// held.
func (ts toastStore) orderToasts() []toast {
	order := make([]toast, 0, len(ts.toasts))
	for _, t := range ts.toasts {
		idx, _ := slices.BinarySearchFunc(order, t, func(a, b toast) int { return a.timestamp.Compare(b.timestamp) })
		order = slices.Insert(order, idx, t)
	}
	return order
}

const title = "⚠️  An Error Occurred ⚠️"

func makeBox(size terminal.Size, ts []toast) gui.Box {
	text := make([]gui.Typography, 0, len(ts)+1)
	text = append(text, gui.Typography{ToPrint: themes.Negative(title), TextLen: len(title) - 10, Alignment: gui.Centre})
	// TODO wrap differently when this might be ontop/underneath the help box.
	maxSize := (size.Width * 3) / 4
	for _, t := range ts {
		for line := range strings.SplitSeq(t.err, "\n") {
			if len([]rune(line)) >= maxSize {
				// TODO split on spaces instead ...
				for _, splitLine := range sliceutils.SplitN([]rune(line), maxSize) {
					text = append(text, gui.Typography{
						ToPrint:        string(splitLine),
						LenFromToPrint: true,
						Alignment:      gui.Centre,
					})
				}
			} else {
				text = append(text, gui.Typography{
					ToPrint:        line,
					LenFromToPrint: true,
					Alignment:      gui.Centre,
				})
			}
		}
	}
	return gui.Box{
		BoxText: text,
		Position: gui.Position{
			Vertical:   gui.Middle,
			Horizontal: gui.Centre,
			Padding:    gui.NoPadding,
		},
		Style:         gui.RoundedCorners,
		Configuration: gui.BoxCfg{},
	}
}
