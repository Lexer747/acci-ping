// Use of this source code is governed by a GPL-2 license that can be found in the LICENSE file.
//
// Copyright 2025 Lexer747
//
// SPDX-License-Identifier: GPL-2.0-only

package themes

import (
	"cmp"
	_ "embed"
	"encoding/json"
	"slices"
	"strconv"
	"strings"

	"github.com/Lexer747/acci-ping/terminal/ansi"
	"github.com/Lexer747/acci-ping/terminal/typography"
	"github.com/Lexer747/acci-ping/utils/check"
	"github.com/Lexer747/acci-ping/utils/errors"
	"github.com/Lexer747/acci-ping/utils/sliceutils"
)

// ParseThemeFromJSON takes bytes (from a file) and returns a theme if one could be parsed or error.
func ParseThemeFromJSON(data []byte) (Theme, error) {
	t := themeJSON{}
	err := json.Unmarshal(data, &t)
	if err != nil {
		return Theme{}, errors.Wrap(err, "failed to parse theme")
	}
	return t.Theme()
}

// DescribeBuiltins gives a detailed slice of strings, where each string represents a builtin theme and it's
// colour palate.
func DescribeBuiltins() []string {
	slices.SortFunc(builtIns, func(a, b themeJSON) int { return cmp.Compare(a.Name, b.Name) })
	return sliceutils.Map(builtIns, func(t themeJSON) string {
		theme := check.Must(t.Theme())
		return "\t- " + theme.titleHighlight.Do(t.Name) + " | Primary:" + theme.primary.Do(typography.Block) +
			" Secondary:" + theme.secondary.Do(typography.Block) +
			" Emphasis:" + theme.emphasis.Do(typography.Block) +
			" Highlight:" + theme.highlight.Do(typography.Block) +
			" TitleHighlight:" + theme.titleHighlight.Do(typography.Block) +
			" Positive:" + theme.positive.Do(typography.Block) +
			" Negative:" + theme.negative.Do(typography.Block)
	})
}

var builtIns = []themeJSON{}
var builtInsLookup = map[string]Theme{}

func normalizeName(s string) string {
	return strings.ToLower(strings.Trim(s, " "))
}

func embedTheme(data []byte, otherNames ...string) Theme {
	t := themeJSON{}
	err := json.Unmarshal(data, &t)
	check.NoErr(err, "compile time theme failed")
	builtIns = append(builtIns, t)
	theme := check.Must(t.Theme())
	builtInsLookup[normalizeName(t.Name)] = theme
	for _, names := range otherNames {
		builtInsLookup[normalizeName(names)] = theme
	}
	return theme
}

//go:embed builtins/dark.json
var darkBytes []byte
var DarkTheme = embedTheme(darkBytes, "d")

//go:embed builtins/light.json
var lightBytes []byte
var LightTheme = embedTheme(lightBytes, "l")

//go:embed builtins/complex.json
var complexBytes []byte
var ComplexTheme = embedTheme(complexBytes, "c")

//go:embed builtins/no-theme.json
var noThemeBytes []byte
var NoTheme = embedTheme(noThemeBytes, "notheme", "no")

// themeJSON is the abstraction that allows us to store themes as json objects. The main feature this is
// currently providing is that terminals support various forms of 3-4 bit, 8 bit and 24 bit and this captures
// these details and knows to convert all them to the uniform [Theme] type for efficient drawing.
type themeJSON struct {
	Name    string      `json:"name"`
	Version string      `json:"version"` // Version is an internal detail and present incase the json format ever needs to change
	Colours coloursJSON `json:"colours"`
}

// TODO maybe a custom marshaller here to batch all colour errors at once

type coloursJSON struct {
	Primary        colourJSON `json:"primary"`
	Secondary      colourJSON `json:"secondary"`
	Highlight      colourJSON `json:"highlight"`
	Emphasis       colourJSON `json:"emphasis"`
	TitleHighlight colourJSON `json:"title-highlight"`
	Positive       colourJSON `json:"positive"`
	DarkPositive   colourJSON `json:"dark-positive"`
	Negative       colourJSON `json:"negative"`
	DarkNegative   colourJSON `json:"dark-negative"`
}

// Bit support depth is currently determined by this type and it's [UnmarshalJSON] function in combination
// with the [ansi] package.
type colourJSON struct {
	bit4     *Colour
	bit8     *int
	bit24    *bit24JSON
	noColour bool
}

func (cj *colourJSON) UnmarshalJSON(input []byte) error {
	s := string(input)
	if strings.HasPrefix(s, `"`) {
		return cj.unmarshalPlain(s)
	} else if strings.HasPrefix(s, `{`) {
		return cj.unmarshalMoreBits(s)
	} else {
		return errors.Errorf("Unknown how to parse colour from '%s'", string(input))
	}
}

func (cj *colourJSON) unmarshalPlain(s string) error {
	toParse := strings.Trim(s, `"`)
	if toParse == "" {
		cj.noColour = true
		return nil
	}
	colour, err := ParseColour(toParse)
	if err != nil {
		return errors.Wrap(err, "while parsing colourJson")
	}
	cj.bit4 = &colour
	return nil
}

func (cj *colourJSON) unmarshalMoreBits(s string) error {
	// To see which type of colour we're marshalling trim any junk then check for the prefix
	toCheck := strings.TrimLeft(s, "{ \n\t\"")
	if strings.HasPrefix(toCheck, "8-bit") {
		type inner struct {
			As8bit int `json:"8-bit"`
		}
		i := inner{}
		err := json.Unmarshal([]byte(s), &i)
		if err != nil {
			return errors.Wrap(err, "while parsing colourJson")
		}
		cj.bit8 = &i.As8bit
		return nil
	} else if strings.HasPrefix(toCheck, "24-bit") {
		type inner struct {
			As24bit bit24JSON `json:"24-bit"`
		}
		i := inner{}
		err := json.Unmarshal([]byte(s), &i)
		if err != nil {
			return errors.Wrap(err, "while parsing colourJson")
		}
		cj.bit24 = &i.As24bit
		return nil
	} else {
		return errors.Errorf("Unknown how to parse colour from '%s'", s)
	}
}

type bit24JSON struct {
	*string
	R *uint8 `json:"r"`
	G *uint8 `json:"g"`
	B *uint8 `json:"b"`
}

func (b24 *bit24JSON) UnmarshalJSON(input []byte) error {
	s := string(input)
	if strings.HasPrefix(s, `"`) {
		return b24.unmarshalSimple(s)
	} else if strings.HasPrefix(s, `{`) {
		return b24.unmarshalComponent(s)
	} else {
		return errors.Errorf("Unknown how to parse colour from '%s'", string(input))
	}
}

func (b24 *bit24JSON) unmarshalComponent(s string) error {
	type inner struct {
		R int `json:"r"`
		G int `json:"g"`
		B int `json:"b"`
	}
	i := inner{}
	err := json.Unmarshal([]byte(s), &i)
	if i.R > 255 || i.R < 0 {
		err = errors.Join(err, errors.Errorf("red component out of range %d, should be within 0 and 255", i.R))
	}
	if i.G > 255 || i.G < 0 {
		err = errors.Join(err, errors.Errorf("green component out of range %d, should be within 0 and 255", i.G))
	}
	if i.B > 255 || i.B < 0 {
		err = errors.Join(err, errors.Errorf("green component out of range %d, should be within 0 and 255", i.B))
	}
	if err != nil {
		return errors.Wrap(err, "while parsing colourJson")
	}
	// G115: These are not an integer overflow because we bounds check above ^

	b24.R = ptr(uint8(i.R)) //nolint:gosec
	b24.G = ptr(uint8(i.G)) //nolint:gosec
	b24.B = ptr(uint8(i.B)) //nolint:gosec
	return nil
}

func (b24 *bit24JSON) unmarshalSimple(s string) error {
	toParse := strings.Trim(s, `"`)
	r, g, b, err := cssStrToRGB(toParse)
	if err != nil {
		return errors.Wrap(err, "while parsing 24bit colour")
	}
	b24.string = &toParse
	b24.R = &r
	b24.G = &g
	b24.B = &b
	return nil
}

func ptr[T any](u T) *T {
	v := &u
	return v
}

func (cj *colourJSON) impl() (colourImpl, error) {
	switch {
	case cj.bit4 != nil:
		return colourImpl{
			Prefix: ansi.CSI + strconv.Itoa(int(*cj.bit4)) + "m",
			Suffix: ansi.FormattingReset,
			Origin: *cj,
		}, nil
	case cj.bit8 != nil:
		if *cj.bit8 > 255 || *cj.bit8 < 0 {
			return colourImpl{}, errors.Errorf("8-bit colour component out of range %d, should be within 0 and 255", *cj.bit8)
		}
		return colourImpl{
			Prefix: ansi.CSI + "38;5;" + strconv.Itoa(*cj.bit8) + "m",
			Suffix: ansi.FormattingReset,
			Origin: *cj,
		}, nil
	case cj.bit24 != nil:
		return colourImpl{
			Prefix: ansi.CSI + "38;2;" +
				strconv.Itoa(int(*cj.bit24.R)) + ";" +
				strconv.Itoa(int(*cj.bit24.G)) + ";" +
				strconv.Itoa(int(*cj.bit24.B)) + "m",
			Suffix: ansi.FormattingReset,
			Origin: *cj,
		}, nil
	case cj.noColour:
		return colourImpl{Origin: *cj}, nil
	}
	return colourImpl{}, errors.Errorf("no colour defined")
}

func (tj themeJSON) Theme() (Theme, error) {
	// the actual json colours may have json unmarshaled "ok" but contain bogus data in which case we will
	// get a descriptive error here about that.
	errs := []error{}
	primary, err := tj.Colours.Primary.impl()
	errs = append(errs, errors.Wrap(err, "primary"))
	secondary, err := tj.Colours.Secondary.impl()
	errs = append(errs, errors.Wrap(err, "secondary"))
	highlight, err := tj.Colours.Highlight.impl()
	errs = append(errs, errors.Wrap(err, "highlight"))
	emphasis, err := tj.Colours.Emphasis.impl()
	errs = append(errs, errors.Wrap(err, "emphasis"))
	titleHighlight, err := tj.Colours.TitleHighlight.impl()
	errs = append(errs, errors.Wrap(err, "title-highlight"))
	positive, err := tj.Colours.Positive.impl()
	errs = append(errs, errors.Wrap(err, "positive"))
	darkPositive, err := tj.Colours.DarkPositive.impl()
	errs = append(errs, errors.Wrap(err, "dark-positive"))
	negative, err := tj.Colours.Negative.impl()
	errs = append(errs, errors.Wrap(err, "negative"))
	darkNegative, err := tj.Colours.DarkNegative.impl()
	errs = append(errs, errors.Wrap(err, "dark-negative"))
	if err := errors.Join(errs...); err != nil {
		return Theme{}, errors.Wrap(err, "Couldn't parse theme, colours had errors")
	}
	return Theme{
		primary:        primary,
		secondary:      secondary,
		highlight:      highlight,
		emphasis:       emphasis,
		titleHighlight: titleHighlight,
		positive:       positive,
		darkPositive:   darkPositive,
		negative:       negative,
		darkNegative:   darkNegative,
	}, nil
}
