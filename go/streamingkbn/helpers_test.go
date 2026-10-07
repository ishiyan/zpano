package streamingkbn

import (
	"math"
	"math/big"
	"testing"
)

// Bacon, Carl R., Practical Portfolio Performance Measurement and
// Attribution, 2nd ed., Wiley, 2008, p. 65 (portfolio returns).
var bacon = []float64{
	0.003, 0.026, 0.011, -0.010,
	0.015, 0.025, 0.016, 0.067,
	-0.014, 0.040, -0.005, 0.081,
	0.040, -0.037, -0.061, 0.017,
	-0.049, -0.022, 0.070, 0.058,
	-0.065, 0.024, -0.005, -0.009,
}

// Bacon, p. 66 (benchmark returns).
var baconBenchmark = []float64{
	0.002, 0.025, 0.018, -0.011,
	0.014, 0.018, 0.014, 0.065,
	-0.015, 0.042, -0.006, 0.083,
	0.039, -0.038, -0.062, 0.015,
	-0.048, 0.021, 0.060, 0.056,
	-0.067, 0.019, -0.003, 0.000,
}

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

func assertAlmostEqual(t *testing.T, name string, got, want float64, n int) {
	t.Helper()
	if !almostEqual(got, want, places(n)) {
		t.Errorf("%s = %v, want %v (places=%d)", name, got, want, n)
	}
}

func assertEqual(t *testing.T, name string, got, want float64) {
	t.Helper()
	if got != want {
		t.Errorf("%s = %v, want exactly %v", name, got, want)
	}
}

func assertEqualInt(t *testing.T, name string, got, want int) {
	t.Helper()
	if got != want {
		t.Errorf("%s = %d, want %d", name, got, want)
	}
}

func assertNaN(t *testing.T, name string, got float64) {
	t.Helper()
	if !math.IsNaN(got) {
		t.Errorf("%s = %v, want NaN", name, got)
	}
}

// assertNaNOrAlmostEqual requires got to be NaN when want is NaN and
// almost equal to want otherwise.
func assertNaNOrAlmostEqual(t *testing.T, name string, got, want float64, n int) {
	t.Helper()
	if math.IsNaN(want) {
		assertNaN(t, name, got)
		return
	}
	assertAlmostEqual(t, name, got, want, n)
}

// assertPanics requires f to panic with the string message want.
func assertPanics(t *testing.T, want string, f func()) {
	t.Helper()
	defer func() {
		r := recover()
		if r == nil {
			t.Errorf("expected panic %q, got none", want)
			return
		}
		if s, ok := r.(string); !ok || s != want {
			t.Errorf("panic = %v, want %q", r, want)
		}
	}()
	f()
}

// naiveSum adds the values left to right without compensation.
func naiveSum(data []float64) float64 {
	s := 0.0
	for _, x := range data {
		s += x
	}
	return s
}

// fsum is a port of CPython's math.fsum (Shewchuk's exact partials with
// a final half-even rounding correction) for finite inputs.
func fsum(data []float64) float64 {
	p := make([]float64, 0, 32)
	for _, x := range data {
		i := 0
		for _, y := range p {
			if math.Abs(x) < math.Abs(y) {
				x, y = y, x
			}
			hi := x + y
			lo := y - (hi - x)
			if lo != 0.0 {
				p[i] = lo
				i++
			}
			x = hi
		}
		p = p[:i]
		if x != 0.0 {
			p = append(p, x)
		}
	}

	n := len(p)
	hi, lo := 0.0, 0.0
	if n > 0 {
		n--
		hi = p[n]
		// Sum exactly from the top, stop when the sum becomes inexact.
		for n > 0 {
			x := hi
			n--
			y := p[n]
			hi = x + y
			yr := hi - x
			lo = y - yr
			if lo != 0.0 {
				break
			}
		}
		// Make half-even rounding work across multiple partials.
		if n > 0 && ((lo < 0.0 && p[n-1] < 0.0) || (lo > 0.0 && p[n-1] > 0.0)) {
			y := lo * 2.0
			x := hi + y
			yr := x - hi
			if y == yr {
				hi = x
			}
		}
	}
	return hi
}

// fmean mirrors Python's statistics.fmean: fsum(data) / n.
func fmean(data []float64) float64 {
	return fsum(data) / float64(len(data))
}

// pvariance is the exact population variance (as Python's
// statistics.pvariance), computed with rational arithmetic and
// rounded to the nearest float.
func pvariance(data []float64) float64 {
	n := new(big.Rat).SetInt64(int64(len(data)))
	sum := new(big.Rat)
	for _, x := range data {
		sum.Add(sum, new(big.Rat).SetFloat64(x))
	}
	mean := new(big.Rat).Quo(sum, n)
	ss := new(big.Rat)
	for _, x := range data {
		d := new(big.Rat).Sub(new(big.Rat).SetFloat64(x), mean)
		ss.Add(ss, d.Mul(d, d))
	}
	v, _ := ss.Quo(ss, n).Float64()
	return v
}

// window returns data[max(0, i-w+1) : i+1].
func window(data []float64, i, w int) []float64 {
	lo := i - w + 1
	if lo < 0 {
		lo = 0
	}
	return data[lo : i+1]
}
