// Use of this source code is governed by a GPL-2 license that can be found in the LICENSE file.
//
// Copyright 2025 Lexer747
//
// SPDX-License-Identifier: GPL-2.0-only

//nolint:staticcheck
package tabcompletion

import (
	"flag"
	"fmt"
	"log/slog"
	"os"
	"path/filepath"
	"slices"
	"strconv"
	"strings"

	"github.com/Lexer747/acci-ping/cmd/tab_completion/tabflags"
	"github.com/Lexer747/acci-ping/utils/check"
	"github.com/Lexer747/acci-ping/utils/errors"
	"github.com/Lexer747/acci-ping/utils/exit"
	"github.com/Lexer747/acci-ping/utils/sliceutils"
)

const debug = false

const AutoCompleteString = "b2827fc8fc8c8267cb15f5a925de7e4712aa04ef2fbd43458326b595d66a36d9" // sha256 of `autoCompleteString`

type Command struct {
	Cmd string
	Fs  *tabflags.FlagSet
}

// Run will for a given command line input, write to stdout the autocomplete suggestion for the input [base]
// and subcommands.
func Run(args []string, base Command, subCommands []Command) {
	if debug {
		cleanup := makeLogFile()
		defer cleanup()
	} else {
		h := slog.DiscardHandler
		slog.SetDefault(slog.New(h))
	}
	// Catch panic's in the case of a mistake in this algorithm simply return no autocomplete suggestions.
	defer func() {
		if err := recover(); err != nil {
			slog.Error("caught panic", "err", err)
			tabExitFailure()
		}
	}()
	slog.Debug("auto complete called with", "args", args)

	// The format of tab completion is quite simple, bash expects the called program (in this case this
	// program) to return a space separated list of possible continuations. How this list is constructed is up
	// to the program.

	// The available args to this program are agreed by the stub script provided by this repo:
	// cmd/tab-completion/acci-ping, if these are not provided this sub command won't be called and no auto
	// completion provided.

	args = args[2:] // the first arg will always be the program name, second arg the auto-complete string

	// reslice back to the original inputs
	_COMP_CWORD, _COMP_LINE, _COMP_WORDS := parseCOMP(args)

	index, err := strconv.Atoi(_COMP_CWORD)
	if err != nil {
		tabExitFailure()
	}

	slog.Debug("inputs", "index", index, "_COMP_CWORD", _COMP_CWORD, "_COMP_LINE", _COMP_LINE, "_COMP_WORDS", _COMP_WORDS)
	choices, err := getChoices(index, _COMP_LINE, base, subCommands)
	slog.Debug("results", "choices", choices, "err", err)
	if err != nil {
		tabExitFailure()
		return
	}
	tabExit(choices)
}

// getChoices is the workhorse of autocomplete, it's more functional in style in that it should not output
// anything to stdout and instead returns the slice of autocomplete suggestions or error if something went
// wrong. Note that it can have side effects if an autocomplete wants a file as input to it's flag and
// therefore may read the current files in the working dir.
func getChoices(index int, _COMP_LINE []string, base Command, cmds []Command) ([]string, error) {
	slog.Debug("info", "len(_COMP_LINE)", len(_COMP_LINE))
	if index == 1 && len(_COMP_LINE) <= 1 {
		// Edge case, when we have exactly one index we should follow up with all the flags for the "base" command
		// as well as the names of the subcommands.
		return flagsAndSubcommands(cmds, getFlag(base.Fs.FlagSet)), nil
	}
	if index == 1 {
		// Edge case, same as above but there is some starting words being entered by the user so use that to
		// restrict the output
		return flagsAndSubcommandsWithPrefix(index, _COMP_LINE, cmds, getFlag(base.Fs.FlagSet)), nil
	}
	if index > 1 {
		// determine if we have a subcommand selected or are embracing the base and selecting flags
		isBase := strings.HasPrefix(_COMP_LINE[1], "-")
		var flags *tabflags.FlagSet
		var wantsFile bool
		var curCommand, fileExt string
		if isBase {
			flags = base.Fs
			wantsFile = base.Fs.WantsFile()
			curCommand = base.Cmd
			fileExt = base.Fs.FileExt()
		} else {
			sub := _COMP_LINE[1]
			choice := sliceutils.Filter(cmds, func(sc Command) bool { return sc.Cmd == sub })
			if len(choice) != 1 {
				return nil, errors.Errorf("No sub command found for %q", sub)
			}
			cmd := choice[0]
			flags = cmd.Fs
			wantsFile = cmd.Fs.WantsFile()
			curCommand = cmd.Cmd
			fileExt = cmd.Fs.FileExt()
		}
		cur := _COMP_LINE[len(_COMP_LINE)-1]
		return suggestionAutoComplete(index, _COMP_LINE, flags, wantsFile, fileExt, cur, curCommand), nil
	}
	slog.Error("unable to solve")
	return nil, errors.Errorf("unexpected inputs")
}

func suggestionAutoComplete(
	index int,
	_COMP_LINE []string,
	flags *tabflags.FlagSet,
	wantsFile bool,
	fileExt, cur, curCommand string,
) []string {
	alreadySet := _COMP_LINE[1:index]
	var files []string
	if wantsFile {
		files = getWorkingDirFiles()
		if fileExt != "" {
			files = sliceutils.Filter(files, func(path string) bool { return filepath.Ext(path) == fileExt })
		}
	}

	prev := _COMP_LINE[index-1]
	ac := flags.GetAutoCompleteFor(prev)
	if ac == nil {
		includeDebug := strings.HasPrefix(cur, "-d")
		names := flags.GetNames(includeDebug, alreadySet)
		return filterByPrefix(slices.Concat(names, files), cur)
	}

	// suggest continuation
	if prev == curCommand {
		return returnFlagSetNames(cur, flags, files, alreadySet)
	}
	toSuggest := *ac
	if toSuggest.WantsFile {
		files = getWorkingDirFiles()
		if toSuggest.FileExt != "" {
			files = sliceutils.Filter(files, func(path string) bool { return filepath.Ext(path) == toSuggest.FileExt })
		}
	}
	if len(toSuggest.Choices) == 0 {
		return returnFlagSetNames(cur, flags, files, alreadySet)
	}
	return filterByPrefix(slices.Concat(toSuggest.Choices, files), cur)
}

func getWorkingDirFiles() (files []string) {
	f, err := os.Open("./")
	if err != nil {
		slog.Error("failed to get files", "err", err)
	}
	files, err = f.Readdirnames(0)
	if err != nil {
		slog.Error("failed to get files", "err", err)
	}
	err = f.Close()
	if err != nil {
		slog.Error("failed to close dir", "err", err)
	}
	return files
}

func returnFlagSetNames(cur string, flags *tabflags.FlagSet, files, alreadySet []string) []string {
	includeDebug := strings.HasPrefix(cur, "-d")
	names := flags.GetNames(includeDebug, alreadySet)
	return filterByPrefix(slices.Concat(names, files), cur)
}

func flagsAndSubcommands(cmds []Command, flags []string) []string {
	subCommandNames := sliceutils.Map(cmds, func(s Command) string { return s.Cmd })
	return slices.Concat(subCommandNames, removeByPrefix(flags, "-debug"))
}

func flagsAndSubcommandsWithPrefix(index int, _COMP_LINE []string, cmds []Command, flags []string) []string {
	prefix := _COMP_LINE[index]
	includeDebug := strings.HasPrefix(prefix, "-d")
	subCommandNames := sliceutils.Map(cmds, func(s Command) string { return s.Cmd })
	if !includeDebug {
		flags = removeByPrefix(flags, "-debug")
	}
	toSearch := slices.Concat(subCommandNames, flags)
	return filterByPrefix(toSearch, prefix)
}

func filterByPrefix(options []string, prefix string) []string {
	return sliceutils.Filter(options, func(s string) bool { return strings.HasPrefix(s, prefix) })
}

func removeByPrefix(options []string, prefix string) []string {
	return sliceutils.Filter(options, func(s string) bool { return !strings.HasPrefix(s, prefix) })
}

func getFlag(fs *flag.FlagSet) []string {
	var result []string
	fs.VisitAll(func(f *flag.Flag) {
		result = append(result, "-"+f.Name)
	})
	return result
}

// parseCOMP relies on the assumption that the input from the arguments to this function has been formed by
// cmd/tab_completion/acci-ping, which is placed in the /etc/bash_completion.d/acci-ping or whatever shell the
// user is using.
func parseCOMP(args []string) (_COMP_CWORD string, _COMP_LINE []string, _COMP_WORDS []string) {
	state := 0
	prev := 0
	// Do a single linear search, each new [AutoCompleteString] will be separating each of the normal
	// arguments.
	for i, arg := range args {
		if arg == AutoCompleteString && state == 0 {
			_COMP_CWORD = args[prev:i][0]
			state = 1
			prev = i
		}
		if arg == AutoCompleteString && state == 1 {
			_COMP_WORDS = args[prev:i]
			state = 2
		}
		if arg == AutoCompleteString && state == 2 {
			_COMP_LINE = args[i+1:]
			if len(_COMP_LINE) == 1 {
				_COMP_LINE = strings.Split(_COMP_LINE[0], " ")
			}
		}
	}
	return _COMP_CWORD, _COMP_LINE, _COMP_WORDS
}

func tabExit(results []string) {
	fmt.Fprint(os.Stdout, strings.Join(results, " "))
	exit.Success()
}
func tabExitFailure() {
	tabExit([]string{})
}

func makeLogFile() (toDefer func()) {
	f, err := os.Create("/home/lexer747/repos/acci-ping/testing.log")
	check.NoErr(err, "could not create Log file")
	h := slog.NewTextHandler(f, &slog.HandlerOptions{
		Level: slog.LevelDebug,
	})
	logger := slog.New(h)
	slog.SetDefault(logger)
	slog.Debug("Logging started")
	return func() {
		slog.Debug("Logging finished, closing")
		check.NoErr(f.Close(), "failed to close log file")
	}
}
