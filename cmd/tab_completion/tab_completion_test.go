// Use of this source code is governed by a GPL-2 license that can be found in the LICENSE file.
//
// Copyright 2025 Lexer747
//
// SPDX-License-Identifier: GPL-2.0-only

//nolint:testpackage
package tabcompletion

import (
	"flag"
	"log"
	"os"
	"path/filepath"
	"slices"
	"strings"
	"testing"

	"github.com/Lexer747/acci-ping/cmd/subcommands/drawframe"
	"github.com/Lexer747/acci-ping/cmd/tab_completion/tabflags"
	"github.com/Lexer747/acci-ping/gui/themes"
	"github.com/Lexer747/acci-ping/utils/application"
	"github.com/Lexer747/acci-ping/utils/sliceutils"
	"gotest.tools/v3/assert"
)

var accipingFlags = make_acciping_Flags()
var subCommands = []Command{
	make_drawframe_Flags(),
	make_version_Flags(),
}

func TestGetChoices(t *testing.T) {
	t.Parallel()
	t.Run("tab", func(t *testing.T) {
		t.Parallel()
		actual, err := runGetChoices("acci-ping")
		assert.NilError(t, err)

		expectedFlags := []string{}
		accipingFlags.Fs.VisitAll(func(f *flag.Flag) {
			if strings.HasPrefix(f.Name, "debug") {
				return
			}
			expectedFlags = append(expectedFlags, "-"+f.Name)
		})

		assertEqual(t, actual, slices.Concat([]string{"drawframe", "version"}, expectedFlags))
	})
	t.Run("start drawframe", func(t *testing.T) {
		t.Parallel()
		actual, err := runGetChoices("acci-ping", "d")
		assert.NilError(t, err)
		assertEqual(t, actual, []string{"drawframe"})
	})
	t.Run("start drawframe 2", func(t *testing.T) {
		t.Parallel()
		actual, err := runGetChoices("acci-ping", "draw")
		assert.NilError(t, err)
		assertEqual(t, actual, []string{"drawframe"})
	})
	t.Run("start version", func(t *testing.T) {
		t.Parallel()
		actual, err := runGetChoices("acci-ping", "v")
		assert.NilError(t, err)
		assertEqual(t, actual, []string{"version"})
	})
	t.Run("start -", func(t *testing.T) {
		t.Parallel()
		actual, err := runGetChoices("acci-ping", "-")
		assert.NilError(t, err)

		expectedFlags := []string{}
		accipingFlags.Fs.VisitAll(func(f *flag.Flag) {
			if strings.HasPrefix(f.Name, "debug") {
				return
			}
			expectedFlags = append(expectedFlags, "-"+f.Name)
		})
		assertEqual(t, actual, expectedFlags)
	})
	t.Run("start drawframe ", func(t *testing.T) {
		t.Parallel()
		actual, err := runGetChoices("acci-ping", "drawframe", "")
		assert.NilError(t, err)

		expectedFlags := []string{}
		drawframe.GetFlags(nil).VisitAll(func(f *flag.Flag) {
			if strings.HasPrefix(f.Name, "debug") {
				return
			}
			expectedFlags = append(expectedFlags, "-"+f.Name)
		})
		expectedFlags = slices.Concat(expectedFlags, filesByExt(".go"))
		assertEqual(t, actual, expectedFlags)
	})
	t.Run("start drawframe -t", func(t *testing.T) {
		t.Parallel()
		actual, err := runGetChoices("acci-ping", "drawframe", "-t")
		assert.NilError(t, err)

		expectedFlags := []string{}
		drawframe.GetFlags(nil).VisitAll(func(f *flag.Flag) {
			if !strings.HasPrefix(f.Name, "t") {
				return
			}
			expectedFlags = append(expectedFlags, "-"+f.Name)
		})
		assert.DeepEqual(t, actual, expectedFlags)
	})
	t.Run("start <random>", func(t *testing.T) {
		t.Parallel()
		starting := []string{}
		accipingFlags.Fs.VisitAll(func(f *flag.Flag) {
			if strings.HasPrefix(f.Name, "debug") {
				return
			}
			starting = append(starting, "-"+f.Name)
		})
		start := sliceutils.TakeRandom(starting)
		expectedFlags := sliceutils.Remove(starting, start)

		actual, err := runGetChoices("acci-ping", start, "")
		assert.NilError(t, err)
		assertEqual(t, actual, expectedFlags)
	})
	t.Run("start -file", func(t *testing.T) {
		t.Parallel()
		expectedFlags := []string{}
		accipingFlags.Fs.VisitAll(func(f *flag.Flag) {
			if strings.HasPrefix(f.Name, "debug") || f.Name == "file" {
				return
			}
			expectedFlags = append(expectedFlags, "-"+f.Name)
		})
		expectedFlags = slices.Concat(expectedFlags, filesByExt(".go"))

		actual, err := runGetChoices("acci-ping", "-file", "")
		assert.NilError(t, err)

		assertEqual(t, actual, expectedFlags)
	})
	t.Run("start -file foobar.txt <random>", func(t *testing.T) {
		t.Parallel()
		starting := []string{}
		accipingFlags.Fs.VisitAll(func(f *flag.Flag) {
			if strings.HasPrefix(f.Name, "debug") {
				return
			}
			starting = append(starting, "-"+f.Name)
		})
		randomChoices := []string{}
		accipingFlags.Fs.VisitAll(func(f *flag.Flag) {
			_, ok := f.Value.(boolFlag)
			if strings.HasPrefix(f.Name, "debug") || !ok {
				return
			}
			randomChoices = append(randomChoices, "-"+f.Name)
		})
		start := sliceutils.TakeRandom(randomChoices)
		expectedFlags := sliceutils.Remove(starting, start, "-file")

		actual, err := runGetChoices("acci-ping", "-file", "foobar.txt", start, "")
		assert.NilError(t, err)

		assertEqual(t, actual, expectedFlags)
	})
	t.Run("start -theme", func(t *testing.T) {
		t.Parallel()
		expectedFlags := slices.Concat(themes.GetBuiltInNames(), filesByExt(""))

		actual, err := runGetChoices("acci-ping", "-theme", "")
		assert.NilError(t, err)

		assertEqual(t, actual, expectedFlags)
	})
	t.Run("start -theme c", func(t *testing.T) {
		t.Parallel()
		expectedFlags := slices.Concat(themes.GetBuiltInNames(), filesByExt(""))
		expectedFlags = sliceutils.Filter(expectedFlags, func(s string) bool { return strings.HasPrefix(s, "c") })

		actual, err := runGetChoices("acci-ping", "-theme", "c")
		assert.NilError(t, err)

		assertEqual(t, actual, expectedFlags)
	})
}

func runGetChoices(args ...string) ([]string, error) {
	return getChoices(max(1, len(args)-1), args, accipingFlags, subCommands)
}

type boolFlag interface {
	flag.Value
	IsBoolFlag() bool
}

//nolint:staticcheck
func make_acciping_Flags() Command {
	f := flag.NewFlagSet("", flag.ContinueOnError)
	tf := tabflags.NewAutoCompleteFlagSet(f, false, "")
	_ = application.NewSharedFlags(tf)

	_ = tf.String("file", "", "skipped for test",
		tabflags.AutoComplete{WantsFile: true, FileExt: ".go"})
	_ = tf.Bool("hide-help", false, "skipped for test")
	_ = tf.Float64("pings-per-minute", 60.0, "skipped for test")
	_ = tf.Bool("debug-error-creator", false, "skipped for test")
	_ = tf.String("url", "www.google.com", "skipped for test", tabflags.AutoComplete{})
	_ = tf.String("theme", "", "skipped for test",
		tabflags.AutoComplete{Choices: themes.GetBuiltInNames(), WantsFile: true})
	_ = tf.String("debug-term-size", "", "skipped for test", tabflags.AutoComplete{Choices: []string{"15x80", "20x85", "HxW"}})
	_ = tf.Bool("follow", false, "skipped for test")
	_ = tf.Int("debug-fps", 240, "skipped for test")
	_ = tf.Bool("logarithmic", false, "skipped for test")
	return Command{Cmd: "acci-ping", Fs: tf}
}

//nolint:staticcheck
func make_drawframe_Flags() Command {
	f := flag.NewFlagSet("", flag.ContinueOnError)
	tf := tabflags.NewAutoCompleteFlagSet(f, true, ".go")
	_ = application.NewSharedFlags(tf)
	_ = tf.Bool("debug-follow", false, "skipped for test")
	_ = tf.String("term-size", "", "skipped for test",
		tabflags.AutoComplete{Choices: []string{"15x80", "20x85", "HxW"}})
	_ = tf.String("theme", "", "skipped for test",
		tabflags.AutoComplete{Choices: themes.GetBuiltInNames(), WantsFile: true})
	_ = tf.Bool("log-scale", false, "skipped for test")
	return Command{Cmd: "drawframe", Fs: tf}
}

//nolint:staticcheck
func make_version_Flags() Command {
	return Command{Cmd: "version", Fs: nil}
}

func filesByExt(ext string) []string {
	entries, err := os.ReadDir("./")
	if err != nil {
		log.Fatal(err)
	}
	files := sliceutils.Map(entries, func(d os.DirEntry) string { return d.Name() })
	return sliceutils.Filter(files, func(f string) bool { return ext == "" || filepath.Ext(f) == ext })
}

func assertEqual(t *testing.T, expected, actual []string) {
	t.Helper()
	slices.Sort(actual)
	slices.Sort(expected)
	assert.DeepEqual(t, actual, expected)
}
