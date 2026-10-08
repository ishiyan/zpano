package core

import (
	"testing"
)

func TestSFMRegressionFullBullAndBearFitsUseExcessReturns(t *testing.T) {
	t.Parallel()
	// Runtime variables so the expressions are evaluated in float64.
	rf, alpha := 0.01, 0.005
	regression := NewSFMRegression(rf)
	for _, excessBenchmark := range []float64{-0.04, -0.02, 0.0, 0.02, 0.05} {
		benchmark := rf + excessBenchmark
		portfolio := rf + alpha + 2*excessBenchmark
		regression.Update(portfolio, benchmark)
	}

	assertAlmostEqual(t, regression.Alpha(), 0.005, places(14), "alpha")
	assertAlmostEqual(t, regression.Beta(), 2.0, places(14), "beta")
	assertAlmostEqual(t, regression.BetaBull(), 2.0, places(14), "beta bull")
	assertAlmostEqual(t, regression.BetaBear(), 2.0, places(14), "beta bear")
	assertAlmostEqual(t, regression.R2(), 1.0, places(14), "r2")

	// Removing an older bear observation leaves too few bear points for a
	// slope, while the full and bull fits remain defined.
	excess := -0.04
	regression.Revert(rf+alpha+2*excess, rf+excess)
	assertAlmostEqual(t, regression.Beta(), 2.0, places(14), "beta")
	assertAlmostEqual(t, regression.BetaBull(), 2.0, places(14), "beta bull")
	assertNaN(t, regression.BetaBear(), "beta bear")
}

func TestSFMRegressionResetAndZeroExcessBenchmark(t *testing.T) {
	t.Parallel()
	regression := NewSFMRegression(0.01)
	regression.Update(0.02, 0.01)
	assertNaN(t, regression.BetaBull(), "beta bull")
	assertNaN(t, regression.BetaBear(), "beta bear")
	regression.Reset()
	for _, g := range []struct {
		name string
		get  func(*SFMRegression) float64
	}{
		{"alpha", (*SFMRegression).Alpha},
		{"beta", (*SFMRegression).Beta},
		{"beta_bull", (*SFMRegression).BetaBull},
		{"beta_bear", (*SFMRegression).BetaBear},
		{"r2", (*SFMRegression).R2},
	} {
		assertNaN(t, g.get(regression), g.name)
	}
}
