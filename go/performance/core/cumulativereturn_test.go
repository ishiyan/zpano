package core

import (
	"math"
	"testing"
)

func TestCumulativeReturnAnnualizedReturnDefinition(t *testing.T) {
	t.Parallel()
	returns := []float64{0.10, -0.05, 0.03, 0.08}
	periodsPerYear := 12.0

	cr := NewCumulativeReturn()
	for _, r := range returns {
		cr.Update(r)
	}

	growth := prod(mapf(returns, func(r float64) float64 { return 1 + r }))
	expected := math.Pow(growth, periodsPerYear/float64(len(returns))) - 1
	actual := cr.AnnualizedGeometricMeanReturn(periodsPerYear)
	assertAlmostEqual(t, actual, expected, places(15), "annualized")
}

func TestCumulativeReturnOneMonthReturn(t *testing.T) {
	t.Parallel()
	// One monthly return of 1%, expected (1.01)^12-1.
	cr := NewCumulativeReturn()
	cr.Update(0.01)

	expected := math.Pow(1.01, 12) - 1
	actual := cr.AnnualizedGeometricMeanReturn(12)
	assertAlmostEqual(t, actual, expected, places(15), "annualized")
}

func TestCumulativeReturnYearlyReturns(t *testing.T) {
	t.Parallel()
	// If the observation itself is yearly, then the annualized geometric
	// mean return equals the geometric mean return.
	cr := NewCumulativeReturn()
	for _, r := range []float64{0.12, -0.04, 0.08} {
		cr.Update(r)
	}

	expected := cr.GeometricMeanReturn()
	actual := cr.AnnualizedGeometricMeanReturn(1)
	assertAlmostEqual(t, actual, expected, places(15), "annualized")
}

func TestCumulativeReturnConstantMonthlyReturn(t *testing.T) {
	t.Parallel()
	// If every monthly return is exactly r, ((1+r)^n)^(12/n) = (1+r)^12.
	cr := NewCumulativeReturn()
	for range 60 {
		cr.Update(0.01)
	}

	expected := math.Pow(1.01, 12) - 1
	actual := cr.AnnualizedGeometricMeanReturn(12)
	assertAlmostEqual(t, actual, expected, places(15), "annualized")
}

func TestCumulativeReturnEmpty(t *testing.T) {
	t.Parallel()
	cr := NewCumulativeReturn()
	assertNaN(t, cr.AnnualizedGeometricMeanReturn(12), "annualized")
}

func TestCumulativeReturnZeroReturns(t *testing.T) {
	t.Parallel()
	// log1p(0) == 0 and expm1(0) == 0, so the result is exactly zero.
	cr := NewCumulativeReturn()
	for range 100 {
		cr.Update(0.0)
	}

	actual := cr.AnnualizedGeometricMeanReturn(252)
	assertAlmostEqual(t, actual, 0, places(15), "annualized")
}

func TestCumulativeReturnConsistencyWithGeometricMeanReturn(t *testing.T) {
	t.Parallel()
	// 1 + annualized = (1 + geometric mean)^p
	returns := []float64{0.0010, -0.0005, 0.0003, 0.0008}
	periodsPerYear := 252.0

	cr := NewCumulativeReturn()
	for _, r := range returns {
		cr.Update(r)
	}

	expected := math.Pow(1+cr.GeometricMeanReturn(), periodsPerYear) - 1
	actual := cr.AnnualizedGeometricMeanReturn(periodsPerYear)
	assertAlmostEqual(t, actual, expected, places(13), "annualized")
}

func TestCumulativeReturnRollingWindowMatchesFreshCalculation(t *testing.T) {
	t.Parallel()
	rng := newRNG(42)
	returns := make([]float64, 100)
	for i := range returns {
		returns[i] = zeroOrGauss(rng, 0.03)
	}
	const windowSize = 7
	cr := NewCumulativeReturn()
	for i, r := range returns {
		if i >= windowSize {
			if err := cr.Revert(returns[i-windowSize]); err != nil {
				t.Fatal(err)
			}
		}
		cr.Update(r)
		w := window(returns, windowSize, i)
		growth := prod(mapf(w, func(x float64) float64 { return 1 + x }))
		if cr.Count() != len(w) {
			t.Fatalf("step %d: count %d, expected %d", i, cr.Count(), len(w))
		}
		assertAlmostEqual(t, cr.CumulativeGeometricReturn(), growth-1, places(14), "cumulative")
		assertAlmostEqual(t, cr.GeometricMeanReturn(),
			math.Pow(growth, 1/float64(len(w)))-1, places(14), "geometric mean")
	}
}

func TestCumulativeReturnRevertEmptyRaises(t *testing.T) {
	t.Parallel()
	cr := NewCumulativeReturn()
	err := cr.Revert(0.01)
	if err == nil || err.Error() != "Cannot revert from an empty accumulator" {
		t.Errorf("expected error, got %v", err)
	}
}

func TestCumulativeReturnReset(t *testing.T) {
	t.Parallel()
	cr := NewCumulativeReturn()
	for _, r := range []float64{0.1, -0.2} {
		cr.Update(r)
	}
	cr.Reset()
	if cr.Count() != 0 {
		t.Errorf("count %d, expected 0", cr.Count())
	}
	if cr.CumulativeGeometricReturn() != 0.0 {
		t.Errorf("cumulative %v, expected 0", cr.CumulativeGeometricReturn())
	}
	assertNaN(t, cr.GeometricMeanReturn(), "geometric mean")
}
