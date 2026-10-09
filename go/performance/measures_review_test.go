package performance

import (
	"math"
	"testing"
)

func TestReviewGainToPainRepetitionAndRolling(t *testing.T) {
	for _, window := range []int{0, 2, 4} {
		m := newTestMeasures(t, 1, 0, 0, window)
		if !math.IsNaN(m.GainToPainRatio()) {
			t.Fatal("empty gain-to-pain must be NaN")
		}
		for range 4 {
			m.AddReturn(0.1, 0)
			m.AddReturn(-0.05, 0)
			assertAlmostEqual(t, m.GainToPainRatio(), 1, 13, "sum/sum independent of repetition")
		}
		m.Reset()
		m.AddReturn(0.1, 0)
		if !math.IsNaN(m.GainToPainRatio()) {
			t.Fatal("gain-to-pain without losses must be NaN")
		}
	}
}

func TestReviewModifiedInformationGeometricSign(t *testing.T) {
	m := newTestMeasures(t, 1, 0, 0, 0)
	m.AddReturn(0.5, 0.05)
	m.AddReturn(-0.3, 0.05)
	if m.ActivePremium() >= 0 {
		t.Fatal("geometric active premium must be negative")
	}
	assertAlmostEqual(t, m.InformationRatioModified(), 0.04473320734100739, 13, "opposite arithmetic/geometric signs")
}

func TestReviewValidationBeforeDataFallback(t *testing.T) {
	m := newTestMeasures(t, 1, 0, 0, 0)
	for _, populated := range []bool{false, true} {
		if populated {
			m.AddReturn(-0.1, 0)
			m.AddReturn(0.2, 0)
		}
		for _, c := range []float64{-1e20, -1, 0, 1, 2, math.NaN(), math.Inf(1), math.Inf(-1)} {
			if _, err := m.IsNormalDistribution(c); err != errConfidence {
				t.Errorf("normality %v: %v", c, err)
			}
			if _, err := m.RewardToConditionalDrawdown(c); err != errConfidence {
				t.Errorf("reward-to-drawdown %v: %v", c, err)
			}
		}
		for _, k := range []float64{0, -1, math.NaN(), math.Inf(-1)} {
			if _, err := m.BiasRatio(k); err != errStdDevMultiplier {
				t.Errorf("bias %v: %v", k, err)
			}
		}
	}
	empty := newTestMeasures(t, 1, 0, 0, 0)
	if !math.IsNaN(must(empty.RewardToConditionalDrawdown(0.95))) {
		t.Fatal("valid empty result must be NaN")
	}
}
