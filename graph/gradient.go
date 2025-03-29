// Use of this source code is governed by a GPL-2 license that can be found in the LICENSE file.
//
// Copyright 2024-2025 Lexer747
//
// SPDX-License-Identifier: GPL-2.0-only

package graph

import (
	"fmt"

	"github.com/Lexer747/acci-ping/graph/terminal/ansi"
	"github.com/Lexer747/acci-ping/graph/terminal/typography"
	"github.com/Lexer747/acci-ping/utils/check"
)

func solve(x []int, y []int) []gradient {
	check.Check(len(x) == len(y), "x and y should be equal len")
	if len(x) <= 1 {
		return []gradient{}
	}
	if len(x) == 2 {
		return []gradient{gradientSolve(x[0], x[1], y[0], y[1])}
	}
	result := make([]gradient, len(x)-1)
	xDirs := make([]direction, len(x)-1)
	yDirs := make([]direction, len(x)-1)
	xEqualsCount := 0
	yEqualsCount := 0
	for i := range len(x) - 1 {
		xDirs[i] = getDir(x[i], x[i+1])
		// // y values are inverted
		yDirs[i] = getDir(y[i+1], y[i])
		if xDirs[i] == equal {
			xEqualsCount++
		}
		if yDirs[i] == equal {
			yEqualsCount++
		}
	}
	if yEqualsCount > len(yDirs)/2 {
		solve := []gradient{}
		var specific bool
		for i := range len(xDirs) - 1 {
			if specific {
				specific = false
				continue
			}
			specific, solve = solveShallowTwoDirections(xDirs[i], yDirs[i], xDirs[i+1], yDirs[i+1])
			result[i] = solve[0]
			if specific {
				result[i+1] = solve[1]
			}
		}
		result[len(result)-1] = solve[1]
	} else {
		solve := []gradient{}
		for i := range len(xDirs) - 1 {
			solve = solveTwoDirections(xDirs[i], yDirs[i], xDirs[i+1], yDirs[i+1])
			result[i] = solve[0]
		}
		result[len(result)-1] = solve[1]
	}

	return result
}

func gradientSolve(beginX, beginY, endX, endY int) gradient {
	xDir := getDir(beginX, endX)
	// y values are inverted
	yDir := getDir(endY, beginY)
	return solveDirections(xDir, yDir)
}

func solveShallowTwoDirections(firstX direction, firstY direction, secondX direction, secondY direction) (bool, []gradient) {
	first := solveDirections(firstX, firstY)
	second := solveDirections(secondX, secondY)
	switch {
	case (first == upSteep && second == horizontal) || (first == upSteep && second == nothing):
		return true, []gradient{topLine, bottomLine}
	case (first == downSteep && second == horizontal) || (first == downSteep && second == nothing):
		return true, []gradient{bottomLine, topLine}
	case (first == horizontal && second == upSteep) || (first == nothing && second == upSteep):
		return false, []gradient{first, topLine}
	case (first == horizontal && second == downSteep) || (first == nothing && second == downSteep):
		return false, []gradient{first, bottomLine}
	default:
		return false, []gradient{first, second}
	}
}

func solveTwoDirections(firstX direction, firstY direction, secondX direction, secondY direction) []gradient {
	first := solveDirections(firstX, firstY)
	second := solveDirections(secondX, secondY)
	switch {
	case first == horizontal && second == vertical:
		return []gradient{gap, vertical}
	case first == vertical && second == horizontal:
		return []gradient{vertical, gap}
	default:
		return []gradient{first, second}
	}
}

func solveDirections(xDir direction, yDir direction) gradient {
	if xDir == positive && yDir == positive {
		return upSteep
	} else if xDir == equal && yDir == positive {
		return vertical
	} else if xDir == negative && yDir == positive {
		return downSteep
	}
	if (xDir == positive || xDir == negative) && yDir == equal {
		return horizontal
	} else if xDir == equal && yDir == equal {
		return nothing
	}
	if xDir == positive && yDir == negative {
		return downSteep
	} else if xDir == equal && yDir == negative {
		return vertical
	} else if xDir == negative && yDir == negative {
		return upSteep
	}
	panic(fmt.Sprintf("Case not covered %d %d", xDir, yDir))
}

func getDir(begin int, end int) direction {
	switch {
	case begin < end:
		return negative
	case begin == end:
		return equal
	default:
		return positive
	}
}

type direction int

const (
	positive direction = 1
	equal    direction = 0
	negative direction = -1
)

type gradient int

const (
	nothing gradient = -1

	gap gradient = 1

	upSteep    gradient = 2
	vertical   gradient = 3
	horizontal gradient = 4
	downSteep  gradient = 5

	topLine    gradient = 6
	bottomLine gradient = 7
)

func (g gradient) draw() string {
	ret := ""
	switch g {
	case nothing:
		return ret
	case gap:
		ret = " "
	case upSteep:
		ret = "/"
	case vertical:
		ret = typography.Vertical
	case horizontal:
		ret = "-"
	case downSteep:
		ret = `\`
	case topLine:
		ret = typography.TopLine
	case bottomLine:
		ret = typography.BottomLine
	default:
		panic(fmt.Sprintf("unexpected gradient %d", g))
	}
	return ansi.Gray(ret)
}
