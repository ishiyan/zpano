package core

import (
	"fmt"
	"math"
	"testing"
)

// ddRun is the compounded drawdown of one continuous losing run.
func ddRun(returns ...float64) float64 {
	compounded := prod(mapf(returns, func(r float64) float64 { return 1.0 + r*0.01 }))
	return (compounded - 1.0) * 100.0
}

func assertRunDrawdownsAlmostEqual(t *testing.T, actual, expected []float64, tol float64, prefix string) {
	t.Helper()
	if len(actual) != len(expected) {
		t.Fatalf("%s: length %d, expected %d", prefix, len(actual), len(expected))
	}
	for i := range actual {
		assertAlmostEqual(t, actual[i], expected[i], tol, fmt.Sprintf("%s step %d", prefix, i))
	}
}

// assertRunsState checks the drawdowns, the run count (when
// expectedRunCount >= 0) and the sums of squares.
func assertRunsState(t *testing.T, acc *ContinuousDrawdownRuns, expectedDrawdowns []float64,
	expectedRunCount int, tol float64, prefix string,
) {
	t.Helper()
	assertRunDrawdownsAlmostEqual(t, acc.Drawdowns(), expectedDrawdowns, tol, prefix)
	if expectedRunCount >= 0 && acc.RunCount() != expectedRunCount {
		t.Errorf("%s: run count %d, expected %d", prefix, acc.RunCount(), expectedRunCount)
	}
	expectedSumSq := pySum(mapf(expectedDrawdowns, func(x float64) float64 { return x * x }))
	assertAlmostEqual(t, acc.SumDrawdownsSquared(), expectedSumSq, tol, prefix+" sum squared")
	assertAlmostEqual(t, acc.SqrtSumDrawdownsSquared(), math.Sqrt(expectedSumSq), tol, prefix+" sqrt sum squared")
}

func feedRuns(acc *ContinuousDrawdownRuns, returns ...float64) {
	for _, r := range returns {
		acc.Update(r)
	}
}

// Expanding-window tests.

func TestContinuousDrawdownRunsExpandingSingleLosingRun(t *testing.T) {
	t.Parallel()
	acc := NewContinuousDrawdownRuns()
	acc.Update(-1.0)
	assertRunsState(t, acc, []float64{ddRun(-1.0)}, 1, places(12), "")
	acc.Update(-2.0)
	assertRunsState(t, acc, []float64{ddRun(-1.0, -2.0)}, 1, places(12), "")
	acc.Update(-3.0)
	assertRunsState(t, acc, []float64{ddRun(-1.0, -2.0, -3.0)}, 1, places(12), "")
}

func TestContinuousDrawdownRunsExpandingMultipleLosingRuns(t *testing.T) {
	t.Parallel()
	acc := NewContinuousDrawdownRuns()
	feedRuns(acc, -1.0, -2.0, 1.0, -3.0, -4.0, 2.0, -5.0)
	expected := []float64{ddRun(-1.0, -2.0), ddRun(-3.0, -4.0), ddRun(-5.0)}
	assertRunsState(t, acc, expected, 3, places(12), "")
}

func TestContinuousDrawdownRunsExpandingNonNegativeReturnsAreSeparators(t *testing.T) {
	t.Parallel()
	acc := NewContinuousDrawdownRuns()
	feedRuns(acc, -2.0, 0.0, -3.0, 1.0, -4.0)
	expected := []float64{ddRun(-2.0), ddRun(-3.0), ddRun(-4.0)}
	assertRunsState(t, acc, expected, 3, places(12), "")
}

func TestContinuousDrawdownRunsExpandingPositiveReturnDoesNotCreateDrawdown(t *testing.T) {
	t.Parallel()
	acc := NewContinuousDrawdownRuns()
	feedRuns(acc, 2.0, 3.0, 0.0, 5.0)
	assertRunsState(t, acc, []float64{}, 0, places(12), "")
}

func TestContinuousDrawdownRunsReset(t *testing.T) {
	t.Parallel()
	acc := NewContinuousDrawdownRuns()
	feedRuns(acc, -1.0, -2.0, 1.0, -3.0)
	if acc.RunCount() <= 0 {
		t.Fatalf("run count %d, expected > 0", acc.RunCount())
	}

	acc.Reset()
	assertRunsState(t, acc, []float64{}, 0, places(12), "")

	// It must also be possible to use it again after reset.
	acc.Update(-4.0)
	assertRunsState(t, acc, []float64{ddRun(-4.0)}, 1, places(12), "")
}

// Rolling-window tests.

func TestContinuousDrawdownRunsRollingWindowEvictionFromFrontOfLosingRun(t *testing.T) {
	t.Parallel()
	acc := NewContinuousDrawdownRuns()
	feedRuns(acc, -1.0, -2.0, -3.0)
	assertRunsState(t, acc, []float64{ddRun(-1.0, -2.0, -3.0)}, 1, places(12), "")
	acc.Revert(-1.0)
	acc.Update(-4.0)
	assertRunsState(t, acc, []float64{ddRun(-2.0, -3.0, -4.0)}, 1, places(12), "")
}

func TestContinuousDrawdownRunsRollingWindowEvictionOfEntireLosingRun(t *testing.T) {
	t.Parallel()
	acc := NewContinuousDrawdownRuns()
	feedRuns(acc, -1.0, -2.0, 1.0)
	assertRunsState(t, acc, []float64{ddRun(-1.0, -2.0)}, 1, places(12), "")
	acc.Revert(-1.0)
	acc.Update(-3.0)
	assertRunsState(t, acc, []float64{ddRun(-2.0), ddRun(-3.0)}, 2, places(12), "")
}

func TestContinuousDrawdownRunsRollingWindowEvictionOfSeparator(t *testing.T) {
	t.Parallel()
	acc := NewContinuousDrawdownRuns()
	feedRuns(acc, -1.0, 1.0, -2.0)
	assertRunsState(t, acc, []float64{ddRun(-1.0), ddRun(-2.0)}, 2, places(12), "")
	acc.Revert(-1.0)
	acc.Update(-3.0)
	assertRunsState(t, acc, []float64{ddRun(-2.0, -3.0)}, 1, places(12), "")
}

func TestContinuousDrawdownRunsRollingWindowMultipleRuns(t *testing.T) {
	t.Parallel()
	acc := NewContinuousDrawdownRuns()
	feedRuns(acc, -1.0, -2.0, 1.0, -3.0, -4.0)
	assertRunsState(t, acc, []float64{ddRun(-1.0, -2.0), ddRun(-3.0, -4.0)}, 2, places(12), "")

	// Slide 1: remove -1%, add +2%
	acc.Revert(-1.0)
	acc.Update(2.0)
	assertRunsState(t, acc, []float64{ddRun(-2.0), ddRun(-3.0, -4.0)}, 2, places(12), "")

	// Slide 2: remove -2%, add -5%
	acc.Revert(-2.0)
	acc.Update(-5.0)
	assertRunsState(t, acc, []float64{ddRun(-3.0, -4.0), ddRun(-5.0)}, 2, places(12), "")
}

func TestContinuousDrawdownRunsRollingWindowNewReturnExtendsExistingRun(t *testing.T) {
	t.Parallel()
	acc := NewContinuousDrawdownRuns()
	feedRuns(acc, 1.0, -2.0, -3.0)
	assertRunsState(t, acc, []float64{ddRun(-2.0, -3.0)}, 1, places(12), "")
	acc.Revert(1.0)
	acc.Update(-4.0)
	assertRunsState(t, acc, []float64{ddRun(-2.0, -3.0, -4.0)}, 1, places(12), "")
}

func TestContinuousDrawdownRunsRollingWindowNewNegativeStartsNewRunAfterSeparator(t *testing.T) {
	t.Parallel()
	acc := NewContinuousDrawdownRuns()
	feedRuns(acc, -2.0, 1.0, -3.0)
	assertRunsState(t, acc, []float64{ddRun(-2.0), ddRun(-3.0)}, 2, places(12), "")
	acc.Revert(-2.0)
	acc.Update(-4.0)
	assertRunsState(t, acc, []float64{ddRun(-3.0, -4.0)}, 1, places(12), "")
}

func TestContinuousDrawdownRunsRevertThenUpdateOrderIsRequired(t *testing.T) {
	t.Parallel()
	acc := NewContinuousDrawdownRuns()
	feedRuns(acc, -1.0, -2.0, 2.0)
	assertRunsState(t, acc, []float64{ddRun(-1.0, -2.0)}, 1, places(12), "")

	// New window: [-2%, +2%, -3%]
	acc.Revert(-1.0)
	acc.Update(-3.0)
	assertRunsState(t, acc, []float64{ddRun(-2.0), ddRun(-3.0)}, 2, places(12), "")
}

// Numerical consistency.

func TestContinuousDrawdownRunsSqrtSumDrawdownsSquared(t *testing.T) {
	t.Parallel()
	acc := NewContinuousDrawdownRuns()
	feedRuns(acc, -1.0, -2.0, 1.0, -3.0)

	dd1 := ddRun(-1.0, -2.0)
	dd2 := ddRun(-3.0)

	expectedSumSq := dd1*dd1 + dd2*dd2
	expectedSqrt := math.Sqrt(expectedSumSq)

	assertAlmostEqual(t, acc.SumDrawdownsSquared(), expectedSumSq, places(12), "sum squared")
	assertAlmostEqual(t, acc.SqrtSumDrawdownsSquared(), expectedSqrt, places(12), "sqrt sum squared")
}

// Brute-force rolling-window test.

func TestContinuousDrawdownRunsRollingWindowMatchesFreshCalculation(t *testing.T) {
	t.Parallel()
	referenceRuns := func(window []float64) []float64 {
		runs := []float64{}
		var current []float64
		for _, r := range window {
			if r < 0 {
				current = append(current, r)
			} else if len(current) > 0 {
				runs = append(runs, ddRun(current...))
				current = nil
			}
		}
		if len(current) > 0 {
			runs = append(runs, ddRun(current...))
		}
		return runs
	}

	rng := newRNG(42)
	for _, windowSize := range []int{1, 2, 3, 5, 12} {
		returns := make([]float64, 150)
		for i := range returns {
			returns[i] = zeroOrGauss(rng, 3.0)
		}
		acc := NewContinuousDrawdownRuns()
		for i, ret := range returns {
			if i >= windowSize {
				acc.Revert(returns[i-windowSize])
			}
			acc.Update(ret)
			expected := referenceRuns(window(returns, windowSize, i))
			assertRunsState(t, acc, expected, len(expected), places(12),
				fmt.Sprintf("window %d step %d", windowSize, i))
		}
	}
}
