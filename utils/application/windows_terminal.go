// Use of this source code is governed by a GPL-2 license that can be found in the LICENSE file.
//
// Copyright 2025 Lexer747
//
// SPDX-License-Identifier: GPL-2.0-only

package application

import (
	_ "embed"
	"encoding/json"
	"io"
	"log/slog"
	"os"
	"path/filepath"
	"slices"

	"github.com/Lexer747/acci-ping/gui/themes"
	"github.com/Lexer747/acci-ping/utils/check"
)

func tryParseWindowsTerminalSettings() (themes.Luminance, bool) {
	log := slog.With("trying to parse windows terminal settings", "entry")
	appdata, ok := os.LookupEnv("LOCALAPPDATA")
	if !ok {
		log.Info("failed to get LOCALAPPDATA")
		return -1, false
	}
	// TODO use:
	//	parent, err := os.FindProcess(os.Getppid())
	// in combo with `tasklist /FI "PID eq 22680" /FO list` to get which terminal is currently being run

	// TODO powershell ran outside terminal?

	for _, path := range windowsTerminalPaths {
		settingsPath := filepath.Join(appdata, path)
		f, err := os.Open(settingsPath)
		if err != nil {
			log.Info("skipping settings", "settings path", settingsPath, "file err", err)
			continue
		}
		defer func() { _ = f.Close() }()
		data, err := io.ReadAll(f)
		if err != nil {
			log.Info("skipping settings", "settings path", settingsPath, "file err", err)
			continue
		}
		settings := windowsTerminalSettings{}
		err = json.Unmarshal(data, &settings)
		if err != nil {
			log.Info("failed to unmarshal settings", "settings path", settingsPath, "json err", err)
			continue
		}
		toFind := settings.Profiles.Defaults.Name
		var found *scheme
		schemes := slices.Concat(settings.Schemes, windowsBuiltins)
		for _, scheme := range schemes {
			if scheme.Name == toFind {
				found = &scheme
				break
			}
		}
		if found == nil {
			log.Info("failed to find settings with name", "settings path", settingsPath, "name", toFind, "schemes", schemes)
			continue
		} else {
			luminance, err := themes.ParseRGB_CSSString(found.Background)
			if err != nil {
				log.Info("failed to unmarshal settings", "settings path", settingsPath, "json err", err)
				continue
			}
			log.Info("Succeed at reading settings", "settings path", settingsPath, "luminance", luminance)
			return luminance, true
		}
	}
	return -1, false
}

// https://learn.microsoft.com/en-us/windows/terminal/install#settings-json-file
var windowsTerminalPaths = []string{
	`\Packages\Microsoft.WindowsTerminal_8wekyb3d8bbwe\LocalState\settings.json`,        // Terminal (stable / general release)
	`\Packages\Microsoft.WindowsTerminalPreview_8wekyb3d8bbwe\LocalState\settings.json`, // Terminal (preview release)
	`\Microsoft\Windows Terminal\settings.json`,                                         // Terminal (unpackaged: Scoop, Chocolatey, etc)
}

type windowsTerminalSettings struct {
	Profiles profiles `json:"profiles"`
	Schemes  []scheme `json:"schemes"`
}

type profiles struct {
	Defaults colourScheme `json:"defaults"`
}

type colourScheme struct {
	Name string `json:"colorScheme"`
}

type scheme struct {
	Name string `json:"name"`

	Background string `json:"background"`
}

//go:embed windows_builtins/builtins.json
var windowsBuiltinsData []byte

var windowsBuiltins = func() []scheme {
	t := windowsTerminalSettings{}
	err := json.Unmarshal(windowsBuiltinsData, &t)
	check.NoErr(err, "failed to compile time init microsoft data")
	return t.Schemes
}()
