//nolint:testpackage
package corona

import (
	"math"
	"testing"
)

// TestCoronaDefaultSmoke feeds the 252-entry TA-Lib series through a default
// Corona and verifies that the engine primes and produces finite DC/DC-median
// values inside the expected [MinimalPeriod, MaximalPeriod] band.
func TestCoronaDefaultSmoke(t *testing.T) {
	t.Parallel()

	c, err := NewCorona(nil)
	if err != nil {
		t.Fatalf("NewCorona: %v", err)
	}

	if c.FilterBankLength() != 49 {
		t.Errorf("FilterBankLength = %d, want 49", c.FilterBankLength())
	}
	if c.MinimalPeriodTimesTwo() != 12 || c.MaximalPeriodTimesTwo() != 60 {
		t.Errorf("half-period range = [%d, %d], want [12, 60]",
			c.MinimalPeriodTimesTwo(), c.MaximalPeriodTimesTwo())
	}

	// The engine should report primed no later than sample 2*MinPeriod = 12.
	input := talibInput()
	primedAt := -1
	for i, v := range input {
		_ = c.Update(v)
		if c.IsPrimed() && primedAt < 0 {
			primedAt = i
		}
	}
	if primedAt < 0 {
		t.Fatal("engine never primed over 252 samples")
	}
	if primedAt+1 != c.MinimalPeriodTimesTwo() {
		t.Errorf("primedAt (1-based) = %d, want %d",
			primedAt+1, c.MinimalPeriodTimesTwo())
	}

	dc := c.DominantCycle()
	dcMed := c.DominantCycleMedian()

	if math.IsNaN(dc) || math.IsInf(dc, 0) {
		t.Errorf("DominantCycle = %v, want finite", dc)
	}
	if math.IsNaN(dcMed) || math.IsInf(dcMed, 0) {
		t.Errorf("DominantCycleMedian = %v, want finite", dcMed)
	}

	min := float64(c.MinimalPeriod())
	max := float64(c.MaximalPeriod())
	if dc < min || dc > max {
		t.Errorf("DominantCycle = %v, want in [%v, %v]", dc, min, max)
	}
	if dcMed < min || dcMed > max {
		t.Errorf("DominantCycleMedian = %v, want in [%v, %v]", dcMed, min, max)
	}

	// MaximalAmplitudeSquared should be non-zero and finite for the last bar.
	m := c.MaximalAmplitudeSquared()
	if m <= 0 || math.IsNaN(m) || math.IsInf(m, 0) {
		t.Errorf("MaximalAmplitudeSquared = %v, want positive finite", m)
	}

	t.Logf("final DominantCycle = %.6f", dc)
	t.Logf("final DominantCycleMedian = %.6f", dcMed)
	t.Logf("final MaximalAmplitudeSquared = %.6f", m)
	t.Logf("primed at sample (1-based) = %d", primedAt+1)

	// Spot-check a handful of bars across the series so that a regression in
	// the filter bank shows up as obviously-wrong numbers in `go test -v`.
	c2, _ := NewCorona(nil)
	sawAboveMin := false
	for i, v := range input {
		c2.Update(v)
		if i == 11 || i == 30 || i == 60 || i == 100 || i == 150 || i == 200 || i == 251 {
			t.Logf("bar %3d: DC=%.4f DCmed=%.4f maxAmp²=%.4f",
				i, c2.DominantCycle(), c2.DominantCycleMedian(), c2.MaximalAmplitudeSquared())
		}
		if c2.IsPrimed() && c2.DominantCycle() > float64(c2.MinimalPeriod()) {
			sawAboveMin = true
		}
	}
	if !sawAboveMin {
		t.Error("DominantCycle never exceeded MinimalPeriod across 252 samples (filter bank likely broken)")
	}
}

// TestCoronaNaNInputIsNoOp verifies that feeding NaN leaves the engine in its
// prior state and returns the prior primed status.
func TestCoronaNaNInputIsNoOp(t *testing.T) {
	t.Parallel()

	c, err := NewCorona(nil)
	if err != nil {
		t.Fatalf("NewCorona: %v", err)
	}

	// Warm the engine past priming.
	for _, v := range talibInput()[:20] {
		c.Update(v)
	}
	if !c.IsPrimed() {
		t.Fatal("expected primed after 20 samples")
	}

	dcBefore := c.DominantCycle()
	dcMedBefore := c.DominantCycleMedian()

	if got := c.Update(math.NaN()); !got {
		t.Errorf("Update(NaN) returned %v, want true (preserves primed)", got)
	}
	if c.DominantCycle() != dcBefore || c.DominantCycleMedian() != dcMedBefore {
		t.Error("NaN input mutated cached outputs")
	}
}

// TestCoronaInvalidParams exercises the parameter validator.
func TestCoronaInvalidParams(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name string
		p    Params
	}{
		{"cutoff too small", Params{HighPassFilterCutoff: 1}},
		{"min too small", Params{MinimalPeriod: 1}},
		{"max <= min", Params{MinimalPeriod: 10, MaximalPeriod: 10}},
		{"negative dB lower", Params{DecibelsLowerThreshold: -1}},
		{"dB upper <= lower", Params{DecibelsLowerThreshold: 6, DecibelsUpperThreshold: 6}},
	}
	for _, tc := range cases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			if _, err := NewCorona(&tc.p); err == nil {
				t.Error("expected error, got nil")
			}
		})
	}
}
