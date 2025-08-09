// Use of this source code is governed by a GPL-2 license that can be found in the LICENSE file.
//
// Copyright 2025 Lexer747
//
// SPDX-License-Identifier: GPL-2.0-only

package themes

import "log/slog"

// Theme contains all the data needed in order to print colour outputs to the terminal.
//
// In general you probably don't want to use a theme type directly as this would require passing it through
// all functions as a param instead there's a global theme provided by the package which be accessed through
// the top level functions:
//
//   - [Primary] is the main font, this should have maximum contrast against the
//     background of the terminal, e.g. if the background is white, primary will
//     colour the text black.
//   - [Secondary] is the alternate default colour, usually a softer version of
//     the primary font and generally will have a lower contrast.
//   - [Highlight] is a bold bright colour. Usually yellowish.
//   - [Emphasis] is another bold bright colour. Usually cyanish.
//   - [TitleHighlight] is an exclusive colour for titles. Usually magentaish.
//   - [Positive] and [DarkPositive] are happy path colours indicating some success.
//   - [Negative] and [DarkNegative] are sad path colours indicating some failure or
//     "bad" values.
//
// This theme should be [LoadTheme] at application startup and any global variables which rely on the theme
// should be re-initialised after the loading has completed. Otherwise all dynamic values which call the top
// level function will use the loaded theme.
type Theme struct {
	// internal details: the easiest way to implement this may have been an interface over these methods
	// instead but then we pay an interface deref and potential re-parsing of the input data. Since this will
	// be a very hot part of the graph drawing we don't want to pay this cost more than once. Hence this
	// [Theme] struct represents the final "no more parsing needed" state of a theme, in which the only work
	// which is done to colour text is a single string concatenation.
	//
	// see [themeJSON] for how the data is stored.

	primary        colourImpl
	secondary      colourImpl
	highlight      colourImpl
	emphasis       colourImpl
	titleHighlight colourImpl
	positive       colourImpl
	darkPositive   colourImpl
	negative       colourImpl
	darkNegative   colourImpl
}

// Primary is the main font, this should have maximum contrast against the background of the terminal, e.g. if
// the background is white, primary will colour the text black.
func Primary(s string) string { return globalTheme.primary.Do(s) }

// Secondary is the alternate default colour, usually a softer version of the primary font and generally will
// have a lower contrast.
func Secondary(s string) string { return globalTheme.secondary.Do(s) }

// Highlight is a bold bright colour. Usually yellowish.
func Highlight(s string) string { return globalTheme.highlight.Do(s) }

// Emphasis is another bold bright colour. Usually cyanish.
func Emphasis(s string) string { return globalTheme.emphasis.Do(s) }

// TitleHighlight is an exclusive colour for titles. Usually magentaish.
func TitleHighlight(s string) string { return globalTheme.titleHighlight.Do(s) }

// Positive and DarkPositive are happy path colours indicating some success.
func Positive(s string) string { return globalTheme.positive.Do(s) }

// Positive and DarkPositive are happy path colours indicating some success.
func DarkPositive(s string) string { return globalTheme.darkPositive.Do(s) }

// Negative and DarkNegative are sad path colours indicating some failure or "bad" values.
func Negative(s string) string { return globalTheme.negative.Do(s) }

// Negative and DarkNegative are sad path colours indicating some failure or "bad" values.
func DarkNegative(s string) string { return globalTheme.darkNegative.Do(s) }

// LookupTheme will search through the built-in themes and return one if found, or false.
func LookupTheme(name string) (Theme, bool) {
	theme, ok := builtInsLookup[normalizeName(name)]
	return theme, ok
}

// GetDefault returns the theme which should used based on a given luminance value, to set this as the global
// theme use [LoadTheme].
func GetDefault(luminance Luminance) Theme {
	if luminance.IsDark() {
		return DarkTheme
	} else {
		return LightTheme
	}
}

// GetLoaded returns the currently loaded theme
func GetLoaded() Theme {
	return globalTheme
}

// LoadTheme updates the global theme with this new theme [t].
func LoadTheme(t Theme) {
	slog.Info("loaded new theme", "theme", t)
	globalTheme = t
}

func (t Theme) String() string {
	return "{" +
		"primary: " + t.primary.String() + " " +
		"secondary: " + t.secondary.String() + " " +
		"highlight: " + t.highlight.String() + " " +
		"emphasis: " + t.emphasis.String() + " " +
		"titleHighlight: " + t.titleHighlight.String() + " " +
		"positive: " + t.positive.String() + " " +
		"darkPositive: " + t.darkPositive.String() + " " +
		"negative: " + t.negative.String() + " " +
		"darkNegative: " + t.darkNegative.String() +
		"}"
}

var globalTheme Theme
