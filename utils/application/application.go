// Use of this source code is governed by a GPL-2 license that can be found in the LICENSE file.
//
// Copyright 2025 Lexer747
//
// SPDX-License-Identifier: GPL-2.0-only

package application

import (
	"io"
	"log/slog"
	"os"
	"runtime"
	"runtime/pprof"
	"runtime/trace"
	"strings"

	"github.com/Lexer747/acci-ping/gui/themes"
	"github.com/Lexer747/acci-ping/terminal"
	"github.com/Lexer747/acci-ping/utils/check"
	"github.com/Lexer747/acci-ping/utils/errors"
)

type BuildInfo struct {
	commit    string
	goVersion string
	branch    string
	timestamp string
	tag       string
}

//nolint:staticcheck
func MakeBuildInfo(COMMIT, GO_VERSION, BRANCH, TIMESTAMP, TAG string) *BuildInfo {
	if COMMIT == "" && GO_VERSION == "" && BRANCH == "" && TIMESTAMP == "" && TAG == "" {
		return nil
	}
	return &BuildInfo{
		commit:    COMMIT,
		goVersion: GO_VERSION,
		branch:    BRANCH,
		timestamp: TIMESTAMP,
		tag:       TAG,
	}
}

func InitLogging(file string, info *BuildInfo) (toDefer func()) {
	if file != "" {
		f, err := os.Create(file)
		check.NoErr(err, "could not create Log file")
		h := slog.NewTextHandler(f, &slog.HandlerOptions{
			Level: slog.LevelDebug,
		})
		logger := slog.New(h)
		if info != nil {
			logger = logger.With(
				"COMMIT", info.commit,
				"BRANCH", info.branch,
				"GO_VERSION", info.goVersion,
				"BUILD_TIMESTAMP", info.timestamp,
				"TAG", info.tag,
			)
		}
		slog.SetDefault(logger)
		slog.Debug("Logging started", "file", file)
		return func() {
			slog.Debug("Logging finished, closing", "file", file)
			check.NoErr(f.Close(), "failed to close log file")
		}
	}
	// If no file is specified we want to stop all logging
	h := slog.NewTextHandler(io.Discard, &slog.HandlerOptions{
		Level: slog.LevelError,
	})
	slog.SetDefault(slog.New(h))
	return func() {}
}

func InitCPUProfiling(cpuprofile string) (toDefer func()) {
	if cpuprofile == "" {
		return func() {}
	}
	cpuFile, err := os.Create(cpuprofile)
	check.NoErr(err, "could not create CPU profile")
	err = pprof.StartCPUProfile(cpuFile)
	check.NoErr(err, "could not start CPU profile")
	traceFile, err := os.Create("trace-" + cpuprofile)
	check.NoErr(err, "could not create Trace CPU profile")
	err = trace.Start(traceFile)
	check.NoErr(err, "could not start Trace CPU profile")

	slog.Debug("Started CPU & Trace profile", "path", cpuprofile)
	return func() {
		slog.Debug("Writing CPU profile", "path", cpuprofile)
		trace.Stop()
		pprof.StopCPUProfile()
		check.NoErr(cpuFile.Sync(), "failed to Sync profile")
		check.NoErr(cpuFile.Close(), "failed to close profile")
		check.NoErr(traceFile.Sync(), "failed to Sync profile")
		check.NoErr(traceFile.Close(), "failed to close profile")
	}
}

func InitMemProfile(memprofile string) (toDefer func()) {
	if memprofile == "" {
		return func() {}
	}
	f, err := os.Create(memprofile)
	check.NoErr(err, "could not create memory profile")

	doMemProfile := func() {
		slog.Debug("Writing memory profile")
		_, err := f.Seek(0, 0)
		check.NoErr(err, "could not truncate memory profile")
		_, err = f.Write([]byte{})
		check.NoErr(err, "could not truncate memory profile")
		runtime.GC() // get up-to-date statistics
		if err := pprof.WriteHeapProfile(f); err != nil {
			check.NoErr(err, "could not write memory profile")
		}
	}

	return func() {
		doMemProfile()
		f.Close()
	}
}

func LoadTheme(themeStr string, term *terminal.Terminal) error {
	theme, ok := themes.LookupTheme(themeStr)
	if !ok {
		colour, ok := term.BackgroundColour()
		// https://github.com/microsoft/terminal/issues/3718
		if !ok && strings.Contains(runtime.GOOS, "windows") {
			// best bet is too read the json profile
			// https://learn.microsoft.com/en-us/windows/terminal/install#settings-json-file (3 locations)
			windows, ok := tryParseWindowsTerminalSettings()
			if ok {
				colour = windows
			}
		}
		theme = themes.GetDefault(colour)
		var err error
		theme, err = loadFileTheme(themeStr, theme)
		if err != nil {
			return err
		}
	}
	themes.LoadTheme(theme)
	return nil
}

// LoadFileTheme takes the config and tries to treat the theme config value as a path from which to read all
// the values a theme needs. This lets the program support custom themes.
func loadFileTheme(themeStr string, fallBack themes.Theme) (t themes.Theme, err error) {
	f, err := os.Open(themeStr)
	defer func() {
		if f != nil {
			err = f.Close()
		}
	}()
	// We assume that no file passed means the default option, any other error implies that something went
	// wrong and the user may want to action this problem
	if err != nil && !errors.Is(err, os.ErrNotExist) {
		return fallBack, errors.Wrapf(err, "failed to load theme from path %q", themeStr)
	} else if err != nil && errors.Is(err, os.ErrNotExist) {
		// swallow this error and carry on
		return fallBack, nil
	}

	data, err := io.ReadAll(f)
	if err != nil {
		return fallBack, errors.Wrapf(err, "failed to load theme from path %q", themeStr)
	}

	loaded, err := themes.ParseThemeFromJSON(data)
	if err != nil {
		return fallBack, errors.Wrapf(err, "failed to load theme from path %q", themeStr)
	}
	return loaded, nil
}
