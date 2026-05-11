package fuzzy

import "math"

// AlphaCut converts a continuous fuzzy output to a crisp discrete value.
//
// The confidence is abs(value) / scale. If confidence ≥ alpha,
// the output is rounded to the nearest multiple of scale with the
// original sign preserved. Otherwise 0 is returned.
func AlphaCut(value, alpha, scale float64) int {
	if scale <= 0.0 {
		return 0
	}
	confidence := math.Abs(value) / scale
	if confidence < alpha-1e-10 {
		return 0
	}
	sign := 1
	if value < 0 {
		sign = -1
	}
	// Round to nearest multiple of scale.
	level := math.Round(confidence)
	if level < 1 {
		level = 1
	}
	return sign * int(level*scale)
}
