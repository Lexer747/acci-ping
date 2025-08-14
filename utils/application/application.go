// Use of this source code is governed by a GPL-2 license that can be found in the LICENSE file.
//
// Copyright 2025 Lexer747
//
// SPDX-License-Identifier: GPL-2.0-only

package application

import (
	"flag"
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

func (b BuildInfo) BuildTimestamp() string {
	return b.timestamp
}

func (b BuildInfo) GoVersion() string {
	return b.goVersion
}

func (b BuildInfo) Branch() string {
	return b.branch
}

func (b BuildInfo) Commit() string {
	return b.commit
}

func (b BuildInfo) Tag() string {
	return b.tag
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

// SharedFlags contains duplicated flags that ANY program may want to interact with. Set it up by calling
// [NewSharedFlags] then doing the normal flag parse. After this you can set up all the profiling by calling
// the public functions:
//   - [SharedFlags.InitLogging]
//   - [SharedFlags.InitCPUProfiling]
//   - [SharedFlags.InitMemProfile]
type SharedFlags struct {
	cpuprofile  *string
	debugStrict *bool
	logFile     *string
	helpDebug   *bool
	memprofile  *string
}

func NewSharedFlags(f *flag.FlagSet) *SharedFlags {
	return &SharedFlags{
		cpuprofile:  f.String("debug-cpuprofile", "", "write cpu profile to `file`"),
		debugStrict: f.Bool("debug-strict", false, "enables more strict operation in which warnings turn into crashes."),
		logFile:     f.String("debug-log", "", "write logs to `file`. (default no logs written)"),
		memprofile:  f.String("debug-memprofile", "", "write memory profile to `file`"),
		helpDebug:   f.Bool("help-debug", false, "prints all additional debug arguments"),
	}
}

func (sf *SharedFlags) InitLogging(info *BuildInfo) (toDefer func()) {
	if sf.logFile != nil && *sf.logFile != "" {
		file := *sf.logFile
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
	h := slog.DiscardHandler
	slog.SetDefault(slog.New(h))
	return func() {}
}

func (sf *SharedFlags) InitCPUProfiling() (toDefer func()) {
	if sf.cpuprofile == nil || *sf.cpuprofile == "" {
		return func() {}
	}
	file := *sf.cpuprofile
	cpuFile, err := os.Create(file)
	check.NoErr(err, "could not create CPU profile")
	err = pprof.StartCPUProfile(cpuFile)
	check.NoErr(err, "could not start CPU profile")
	traceFile, err := os.Create("trace-" + file)
	check.NoErr(err, "could not create Trace CPU profile")
	err = trace.Start(traceFile)
	check.NoErr(err, "could not start Trace CPU profile")

	slog.Debug("Started CPU & Trace profile", "path", file)
	return func() {
		slog.Debug("Writing CPU profile", "path", file)
		trace.Stop()
		pprof.StopCPUProfile()
		check.NoErr(cpuFile.Sync(), "failed to Sync profile")
		check.NoErr(cpuFile.Close(), "failed to close profile")
		check.NoErr(traceFile.Sync(), "failed to Sync profile")
		check.NoErr(traceFile.Close(), "failed to close profile")
	}
}

func (sf *SharedFlags) InitMemProfile() (toDefer func()) {
	if sf.memprofile == nil || *sf.memprofile == "" {
		return func() {}
	}
	file := *sf.memprofile
	f, err := os.Create(file)
	check.NoErr(err, "could not create memory profile")

	doMemProfile := func() {
		slog.Debug("Writing memory profile")
		_, err := f.Seek(0, 0)
		check.NoErr(err, "could not truncate memory profile")
		_, err = f.Write([]byte{})
		check.NoErr(err, "could not truncate memory profile")
		runtime.GC() // get up-to-date statistics
		err = pprof.WriteHeapProfile(f)
		if err != nil {
			check.NoErr(err, "could not write memory profile")
		}
	}

	return func() {
		doMemProfile()
		f.Close()
	}
}

func (sf *SharedFlags) Profiling() bool {
	if sf.cpuprofile == nil || sf.memprofile == nil {
		return false
	}
	return *sf.cpuprofile != "" || *sf.memprofile != ""
}

func (sf *SharedFlags) DebugStrict() bool {
	if sf.debugStrict == nil {
		return false
	}
	return *sf.debugStrict
}

func (sf *SharedFlags) HelpDebug() bool {
	if sf.helpDebug == nil {
		return false
	}
	return *sf.helpDebug
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
