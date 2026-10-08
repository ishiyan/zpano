package core

import (
	"fmt"
	"testing"
)

var winLossGetters = []namedGetter[*WinLoss]{
	{"non_zero_returns_count", intGetter((*WinLoss).NonZeroReturnsCount)},
	{"non_zero_returns_mean", (*WinLoss).NonZeroReturnsMean},
	{"winning_returns_count", intGetter((*WinLoss).WinningReturnsCount)},
	{"winning_returns_sum", (*WinLoss).WinningReturnsSum},
	{"winning_returns_mean", (*WinLoss).WinningReturnsMean},
	{"losing_returns_count", intGetter((*WinLoss).LosingReturnsCount)},
	{"losing_returns_sum", (*WinLoss).LosingReturnsSum},
	{"losing_returns_mean", (*WinLoss).LosingReturnsMean},
}

func winLossReference(returns []float64) map[string]float64 {
	wins := filter(returns, func(r float64) bool { return r > 0 })
	losses := filter(returns, func(r float64) bool { return r < 0 })
	nonZero := filter(returns, func(r float64) bool { return r != 0 })
	return map[string]float64{
		"non_zero_returns_count": float64(len(nonZero)),
		"non_zero_returns_mean":  pyMean(nonZero),
		"winning_returns_count":  float64(len(wins)),
		"winning_returns_sum":    pySum(wins),
		"winning_returns_mean":   pyMean(wins),
		"losing_returns_count":   float64(len(losses)),
		"losing_returns_sum":     pySum(losses),
		"losing_returns_mean":    pyMean(losses),
	}
}

func TestWinLossEmpty(t *testing.T) {
	t.Parallel()
	assertGettersMatch(t, NewWinLoss(), winLossGetters, winLossReference(nil), places(15), "")
}

func TestWinLossHandComputed(t *testing.T) {
	t.Parallel()
	wl := NewWinLoss()
	for _, r := range []float64{0.02, 0.0, -0.01, 0.04, 0.0, -0.03} {
		wl.Update(r)
	}
	v002 := 0.02
	if wl.NonZeroReturnsCount() != 4 {
		t.Errorf("non-zero count %d, expected 4", wl.NonZeroReturnsCount())
	}
	assertAlmostEqual(t, wl.NonZeroReturnsMean(), v002/4, places(16), "non-zero mean")
	if wl.WinningReturnsCount() != 2 {
		t.Errorf("winning count %d, expected 2", wl.WinningReturnsCount())
	}
	assertAlmostEqual(t, wl.WinningReturnsMean(), 0.03, places(16), "winning mean")
	if wl.LosingReturnsCount() != 2 {
		t.Errorf("losing count %d, expected 2", wl.LosingReturnsCount())
	}
	assertAlmostEqual(t, wl.LosingReturnsMean(), -0.02, places(16), "losing mean")
}

func TestWinLossRollingWindowMatchesReference(t *testing.T) {
	t.Parallel()
	rng := newRNG(42)
	returns := make([]float64, 120)
	for i := range returns {
		returns[i] = zeroOrGauss(rng, 0.03)
	}
	const w = 6
	wl := NewWinLoss()
	for i, r := range returns {
		if i >= w {
			wl.Revert(returns[i-w])
		}
		wl.Update(r)
		assertGettersMatch(t, wl, winLossGetters, winLossReference(window(returns, w, i)),
			places(15), fmt.Sprintf("step %d", i))
	}
}

func TestWinLossReset(t *testing.T) {
	t.Parallel()
	wl := NewWinLoss()
	for _, r := range []float64{0.01, -0.02} {
		wl.Update(r)
	}
	wl.Reset()
	assertGettersMatch(t, wl, winLossGetters, winLossReference(nil), places(15), "")
}
