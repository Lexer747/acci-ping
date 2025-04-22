// Use of this source code is governed by a GPL-2 license that can be found in the LICENSE file.
//
// Copyright 2024-2025 Lexer747
//
// SPDX-License-Identifier: GPL-2.0-only

package gradient

import (
	"fmt"

	"github.com/Lexer747/acci-ping/gui/themes"
	"github.com/Lexer747/acci-ping/terminal/typography"
	"github.com/Lexer747/acci-ping/utils/check"
)

func Solve(x []int, y []int) []Solution {
	check.Check(len(x) == len(y), "x and y should be equal len")
	if len(x) <= 1 {
		return []Solution{}
	}
	if len(x) == 2 {
		return []Solution{gradientSolve(x[0], y[0], x[1], y[1])}
	}
	result := make([]Solution, len(x)-1)
	xDirs := make([]direction, len(x)-1)
	yDirs := make([]direction, len(x)-1)
	xEqualsCount := 0
	yEqualsCount := 0
	for i := range len(x) - 1 {
		xDirs[i] = getDir(x[i], x[i+1])
		// y values are inverted
		yDirs[i] = getDir(y[i+1], y[i])
		if xDirs[i] == equal {
			xEqualsCount++
		}
		if yDirs[i] == equal {
			yEqualsCount++
		}
	}
	if yEqualsCount > len(yDirs)/2 {
		solve := []Solution{}
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
		solve := []Solution{}
		for i := range len(xDirs) - 1 {
			solve = solveTwoDirections(xDirs[i], yDirs[i], xDirs[i+1], yDirs[i+1])
			result[i] = solve[0]
		}
		result[len(result)-1] = solve[1]
	}

	return result
}

func gradientSolve(beginX, beginY, endX, endY int) Solution {
	xDir := getDir(beginX, endX)
	// y values are inverted
	yDir := getDir(endY, beginY)
	return solveDirections(xDir, yDir)
}

func solveShallowTwoDirections(firstX direction, firstY direction, secondX direction, secondY direction) (bool, []Solution) {
	first := solveDirections(firstX, firstY)
	second := solveDirections(secondX, secondY)
	switch {
	case (first == upSteep && second == horizontal) || (first == upSteep && second == nothing):
		return true, []Solution{topLine, bottomLine}
	case (first == downSteep && second == horizontal) || (first == downSteep && second == nothing):
		return true, []Solution{bottomLine, topLine}
	case (first == horizontal && second == upSteep) || (first == nothing && second == upSteep):
		return false, []Solution{first, topLine}
	case (first == horizontal && second == downSteep) || (first == nothing && second == downSteep):
		return false, []Solution{first, bottomLine}
	default:
		return false, []Solution{first, second}
	}
}

func solveTwoDirections(firstX direction, firstY direction, secondX direction, secondY direction) []Solution {
	first := solveDirections(firstX, firstY)
	second := solveDirections(secondX, secondY)
	switch {
	case first == horizontal && second == vertical:
		return []Solution{gap, vertical}
	case first == vertical && second == horizontal:
		return []Solution{vertical, gap}
	default:
		return []Solution{first, second}
	}
}

func solveDirections(xDir direction, yDir direction) Solution {
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

type Solution int

const (
	nothing Solution = -1

	gap Solution = 1

	upSteep    Solution = 2
	vertical   Solution = 3
	horizontal Solution = 4
	downSteep  Solution = 5

	topLine    Solution = 6
	bottomLine Solution = 7
)

func (g Solution) Draw() string {
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
	return themes.Secondary(ret)
}
