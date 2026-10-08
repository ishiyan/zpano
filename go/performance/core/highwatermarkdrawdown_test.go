package core

import (
	"fmt"
	"slices"
	"testing"
)

// expectedHWMDrawdowns is an independent reference implementation of
// chronological high-water-mark drawdowns, matching R PerformanceAnalytics
// Drawdowns(): the high-water mark starts at the initial equity 1.
func expectedHWMDrawdowns(returns []float64) []float64 {
	equity := 1.0
	peak := 1.0
	result := []float64{}
	for _, ret := range returns {
		equity *= 1.0 + ret
		var dd float64
		if equity >= peak {
			peak = equity
			dd = 0.0
		} else {
			dd = equity/peak - 1.0
		}
		result = append(result, dd)
	}
	return result
}

// expectedRollingHWMDrawdowns returns the drawdowns of the rolling window
// ending at index i: a fresh calculation over the returns in the window.
func expectedRollingHWMDrawdowns(returns []float64, windowSize, i int) []float64 {
	lo := 0
	if windowSize > 0 {
		lo = max(0, i-windowSize+1)
	}
	return expectedHWMDrawdowns(returns[lo : i+1])
}

func assertHWMState(t *testing.T, acc *HighWaterMarkDrawdown, expected []float64, tol float64, msg string) {
	t.Helper()
	actual := acc.Drawdowns()
	if len(actual) != len(expected) {
		t.Fatalf("%s: %d drawdowns, expected %d", msg, len(actual), len(expected))
	}
	for i := range actual {
		assertAlmostEqual(t, actual[i], expected[i], tol, fmt.Sprintf("%s drawdown %d", msg, i))
	}
	if acc.DrawdownsCount() != len(expected) {
		t.Errorf("%s: count %d, expected %d", msg, acc.DrawdownsCount(), len(expected))
	}
	if len(expected) > 0 {
		assertAlmostEqual(t, acc.Drawdown(), expected[len(expected)-1], tol, msg+" drawdown")
		assertAlmostEqual(t, acc.MaximumDrawdown(), slices.Min(expected), tol, msg+" maximum drawdown")
		expectedMean := pySum(expected) / float64(len(expected))
		expectedSquaredMean := pySum(mapf(expected, func(x float64) float64 { return x * x })) /
			float64(len(expected))
		assertAlmostEqual(t, acc.DrawdownsMean(), expectedMean, tol, msg+" mean")
		assertAlmostEqual(t, acc.DrawdownsSquaredMean(), expectedSquaredMean, tol, msg+" squared mean")
	} else {
		assertNaN(t, acc.Drawdown(), msg+" drawdown")
		assertNaN(t, acc.MaximumDrawdown(), msg+" maximum drawdown")
		assertNaN(t, acc.DrawdownsMean(), msg+" mean")
		assertNaN(t, acc.DrawdownsSquaredMean(), msg+" squared mean")
	}
}

func feedHWM(acc *HighWaterMarkDrawdown, returns ...float64) *HighWaterMarkDrawdown {
	for _, r := range returns {
		acc.Update(r)
	}
	return acc
}

// Expanding-window tests.

func TestHighWaterMarkDrawdownExpandingEmpty(t *testing.T) {
	t.Parallel()
	assertHWMState(t, NewHighWaterMarkDrawdown(0), []float64{}, places(14), "")
}

func TestHighWaterMarkDrawdownExpandingAllPositiveReturns(t *testing.T) {
	t.Parallel()
	acc := feedHWM(NewHighWaterMarkDrawdown(0), 0.10, 0.05, 0.20)
	assertHWMState(t, acc, []float64{0.0, 0.0, 0.0}, places(14), "")
}

func TestHighWaterMarkDrawdownExpandingFirstNegativeReturn(t *testing.T) {
	t.Parallel()
	acc := feedHWM(NewHighWaterMarkDrawdown(0), -0.05, -0.02, 0.10)
	assertHWMState(t, acc, []float64{-0.05, -0.069, 0.0}, places(14), "")
}

func TestHighWaterMarkDrawdownExpandingSimpleDrawdownAndRecovery(t *testing.T) {
	t.Parallel()
	acc := feedHWM(NewHighWaterMarkDrawdown(0), 0.10, -0.05, 0.10)
	assertHWMState(t, acc, []float64{0.0, -0.05, 0.0}, places(14), "")
}

func TestHighWaterMarkDrawdownExpandingCompoundedDrawdown(t *testing.T) {
	t.Parallel()
	acc := feedHWM(NewHighWaterMarkDrawdown(0), 0.10, -0.10, -0.10)
	assertHWMState(t, acc, []float64{0.0, -0.10, -0.19}, places(14), "")
}

func TestHighWaterMarkDrawdownExpandingNewHighWaterMarkResetsDrawdown(t *testing.T) {
	t.Parallel()
	acc := feedHWM(NewHighWaterMarkDrawdown(0), 0.10, -0.05, 0.06, -0.02)
	assertHWMState(t, acc, []float64{0.0, -0.05, 0.0, -0.02}, places(14), "")
}

func TestHighWaterMarkDrawdownExpandingReset(t *testing.T) {
	t.Parallel()
	acc := feedHWM(NewHighWaterMarkDrawdown(0), 0.10, -0.05, -0.02)
	if acc.DrawdownsCount() <= 0 {
		t.Fatalf("count %d, expected > 0", acc.DrawdownsCount())
	}
	acc.Reset()
	assertHWMState(t, acc, []float64{}, places(14), "")

	// The accumulator can be reused, starting from equity 1.0 again.
	acc.Update(-0.05)
	assertHWMState(t, acc, []float64{-0.05}, places(14), "")
}

func TestHighWaterMarkDrawdownExpandingMatchesReference(t *testing.T) {
	t.Parallel()
	rng := newRNG(42)
	returns := make([]float64, 200)
	for i := range returns {
		returns[i] = gauss(rng, 0.0, 0.03)
	}
	acc := feedHWM(NewHighWaterMarkDrawdown(0), returns...)
	assertHWMState(t, acc, expectedHWMDrawdowns(returns), places(13), "")
}

func TestHighWaterMarkDrawdownZeroSizeMeansExpanding(t *testing.T) {
	t.Parallel()
	returns := []float64{0.10, -0.05, -0.02, 0.05}
	acc := feedHWM(NewHighWaterMarkDrawdown(0), returns...)
	assertHWMState(t, acc, expectedHWMDrawdowns(returns), places(14), "")
}

func TestHighWaterMarkDrawdownNegativeSizeMeansExpanding(t *testing.T) {
	t.Parallel()
	// The constructor normalizes non-positive window sizes to zero, so a
	// negative window size also means expanding mode.
	returns := []float64{0.10, -0.05, -0.02}
	acc := feedHWM(NewHighWaterMarkDrawdown(-10), returns...)
	assertHWMState(t, acc, expectedHWMDrawdowns(returns), places(14), "")
}

// Rolling-window tests: the window equals a fresh calculation over its
// returns, starting from the equity just before the window.

func TestHighWaterMarkDrawdownRollingWindowPeakEviction(t *testing.T) {
	t.Parallel()
	acc := feedHWM(NewHighWaterMarkDrawdown(3), 0.10, -0.05, -0.02)
	assertHWMState(t, acc, []float64{0.0, -0.05, -0.069}, places(14), "")
	acc.Update(0.03)
	assertHWMState(t, acc, []float64{-0.05, -0.069, -0.04107}, places(14), "")
}

func TestHighWaterMarkDrawdownRollingWindowEvictedPeakFollowedByNewPeak(t *testing.T) {
	t.Parallel()
	acc := feedHWM(NewHighWaterMarkDrawdown(3), 0.05, -0.02, 0.10)
	assertHWMState(t, acc, []float64{0.0, -0.02, 0.0}, places(14), "")
	acc.Update(-0.03)
	assertHWMState(t, acc, []float64{-0.02, 0.0, -0.03}, places(14), "")
}

func TestHighWaterMarkDrawdownRollingWindowPeakEvictionRecomputesDrawdowns(t *testing.T) {
	t.Parallel()
	acc := feedHWM(NewHighWaterMarkDrawdown(3), 0.10, -0.05, -0.05)
	assertHWMState(t, acc, []float64{0.0, -0.05, -0.0975}, places(14), "")
	acc.Update(0.01)
	assertHWMState(t, acc, []float64{-0.05, -0.0975, -0.088475}, places(14), "")
}

func TestHighWaterMarkDrawdownRollingWindowAllNegativeReturns(t *testing.T) {
	t.Parallel()
	returns := []float64{-0.01, -0.02, -0.03, -0.04}
	acc := NewHighWaterMarkDrawdown(3)
	for i, ret := range returns {
		acc.Update(ret)
		assertHWMState(t, acc, expectedRollingHWMDrawdowns(returns, 3, i), places(14), fmt.Sprintf("step %d", i))
	}
}

func TestHighWaterMarkDrawdownRollingWindowSizeOne(t *testing.T) {
	t.Parallel()
	returns := []float64{0.10, -0.05, -0.02, 0.03, -0.04}
	acc := NewHighWaterMarkDrawdown(1)
	for i, ret := range returns {
		acc.Update(ret)
		assertHWMState(t, acc, []float64{min(ret, 0.0)}, places(14), fmt.Sprintf("step %d", i))
	}
}

func TestHighWaterMarkDrawdownRollingWindowMatchesFreshCalculation(t *testing.T) {
	t.Parallel()
	rng := newRNG(7)
	for _, windowSize := range []int{2, 3, 4, 7, 20} {
		for range 10 {
			returns := make([]float64, 80)
			for i := range returns {
				returns[i] = gauss(rng, 0.0, 0.03)
			}
			acc := NewHighWaterMarkDrawdown(windowSize)
			for i, ret := range returns {
				acc.Update(ret)
				assertHWMState(t, acc, expectedRollingHWMDrawdowns(returns, windowSize, i), places(13),
					fmt.Sprintf("window %d step %d", windowSize, i))
			}
		}
	}
}

func TestHighWaterMarkDrawdownRollingWindowRecomputeFlag(t *testing.T) {
	t.Parallel()
	// Evicting an observation can only change the remaining high-water
	// marks if its return was negative (its equity is below the window's
	// starting equity).
	acc := NewHighWaterMarkDrawdown(2)
	check := func(ret float64, expected bool) {
		t.Helper()
		if got := acc.Update(ret); got != expected {
			t.Errorf("Update(%v) = %v, expected %v", ret, got, expected)
		}
	}
	check(0.10, false)
	check(-0.05, false)
	// Evicts +10%: the remaining peaks don't change.
	check(-0.02, false)
	// Evicts -5% and the new first observation (-2%) is also below the old
	// starting equity: recompute.
	check(0.01, true)
	assertHWMState(t, acc, []float64{-0.02, -0.0102}, places(14), "")
	// Evicts -2%; the new first observation (equity 1.0343) is still below
	// the old starting equity (1.045): recompute.
	check(0.03, true)
	assertHWMState(t, acc, []float64{0.0, 0.0}, places(14), "")
	// Evicts +1%: its equity is above the window's starting equity, so the
	// remaining peaks don't change.
	check(-0.01, false)
	assertHWMState(t, acc, []float64{0.0, -0.01}, places(14), "")
}
