package core

import (
	"math"
	"math/rand/v2"
	"testing"
)

// places converts a Python unittest places=N to an absolute tolerance,
// round(a-b, N) == 0 being approximately |a-b| <= 0.5·10⁻ᴺ.
func places(n int) float64 {
	return 0.5 * math.Pow10(-n)
}

// almostEqual mirrors unittest.assertAlmostEqual: exact equality passes,
// otherwise the absolute difference must be within tol (NaN never passes).
func almostEqual(a, b, tol float64) bool {
	return a == b || math.Abs(a-b) <= tol
}

// assertAlmostEqual fails the test if actual and expected differ by more
// than tol.
func assertAlmostEqual(t *testing.T, actual, expected, tol float64, msg string) {
	t.Helper()
	if !almostEqual(actual, expected, tol) {
		t.Errorf("%s: expected %v, got %v (diff %g, tol %g)",
			msg, expected, actual, math.Abs(actual-expected), tol)
	}
}

// assertNaN fails the test if actual is not NaN.
func assertNaN(t *testing.T, actual float64, msg string) {
	t.Helper()
	if !math.IsNaN(actual) {
		t.Errorf("%s: expected NaN, got %v", msg, actual)
	}
}

// assertNaNOrAlmostEqual expects NaN when expected is NaN and an almost
// equal value otherwise.
func assertNaNOrAlmostEqual(t *testing.T, actual, expected, tol float64, msg string) {
	t.Helper()
	if math.IsNaN(expected) {
		assertNaN(t, actual, msg)
		return
	}
	assertAlmostEqual(t, actual, expected, tol, msg)
}

// pySum mirrors Python's built-in sum() for floats (CPython 3.12+ uses
// Neumaier compensated summation).
func pySum(values []float64) float64 {
	result, c := 0.0, 0.0
	for _, x := range values {
		t := result + x
		if math.Abs(result) >= math.Abs(x) {
			c += (result - t) + x
		} else {
			c += (x - t) + result
		}
		result = t
	}
	if c != 0 && !math.IsInf(c, 0) && !math.IsNaN(c) {
		result += c
	}
	return result
}

// prod mirrors math.prod: the product of values, left to right.
func prod(values []float64) float64 {
	p := 1.0
	for _, v := range values {
		p *= v
	}
	return p
}

// mapf applies f to every value.
func mapf(values []float64, f func(float64) float64) []float64 {
	out := make([]float64, len(values))
	for i, v := range values {
		out[i] = f(v)
	}
	return out
}

// filter keeps the values for which keep is true.
func filter(values []float64, keep func(float64) bool) []float64 {
	out := []float64{}
	for _, v := range values {
		if keep(v) {
			out = append(out, v)
		}
	}
	return out
}

// newRNG is the counterpart of Python's random.Random(seed).
func newRNG(seed uint64) *rand.Rand {
	return rand.New(rand.NewPCG(seed, 0))
}

// gauss mirrors random.gauss(mu, sigma).
func gauss(rng *rand.Rand, mu, sigma float64) float64 {
	return mu + sigma*rng.NormFloat64()
}

// choice mirrors random.choice(values).
func choice(rng *rand.Rand, values ...float64) float64 {
	return values[rng.IntN(len(values))]
}

// zeroOrGauss mirrors rng.choice([0.0, rng.gauss(0.0, sigma)]).
func zeroOrGauss(rng *rand.Rand, sigma float64) float64 {
	g := gauss(rng, 0.0, sigma)
	return choice(rng, 0.0, g)
}

// window returns values[max(0, i-w+1) : i+1].
func window(values []float64, w, i int) []float64 {
	return values[max(0, i-w+1) : i+1]
}

// nanDiv returns a / b, or NaN when b is zero.
func nanDiv(a, b float64) float64 {
	if b != 0 {
		return a / b
	}
	return math.NaN()
}

// pyMean returns sum(values) / len(values), NaN when empty.
func pyMean(values []float64) float64 {
	if len(values) == 0 {
		return math.NaN()
	}
	return pySum(values) / float64(len(values))
}
