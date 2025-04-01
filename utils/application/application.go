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

	"github.com/Lexer747/acci-ping/utils/check"
)

type BuildInfo struct {
	commit    string
	goVersion string
	branch    string
	timestamp string
}

//nolint:staticcheck
func MakeBuildInfo(COMMIT, GO_VERSION, BRANCH, TIMESTAMP string) *BuildInfo {
	if COMMIT == "" && GO_VERSION == "" && BRANCH == "" && TIMESTAMP == "" {
		return nil
	}
	return &BuildInfo{
		commit:    COMMIT,
		goVersion: GO_VERSION,
		branch:    BRANCH,
		timestamp: TIMESTAMP,
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
	f, err := os.Create(cpuprofile)
	check.NoErr(err, "could not create CPU profile")
	err = pprof.StartCPUProfile(f)
	check.NoErr(err, "could not start CPU profile")
	slog.Debug("Started CPU profile", "path", cpuprofile)
	return func() {
		slog.Debug("Writing CPU profile", "path", cpuprofile)
		pprof.StopCPUProfile()
		check.NoErr(f.Sync(), "failed to Sync profile")
		check.NoErr(f.Close(), "failed to close profile")
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
