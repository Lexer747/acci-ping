// Use of this source code is governed by a GPL-2 license that can be found in the LICENSE file.
//
// Copyright 2025 Lexer747
//
// SPDX-License-Identifier: GPL-2.0-only

package themes

var CSSStrToRGB = cssStrToRGB
var Impl = func(cj ColourJSON) (colourImpl, error) {
	return (&cj).impl()
}

type ColourJSON = colourJSON
