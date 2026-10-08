package performance

import (
	"fmt"
	"math"
	"testing"

	"zpano/performance/referencedata"
)

func TestTailRatioMatchesReferenceImplementationOutput(t *testing.T) {
	t.Parallel()
	for cutoff, expected := range referencedata.TailRatioExpectedValuesByCutoffReference {
		actual := runStream(t, stream{}, func(m *Measures) float64 { return must(m.TailRatio(cutoff)) })
		assertSeriesEqual(t, actual, expected, places(15), fmt.Sprintf("tail ratio (yearly, cutoff %v)", cutoff))
	}
}

func TestTailRatioValidation(t *testing.T) {
	t.Parallel()
	m := makeMeasures(t, 0, 0, 0, false, false)
	for _, c := range []float64{0.5, 1.0, 0.2, math.NaN()} {
		if _, err := m.TailRatio(c); err == nil || err.Error() != "cutoff must be between 0.5 and 1.0" {
			t.Errorf("cutoff %v: expected error, got %v", c, err)
		}
	}
}

// assertDailyMonthlyYearlySameRefs checks the same reference map with the
// periodic rate annualized for daily, monthly and yearly data.
func assertDailyMonthlyYearlySameRefs(t *testing.T, refs map[float64][]float64, f func(*Measures) float64, tol tolerance, name string) {
	t.Helper()
	for rf, expected := range refs {
		assertSeriesEqual(t, runStream(t, dailyRf(rf), f), expected, tol, fmt.Sprintf("%s (daily, Rf %v)", name, rf))
	}
	for rf, expected := range refs {
		assertSeriesEqual(t, runStream(t, monthlyRf(rf), f), expected, tol, fmt.Sprintf("%s (monthly, Rf %v)", name, rf))
	}
	for rf, expected := range refs {
		assertSeriesEqual(t, runStream(t, yearlyRf(rf), f), expected, tol, fmt.Sprintf("%s (yearly, Rf %v)", name, rf))
	}
}

func TestKellyRatioMatchesPerformanceAnalyticsOutput(t *testing.T) {
	t.Parallel()
	assertDailyMonthlyYearlySameRefs(t, referencedata.KellyRatioExpectedValuesByRfPerfan, (*Measures).KellyRatio, places(11), "Kelly ratio")
}

func TestKellyRatioFullMatchesPerformanceAnalyticsOutput(t *testing.T) {
	t.Parallel()
	assertDailyMonthlyYearlySameRefs(t, referencedata.KellyRatioExpectedValuesByRfFullPerfan, (*Measures).KellyRatioFull, places(11), "Kelly ratio full")
}

func TestHurstExponentMatchesPerformanceAnalyticsOutput(t *testing.T) {
	t.Parallel()
	actual := runStream(t, stream{}, (*Measures).HurstExponent)
	assertSeriesEqual(t, actual, referencedata.HurstExponentExpectedValuesPerfan, places(14), "Hurst exponent")
}

func TestBiasRatioMatchesReferenceImplementationOutput(t *testing.T) {
	t.Parallel()
	for mult, expected := range referencedata.BiasRatioExpectedValuesByMultReference {
		actual := runStream(t, stream{}, func(m *Measures) float64 { return must(m.BiasRatio(mult)) })
		assertSeriesEqual(t, actual, expected, places(15), fmt.Sprintf("bias ratio (yearly, std_dev_multiplier %v)", mult))
	}
}

func TestBiasRatioValidation(t *testing.T) {
	t.Parallel()
	m := makeMeasures(t, 0, 0, 0, false, false)
	for _, k := range []float64{0, -1} {
		if _, err := m.BiasRatio(k); err == nil || err.Error() != "std_dev_multiplier must be positive" {
			t.Errorf("multiplier %v: expected error, got %v", k, err)
		}
	}
}

func TestKRatioMatchesReferenceImplementationOutput(t *testing.T) {
	t.Parallel()
	actual := runStream(t, stream{}, (*Measures).KRatio)
	assertSeriesEqual(t, actual, referencedata.KRatioExpectedValuesReference, places(14), "K-ratio")
}

func TestGainToPainRatioMatchesReferenceImplementationOutput(t *testing.T) {
	t.Parallel()
	actual := runStream(t, stream{}, (*Measures).GainToPainRatio)
	assertSeriesEqual(t, actual, referencedata.GainToPainRatioExpectedValuesReference, places(15), "Gain-to-pain ratio")
}
