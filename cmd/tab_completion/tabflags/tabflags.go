// Use of this source code is governed by a GPL-2 license that can be found in the LICENSE file.
//
// Copyright 2025 Lexer747
//
// SPDX-License-Identifier: GPL-2.0-only

package tabflags

import (
	"flag"
	"maps"
	"slices"
	"strings"

	"github.com/Lexer747/acci-ping/utils/iterutils"
)

type AutoComplete struct {
	Choices   []string
	WantsFile bool
	FileExt   string
}

type Flag struct {
	// Name is the actual name as it appears on the CLI without the dash.
	Name string
	// AutoComplete specifies if the flag has any auto complete options.
	AutoComplete *AutoComplete
}

// FlagSet is an extension of [flag.FlagSet] that enables auto completion providers for a command.
type FlagSet struct {
	*flag.FlagSet

	flags    []Flag
	nameToAc map[string]*AutoComplete

	wantsFile bool
	fileExt   string
}

// NewAutoCompleteFlagSet wraps a [flag.FlagSet] with autocomplete configuration, if [wantsFile] is set then
// it's expected that the overall command wants a file so autocomplete suggestions from the working dir will
// be given. If [fileExt] is set then only the files which match the extension are suggested.
func NewAutoCompleteFlagSet(f *flag.FlagSet, wantsFile bool, fileExt string) *FlagSet {
	return &FlagSet{
		FlagSet:   f,
		flags:     []Flag{},
		nameToAc:  map[string]*AutoComplete{},
		wantsFile: wantsFile,
		fileExt:   fileExt,
	}
}

// WantsFile returns true if this flag set wants a file as a free form arg.
func (f *FlagSet) WantsFile() bool {
	return f.wantsFile
}

// FileExt returns the specified file extensions to look for if [FlagSet.WantsFile] returns true.
func (f *FlagSet) FileExt() string {
	return f.fileExt
}

// GetAutoCompleteFor returns the autocomplete configuration for the given CLI flag. Returns the nil if the
// param isn't found or not autocomplete config exists.
func (f *FlagSet) GetAutoCompleteFor(flagName string) *AutoComplete {
	ac, ok := f.nameToAc[strings.TrimPrefix(flagName, "-")]
	if !ok {
		return nil
	}
	return ac
}

// GetNames returns all the names that this flag set knows about, skipping "-debug" prefix unless
// [includeDebug] is true. Also skips any names in [toSkip].
func (f *FlagSet) GetNames(includeDebug bool, toSkip []string) []string {
	keys := maps.Keys(f.nameToAc)
	withDash := iterutils.Map(keys, func(n string) string {
		return "-" + n
	})
	filtered := iterutils.Filter(withDash, func(n string) bool {
		if !includeDebug && strings.HasPrefix(n, "-debug") {
			return false
		}
		return !slices.Contains(toSkip, n)
	})
	names := slices.Collect(filtered)
	slices.Sort(names)
	return names
}

func (f *FlagSet) Has(flagName string) bool {
	_, has := f.nameToAc[strings.TrimPrefix(flagName, "-")]
	return has
}

// String see [FlagSet.String], also pass the autocomplete configuration.
//
// [FlagSet.String]: https://pkg.go.dev/flag#String
func (f *FlagSet) String(name string, value string, usage string, ac AutoComplete) *string {
	ret := f.FlagSet.String(name, value, usage)
	f.addAc(name, &ac)
	return ret
}

// Bool see [FlagSet.Bool]
//
// [FlagSet.Bool]: https://pkg.go.dev/flag#Bool
func (f *FlagSet) Bool(name string, value bool, usage string) *bool {
	ret := f.FlagSet.Bool(name, value, usage)
	f.addAc(name, nil)
	return ret
}

// Bool see [FlagSet.Float64]
//
// [FlagSet.Float64]: https://pkg.go.dev/flag#Float64
func (f *FlagSet) Float64(name string, value float64, usage string) *float64 {
	ret := f.FlagSet.Float64(name, value, usage)
	f.addAc(name, nil)
	return ret
}

// Bool see [FlagSet.Int]
//
// [FlagSet.Int]: https://pkg.go.dev/flag#Int
func (f *FlagSet) Int(name string, value int, usage string) *int {
	ret := f.FlagSet.Int(name, value, usage)
	f.addAc(name, nil)
	return ret
}

func (f *FlagSet) addAc(name string, ac *AutoComplete) {
	flag := Flag{
		Name:         name,
		AutoComplete: ac,
	}
	f.flags = append(f.flags, flag)
	f.nameToAc[name] = ac
}
