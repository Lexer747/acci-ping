// Use of this source code is governed by a GPL-2 license that can be found in the LICENSE file.
//
// Copyright 2024-2025 Lexer747
//
// SPDX-License-Identifier: GPL-2.0-only

package numeric_test

import (
	"fmt"
	"math"
	"testing"
	"time"

	"pgregory.net/rapid"

	"github.com/Lexer747/acci-ping/utils/numeric"
	"github.com/Lexer747/acci-ping/utils/th"
)

func TestNormalize(t *testing.T) {
	t.Parallel()
	type Case struct {
		Min, Max       float64
		NewMin, NewMax float64
		Inputs         []float64
		Expected       []float64
	}
	cases := []Case{
		{
			Min:    float64(7_657_469 * time.Microsecond),
			Max:    float64(12_301_543 * time.Microsecond),
			NewMin: 2,
			NewMax: 24,
			Inputs: []float64{
				float64(7_706_944 * time.Microsecond),
				float64(7_750_314 * time.Microsecond),
				float64(7_789_195 * time.Microsecond),
				float64(12_301_543 * time.Microsecond),
				float64(7_657_469 * time.Microsecond),
			},
			Expected: []float64{
				2.23,
				2.44,
				2.62,
				24,
				2,
			},
		},
	}
	for i, test := range cases {
		t.Run(fmt.Sprintf("%d:%f->%f|%+v", i, test.Min, test.Max, test.Inputs), func(t *testing.T) {
			t.Parallel()
			for i, input := range test.Inputs {
				actual := numeric.NormalizeToRange(input, test.Min, test.Max, test.NewMin, test.NewMax)
				th.AssertFloatEqual(t, test.Expected[i], actual, 3)
			}
		})
	}
}

func TestExponent_Property(t *testing.T) {
	t.Parallel()
	rapid.Check(t, func(t *rapid.T) {
		var (
			a = rapid.Float64().Draw(t, "a")
			b = rapid.Float64().Draw(t, "b")
		)
		if numeric.Exponent(a)*numeric.Exponent(b) != numeric.Exponent(b)*numeric.Exponent(a) {
			t.Fatalf("Exponent() is not commutative")
		}
	})
	rapid.Check(t, func(t *rapid.T) {
		var (
			a = rapid.Float64().Draw(t, "a")
		)
		switch {
		case math.Abs(a) <= 1 && math.Abs(a) >= 0:
			if isPos(numeric.Exponent(a)) {
				t.Fatalf("Exponent() numbers smaller than 0 should have negative exponents: %f", numeric.Exponent(a))
			}
		case math.Abs(a) > 1:
			if isNeg(numeric.Exponent(a)) {
				t.Fatalf("Exponent() numbers larger than 0 should have positive exponents: %f", numeric.Exponent(a))
			}
		default:
			panic("uncovered case")
		}
	})
}

func TestNormalize_Property(t *testing.T) {
	t.Parallel()
	rapid.Check(t, func(t *rapid.T) {
		var (
			a = rapid.Float64().Draw(t, "a")
		)
		normalized := numeric.Normalize(a, -math.MaxFloat64, math.MaxFloat64)
		if normalized < 0 || normalized > 1 {
			t.Fatalf("Normalize() was not in [0, 1] range: %f", normalized)
		}
	})
	rapid.Check(t, func(t *rapid.T) {
		var (
			a        = rapid.Float64().Draw(t, "a")
			minInput = rapid.Float64().Draw(t, "min")
			maxInput = rapid.Float64Min(minInput).Draw(t, "max")
		)
		normalized := numeric.NormalizeToRange(a, -math.MaxFloat64, math.MaxFloat64, minInput, maxInput)
		if normalized < minInput || normalized > maxInput {
			t.Fatalf("Normalize() was not in [%f, %f] range: %f", normalized, minInput, maxInput)
		}
	})
}

func isNeg(f float64) bool {
	return f < 0
}

func isPos(f float64) bool {
	return f > 0
}
