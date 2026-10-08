package core

import (
	"errors"
	"math"
	"slices"
)

var (
	errPercentileQ     = errors.New("q must be between 0 and 1")
	errPercentileEmpty = errors.New("window must not be empty")
)

// Percentile computes the q-th percentile of window using NumPy's
// method="linear" definition. q is in the range [0, 1].
//
// The input slice is not modified (a sorted copy is used).
//
// Returns an error if q is outside [0, 1] or the window is empty.
func Percentile(window []float64, q float64) (float64, error) {
	if !(0 <= q && q <= 1) {
		return math.NaN(), errPercentileQ
	}

	values := slices.Clone(window)
	if len(values) == 0 {
		return math.NaN(), errPercentileEmpty
	}
	slices.Sort(values)
	n := len(values)

	if n == 1 {
		return values[0], nil
	}

	idx := q * float64(n-1)
	lo := int(idx)

	if lo >= n-1 {
		return values[n-1], nil
	}

	hi := lo + 1
	frac := idx - float64(lo)

	return values[lo] + frac*(values[hi]-values[lo]), nil
}
