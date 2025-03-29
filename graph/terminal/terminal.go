// Use of this source code is governed by a GPL-2 license that can be found in the LICENSE file.
//
// Copyright 2024-2025 Lexer747
//
// SPDX-License-Identifier: GPL-2.0-only

package terminal

import (
	"context"
	"io"
	"log/slog"
	"os"
	"slices"
	"strconv"
	"strings"
	"sync"

	"github.com/Lexer747/acci-ping/graph/terminal/ansi"
	"github.com/Lexer747/acci-ping/utils"
	"github.com/Lexer747/acci-ping/utils/bytes"
	"github.com/Lexer747/acci-ping/utils/errors"

	"golang.org/x/term"
)

// Size represents the size of a terminal, the units are in terms of numbers of characters.
type Size struct {
	Height int // Height can also be thought of as "y" in the graph context, or "number of rows"
	Width  int // Width can also be thought of as "x" in the graph context, or "number of columns"
}

func (s Size) String() string {
	return "W: " + strconv.Itoa(s.Width) + " H: " + strconv.Itoa(s.Height)
}

// ParseSize will parse a string in the form `<H>x<W>`
func ParseSize(s string) (Size, bool) {
	split := strings.Split(s, "x")
	if len(split) != 2 {
		return Size{}, false
	}
	height, hErr := strconv.ParseInt(split[0], 10, 32)
	width, wErr := strconv.ParseInt(split[1], 10, 32)
	if hErr != nil || wErr != nil {
		return Size{}, false
	}
	return Size{Height: int(height), Width: int(width)}, true
}

// Terminal is the datatype which models the terminal it's zero value is not usable and instead should be
// constructed via the functions provided:
//   - [NewTerminal]
//   - [NewFixedSizeTerminal]
//   - [NewParsedFixedSizeTerminal]
//   - [NewTestTerminal]
type Terminal struct {
	size      Size
	listeners []ConditionalListener
	fallbacks []Listener

	stdin  io.Reader
	stdout io.Writer

	stdinFd              int
	terminalSizeCallBack func() (Size, error)

	isTestTerminal bool
	isDynamicSize  bool

	// should be called if a panic occurs otherwise stacktraces are unreadable
	cleanup func()

	listenMutex *sync.Mutex
}

// NewTerminal creates a new terminal which immediately tries to verify that it can operate for
// [Terminal.StartRaw] in which it listens for key strokes, it also tries to read the terminal size from the
// environment on startup.
func NewTerminal() (*Terminal, error) {
	stdoutIsTerm := term.IsTerminal(int(os.Stdout.Fd()))
	stdErrIsTerm := term.IsTerminal(int(os.Stderr.Fd()))
	sizeErr := errors.Errorf("Not an expected terminal environment cannot get terminal size")
	if !(stdoutIsTerm || stdErrIsTerm) {
		return nil, sizeErr
	}
	size, err := getCurrentTerminalSize(os.Stdout)
	if err != nil {
		return nil, err
	}
	t := &Terminal{
		size:          size,
		listeners:     []ConditionalListener{},
		fallbacks:     []Listener{},
		stdin:         os.Stdin,
		stdout:        os.Stdout,
		listenMutex:   &sync.Mutex{},
		isDynamicSize: true,
		stdinFd:       int(os.Stdin.Fd()),
		terminalSizeCallBack: func() (Size, error) {
			if stdErrIsTerm {
				return getCurrentTerminalSize(os.Stderr)
			} else if stdoutIsTerm {
				return getCurrentTerminalSize(os.Stdout)
			} else {
				return Size{}, sizeErr
			}
		},
		isTestTerminal: false,
	}
	return t, t.supportsRaw(os.Stdin)
}

// NewFixedSizeTerminal creates a new terminal with a starting size, note that unless [Terminal.StartRaw] is
// called this terminal will not set up for listening to key strokes which is helpful for situations in which
// a real terminal environment is not setup.
func NewFixedSizeTerminal(s Size) (*Terminal, error) {
	t := &Terminal{
		size:          s,
		listeners:     []ConditionalListener{},
		fallbacks:     []Listener{},
		stdin:         os.Stdin,
		stdout:        os.Stdout,
		listenMutex:   &sync.Mutex{},
		isDynamicSize: false,
	}
	return t, nil
}

// NewParsedFixedSizeTerminal will construct a new fixed size terminal which cannot change size, parsing the
// size from the input parameter string, which is in format <H>x<W>, where H and W are integers.
func NewParsedFixedSizeTerminal(size string) (*Terminal, error) {
	s, ok := ParseSize(size)
	if !ok {
		return nil, errors.Errorf("Cannot parse %q as terminal a size, should be in the form \"<H>x<W>\", where H and W are integers.", size)
	}
	return NewFixedSizeTerminal(s)
}

// NewTestTerminal builds a terminal in which no real file interactions occur by default, instead all normal
// operations to stdout and stdin and performed on the two interfaces. And when the terminal is instructed to
// compute a new size with [Terminal.UpdateSize] it instead call the [terminalSizeCallBack] to retrieve a new
// size and store it. This is helpful for test environments so that the output of the terminal can be
// inspected and asserted on.
func NewTestTerminal(stdinReader io.Reader, stdoutWriter io.Writer, terminalSizeCallBack func() Size) (*Terminal, error) {
	size := terminalSizeCallBack()
	return &Terminal{
		size:                 size,
		listeners:            []ConditionalListener{},
		fallbacks:            []Listener{},
		stdin:                stdinReader,
		stdout:               stdoutWriter,
		terminalSizeCallBack: func() (Size, error) { return terminalSizeCallBack(), nil },
		isTestTerminal:       true,
		isDynamicSize:        true,
		listenMutex:          &sync.Mutex{},
	}, nil
}

// GetSize gets the cached size of the terminal, as in it returns the value most recently attained from
// [Terminal.UpdateSize], or if that has not been called the size of the terminal as initialised.
func (t *Terminal) GetSize() Size {
	return t.size
}

// UpdateSize the terminals stored size. Retrieve the result with [Terminal.GetSize].
func (t *Terminal) UpdateSize() error {
	if !t.isDynamicSize {
		return nil
	}
	var err error
	t.size, err = t.terminalSizeCallBack()
	return err
}

type Listener struct {
	// Name is used for if a listener errors for easier identification, it may be omitted.
	Name string
	// Action the callback which will be invoked when a user inputs the applicable rune, the rune passed is
	// the same rune passed to applicable. Note the terminal size will have been updated before this called,
	// but this is actually racey if the user is typing while changing size. If an error occurs in this action
	// the terminal will panic and exit.
	Action func(rune) error
}

type ConditionalListener struct {
	Listener
	// Applicable is the applicability of this listen, i.e. for which input runes do you want this action to
	// be fired
	Applicable func(rune) bool
}

type userControlCErr struct{}

// UserCancelled is the error returned by the terminal when ctrl-c is entered by the user, this stops the
// [Terminal.StartRaw] infinite loop.
var UserCancelled = userControlCErr{}

func (userControlCErr) Error() string {
	return "user cancelled"
}

// StartRaw takes ownership of the stdin/stdout and control of the incoming context. It will asynchronously
// block on the users input and forward characters to the relevant listener. By default a `ctrl+C` listener is
// added which will call the [stop] function when detected.
//
// The first return value is a clean up function which recover from a panic, putting the terminal back into
// normal mode and unhooking the listeners so that the program terminates gracefully upon a panic in another
// thread. It should be called like so:
//
//	term, _ := terminal.NewTerminal()
//	cleanup, _ := term.StartRaw(ctx, stop)
//	defer cleanup() // Graceful panic recovery
//	<-ctx.Done() // Wait till user cancels with ctrl+C
//
// To block a main thread until the `ctrl+C` listener is hit, simply wait on the input [ctx.Done()] channel.
//
// The `ctrl-c` listener will also provide the [terminal.UserControlCErr] cause when this happens for use with
// [error.Is].
func (t *Terminal) StartRaw(
	ctx context.Context,
	stop context.CancelCauseFunc,
	listeners []ConditionalListener,
	fallbacks []Listener,
) (func(), error) {
	restore := func() {}
	if !t.isTestTerminal {
		oldState, err := term.MakeRaw(t.stdinFd)
		if err != nil {
			return nil, errors.Wrap(err, "failed to set terminal to raw mode")
		}
		restore = func() { _ = term.Restore(t.stdinFd, oldState) }
	}
	ctrlCAction := func(rune) error {
		t.Print(ansi.ShowCursor)
		restore()
		stop(UserCancelled)
		return nil
	}
	t.cleanup = func() {
		_ = ctrlCAction('\x00')
	}

	controlCListener := ConditionalListener{
		Applicable: func(r rune) bool { return r == '\x03' },
		Listener: Listener{
			Name:   "ctrl+c",
			Action: ctrlCAction,
		},
	}
	t.listeners = slices.Concat(t.listeners, []ConditionalListener{controlCListener}, listeners)
	if fallbacks != nil {
		t.fallbacks = fallbacks
	}
	t.Print(ansi.HideCursor)
	go t.beingListening(ctx)
	return t.cleanup, nil
}

type ClearBehaviour int

const (
	UpdateSize            ClearBehaviour = 1 // Ensures the size is updated before the clear is called.
	MoveHome              ClearBehaviour = 2 // Move home will move the cursor back to the home position after the clear completes.
	UpdateSizeAndMoveHome ClearBehaviour = 3 // Does both.
)

// ClearScreen will "clear" the current terminal, I'm unsure exactly how this works in terms of terminal
// history and this is an area for improvement still. The parameter indicates what should happen after the
// terminal is cleared.
func (t *Terminal) ClearScreen(behaviour ClearBehaviour) error {
	if behaviour == UpdateSize || behaviour == UpdateSizeAndMoveHome {
		if err := t.UpdateSize(); err != nil {
			return errors.Wrap(err, "while ClearScreen")
		}
	}
	t.Print(strings.Repeat("\n", t.size.Height))
	err := t.Print(ansi.Clear)
	if behaviour == MoveHome || behaviour == UpdateSizeAndMoveHome {
		err = errors.Join(err, t.Print(ansi.Home))
	}
	return errors.Wrap(err, "while ClearScreen")
}

// Print will write the string [s] to the stdout controlled by the terminal, returning an error if that fails.
func (t *Terminal) Print(s string) error {
	err := utils.Err(t.Write([]byte(s)))
	return err
}

// Write will print the passed bytes to the stdout controlled by the terminal, returning the standard number
// of bytes written and error.
func (t *Terminal) Write(b []byte) (int, error) {
	return t.stdout.Write(b)
}

type listenResult struct {
	n   int
	err error
}

func (t *Terminal) beingListening(ctx context.Context) {
	buffer := make([]byte, 20)
	listenChannel := make(chan listenResult, 20)
	processingChannel := make(chan struct{})
	// Create a go-routine which continuously reads from stdin
	go func() {
		defer t.cleanup()
		// This is blocking hence why the go-routine wrapper exists, we still only free ourself when
		// the outer context is done which is racey.
		t.listen(ctx, listenChannel, processingChannel, buffer)
	}()

	defer t.cleanup()
	for {
		// Spin forever, waiting on input from the context which has cancelled us from else where, or a new
		// input char.
		select {
		case <-ctx.Done():
			return
		case received := <-listenChannel:
			if received.err != nil {
				panic(errors.Wrap(received.err, "unexpected read failure in terminal"))
			}
			if err := t.UpdateSize(); err != nil {
				panic(errors.Wrap(err, "unexpected read failure in terminal"))
			}
			if received.n <= 0 {
				return // cancelled
			}
			heard := string(buffer[:received.n])
			slog.Debug("got keyboard input", "received", heard)
			for _, r := range heard {
				t.processListenedRune(r)
			}
			// if we don't have the processing signal this clear would be racey against stdin.
			bytes.Clear(buffer, received.n)
			processingChannel <- struct{}{}
		}
	}
}

// processListenedRune should only be called by the listener thread
func (t *Terminal) processListenedRune(r rune) {
	runFallback := true
	for _, l := range t.listeners {
		if !l.Applicable(r) {
			continue
		}
		err := l.Action(r)
		if err != nil {
			panic(errors.Wrapf(err, "unexpected failure Action %q in terminal", l.Name))
		}
		runFallback = false
	}
	if runFallback {
		for _, l := range t.fallbacks {
			err := l.Action(r)
			if err != nil {
				panic(errors.Wrapf(err, "unexpected failure Action %q in terminal", l.Name))
			}
		}
	}
}

func (t *Terminal) listen(
	ctx context.Context,
	listenChannel chan listenResult,
	processingChannel chan struct{},
	buffer []byte,
) {
	defer close(listenChannel)
	for {
		select {
		case <-ctx.Done():
			return
		default:
			// We "listen" on the stdin waiting for user input.
			n, readErr := t.stdin.Read(buffer)
			listenChannel <- listenResult{n: n, err: readErr}
			// Now wait for the main loop to be done with that last read
			<-processingChannel
		}
	}
}

// getCurrentTerminalSize gets the current terminal size or error if the program doesn't have a terminal
// attached (e.g. go tests).
func getCurrentTerminalSize(file *os.File) (Size, error) {
	w, h, err := term.GetSize(int(file.Fd()))
	return Size{Height: h, Width: w}, errors.Wrap(err, "failed to get terminal size")
}

func (t *Terminal) supportsRaw(file *os.File) error {
	inFd := int(file.Fd())
	oldState, makeRawErr := term.MakeRaw(inFd)
	var restoreErr error
	if oldState != nil {
		restoreErr = term.Restore(inFd, oldState)
	}
	return errors.Wrap(errors.Join(makeRawErr, restoreErr), "failed to set terminal to raw mode")
}
