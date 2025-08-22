package service

import (
	"math"
	"testing"
)

const eps = 1e-9

func almostEqual(a, b float64) bool {
	return math.Abs(a-b) <= eps
}

func TestAdd_SimpleIntegers(t *testing.T) {
	got := Add(2, 3)   // untyped constants -> OK for float64
	want := float64(5) // must be float64 for comparison
	if !almostEqual(got, want) {
		t.Fatalf("Add(2,3)=%g; want %g", got, want)
	}
}

func TestAdd_Floats(t *testing.T) {
	tcs := []struct {
		name     string
		a, b     float64
		expected float64
	}{
		{"simple decimals", 2.5, 3.1, 5.6},
		{"negatives", -1.5, -2.5, -4.0},
		{"mix", 10.75, -0.25, 10.5},
		{"zero", 0, 0, 0},
	}

	for _, tc := range tcs {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			got := Add(tc.a, tc.b)
			if !almostEqual(got, tc.expected) {
				t.Fatalf("Add(%g,%g)=%g; want %g (diff %g)",
					tc.a, tc.b, got, tc.expected, math.Abs(got-tc.expected))
			}
		})
	}
}
