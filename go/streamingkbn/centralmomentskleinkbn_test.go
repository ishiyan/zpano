package streamingkbn

import (
	"fmt"
	"math"
	"testing"
)

// Reference values for bacon, computed with exact rational arithmetic
// (fractions.Fraction on the binary float inputs, square roots with
// 50-digit decimal.Decimal), rounded to the nearest float. The names
// follow scipy.stats: skew(bias=...), kurtosis(bias=..., fisher=...).
const (
	centralMean                = 0.009000000000000001
	centralVarianceDdof0       = 0.0014989166666666668
	centralVarianceDdof1       = 0.0015640869565217393
	centralStdDdof0            = 0.03871584516275819
	centralStdDdof1            = 0.039548539246370897
	centralSkewBiased          = -0.08256245520856804 // skew(bias=True)
	centralSkewUnbiased        = -0.08817174934967535 // skew(bias=False)
	centralKurtBiasedFisher    = -0.5675462058921257  // kurtosis(bias=True, fisher=True)
	centralKurtBiasedPearson   = 2.4324537941078743   // kurtosis(bias=True, fisher=False)
	centralKurtUnbiasedFisher  = -0.40766032118608714 // kurtosis(bias=False, fisher=True)
	centralKurtUnbiasedPearson = 2.592339678813913    // kurtosis(bias=False, fisher=False)

	// Same statistics for [1e4 + x for x in bacon], exact for the shifted floats.
	centralOffsetSkewBiased       = -0.08256245521966786
	centralOffsetKurtBiasedFisher = -0.5675462058934164
)

func feedCentral(m *CentralMomentsKleinKBN, data []float64) *CentralMomentsKleinKBN {
	for _, x := range data {
		m.Update(x)
	}
	return m
}

func TestCentralMomentsKleinKBN_SimpleUpdate(t *testing.T) {
	t.Parallel()

	m := feedCentral(NewCentralMomentsKleinKBN(0, true, true), []float64{1.0, 2.0, 3.0, 4.0})
	assertEqualInt(t, "n", m.N(), 4)
	assertAlmostEqual(t, "mean", m.Mean(), 2.5, 15)
	assertAlmostEqual(t, "variance", m.Variance(), 1.25, 15)
	assertAlmostEqual(t, "skewness", m.Skewness(), 0.0, 14)
	assertAlmostEqual(t, "kurtosis", m.Kurtosis(), -1.36, 13)
}

func TestCentralMomentsKleinKBN_BaconMeanVariance(t *testing.T) {
	t.Parallel()

	m0 := feedCentral(NewCentralMomentsKleinKBN(0, true, true), bacon)
	m1 := feedCentral(NewCentralMomentsKleinKBN(1, true, true), bacon)
	assertAlmostEqual(t, "mean", m0.Mean(), centralMean, 16)
	assertAlmostEqual(t, "variance ddof=0", m0.Variance(), centralVarianceDdof0, 16)
	assertAlmostEqual(t, "variance ddof=1", m1.Variance(), centralVarianceDdof1, 16)
	assertAlmostEqual(t, "standard_deviation ddof=0", m0.StandardDeviation(), centralStdDdof0, 15)
	assertAlmostEqual(t, "standard_deviation ddof=1", m1.StandardDeviation(), centralStdDdof1, 15)
}

func TestCentralMomentsKleinKBN_BaconSkewnessKurtosis(t *testing.T) {
	t.Parallel()

	cases := []struct {
		bias, fisher bool
		skew, kurt   float64
	}{
		{true, true, centralSkewBiased, centralKurtBiasedFisher},
		{true, false, centralSkewBiased, centralKurtBiasedPearson},
		{false, true, centralSkewUnbiased, centralKurtUnbiasedFisher},
		{false, false, centralSkewUnbiased, centralKurtUnbiasedPearson},
	}
	for _, c := range cases {
		m := feedCentral(NewCentralMomentsKleinKBN(0, c.bias, c.fisher), bacon)
		label := fmt.Sprintf("bias=%v fisher=%v", c.bias, c.fisher)
		assertAlmostEqual(t, label+" skewness", m.Skewness(), c.skew, 14)
		assertAlmostEqual(t, label+" kurtosis", m.Kurtosis(), c.kurt, 13)
	}
}

func TestCentralMomentsKleinKBN_LargeOffset(t *testing.T) {
	t.Parallel()

	// Central moments don't suffer from the cancellation of raw power sums.
	shifted := make([]float64, len(bacon))
	for i, x := range bacon {
		shifted[i] = 1e4 + x
	}
	m := feedCentral(NewCentralMomentsKleinKBN(0, true, true), shifted)
	mean := float64(centralMean)
	assertAlmostEqual(t, "mean", m.Mean(), 1e4+mean, 11)
	assertAlmostEqual(t, "variance", m.Variance(), centralVarianceDdof0, 13)
	assertAlmostEqual(t, "skewness", m.Skewness(), centralOffsetSkewBiased, 10)
	assertAlmostEqual(t, "kurtosis", m.Kurtosis(), centralOffsetKurtBiasedFisher, 10)
}

func TestCentralMomentsKleinKBN_ScaleInvariance(t *testing.T) {
	t.Parallel()

	scaled := make([]float64, len(bacon))
	for i, x := range bacon {
		scaled[i] = x * 1e-6
	}
	m := feedCentral(NewCentralMomentsKleinKBN(0, true, true), scaled)
	assertAlmostEqual(t, "skewness", m.Skewness(), centralSkewBiased, 14)
	assertAlmostEqual(t, "kurtosis", m.Kurtosis(), centralKurtBiasedFisher, 13)
}

func TestCentralMomentsKleinKBN_Empty(t *testing.T) {
	t.Parallel()

	m := NewCentralMomentsKleinKBN(1, true, true)
	assertEqualInt(t, "n", m.N(), 0)
	assertEqual(t, "mean", m.Mean(), 0.0)
	assertNaN(t, "variance", m.Variance())
	assertNaN(t, "standard_deviation", m.StandardDeviation())
	assertNaN(t, "skewness", m.Skewness())
	assertNaN(t, "kurtosis", m.Kurtosis())
}

func TestCentralMomentsKleinKBN_Ddof(t *testing.T) {
	t.Parallel()

	m := feedCentral(NewCentralMomentsKleinKBN(1, true, true), []float64{1.0, 2.0, 3.0})
	assertAlmostEqual(t, "variance", m.Variance(), 1.0, 15)
	assertAlmostEqual(t, "standard_deviation", m.StandardDeviation(), 1.0, 15)
	m = feedCentral(NewCentralMomentsKleinKBN(1, true, true), []float64{1.0})
	assertNaN(t, "variance", m.Variance())
	assertNaN(t, "standard_deviation", m.StandardDeviation())
}

func TestCentralMomentsKleinKBN_InvalidDdof(t *testing.T) {
	t.Parallel()

	// Non-integer and bool ddof values are rejected by the Go type system.
	assertPanics(t, "ddof must be a nonnegative integer", func() {
		NewCentralMomentsKleinKBN(-1, true, true)
	})
	assertPanics(t, "ddof must be a nonnegative integer", func() {
		NewCentralMomentsKleinKBN(1, true, true).SetDdof(-1)
	})
}

func TestCentralMomentsKleinKBN_MinimumSampleSizes(t *testing.T) {
	t.Parallel()

	data := []float64{1.0, 2.0, 4.0, 8.0}
	// (bias, fisher) -> minimum n for (skewness, kurtosis)
	cases := []struct {
		bias, fisher  bool
		skewN, kurtN int
	}{
		{true, true, 2, 2},
		{true, false, 2, 2},
		{false, true, 3, 4},
		{false, false, 3, 4},
	}
	for _, c := range cases {
		m := NewCentralMomentsKleinKBN(0, c.bias, c.fisher)
		for i, x := range data {
			m.Update(x)
			n := i + 1
			if s := m.Skewness(); math.IsNaN(s) != (n < c.skewN) {
				t.Errorf("bias=%v fisher=%v n=%d skewness = %v, want NaN=%v",
					c.bias, c.fisher, n, s, n < c.skewN)
			}
			if k := m.Kurtosis(); math.IsNaN(k) != (n < c.kurtN) {
				t.Errorf("bias=%v fisher=%v n=%d kurtosis = %v, want NaN=%v",
					c.bias, c.fisher, n, k, n < c.kurtN)
			}
		}
	}
}

func TestCentralMomentsKleinKBN_ConstantData(t *testing.T) {
	t.Parallel()

	m := feedCentral(NewCentralMomentsKleinKBN(0, true, true), []float64{3.0, 3.0, 3.0, 3.0, 3.0})
	assertEqual(t, "mean", m.Mean(), 3.0)
	assertEqual(t, "variance", m.Variance(), 0.0)
	assertEqual(t, "standard_deviation", m.StandardDeviation(), 0.0)
	assertNaN(t, "skewness", m.Skewness())
	assertNaN(t, "kurtosis", m.Kurtosis())
}

func TestCentralMomentsKleinKBN_RevertLifoSimple(t *testing.T) {
	t.Parallel()

	data := []float64{10.0, 18.0, 5.0}
	mFull := feedCentral(NewCentralMomentsKleinKBN(0, true, true), data)
	mPart := feedCentral(NewCentralMomentsKleinKBN(0, true, true), data[:2])
	mFull.Revert(data[2])

	assertEqualInt(t, "n", mFull.N(), 2)
	assertAlmostEqual(t, "mean", mFull.Mean(), mPart.Mean(), 15)
	assertAlmostEqual(t, "variance", mFull.Variance(), mPart.Variance(), 15)
	assertAlmostEqual(t, "skewness", mFull.Skewness(), mPart.Skewness(), 14)
	assertAlmostEqual(t, "kurtosis", mFull.Kurtosis(), mPart.Kurtosis(), 13)
}

func TestCentralMomentsKleinKBN_RevertLifoBacon(t *testing.T) {
	t.Parallel()

	for _, c := range []struct{ bias, fisher bool }{{true, true}, {false, false}} {
		mFull := feedCentral(NewCentralMomentsKleinKBN(0, c.bias, c.fisher), bacon)
		mPart := feedCentral(NewCentralMomentsKleinKBN(0, c.bias, c.fisher), bacon[:len(bacon)-1])
		mFull.Revert(bacon[len(bacon)-1])

		label := fmt.Sprintf("bias=%v fisher=%v", c.bias, c.fisher)
		assertAlmostEqual(t, label+" mean", mFull.Mean(), mPart.Mean(), 15)
		assertAlmostEqual(t, label+" variance", mFull.Variance(), mPart.Variance(), 15)
		assertAlmostEqual(t, label+" skewness", mFull.Skewness(), mPart.Skewness(), 13)
		assertAlmostEqual(t, label+" kurtosis", mFull.Kurtosis(), mPart.Kurtosis(), 12)
	}
}

func TestCentralMomentsKleinKBN_RevertThenUpdate(t *testing.T) {
	t.Parallel()

	m := feedCentral(NewCentralMomentsKleinKBN(0, true, true), bacon)
	tail := bacon[12:]
	for i := len(tail) - 1; i >= 0; i-- {
		m.Revert(tail[i])
	}
	feedCentral(m, tail)
	assertAlmostEqual(t, "mean", m.Mean(), centralMean, 15)
	assertAlmostEqual(t, "variance", m.Variance(), centralVarianceDdof0, 15)
	assertAlmostEqual(t, "skewness", m.Skewness(), centralSkewBiased, 12)
	assertAlmostEqual(t, "kurtosis", m.Kurtosis(), centralKurtBiasedFisher, 12)
}

func TestCentralMomentsKleinKBN_RevertLifoRoundtrip(t *testing.T) {
	t.Parallel()

	m := feedCentral(NewCentralMomentsKleinKBN(0, true, true), bacon)
	for i := len(bacon) - 1; i >= 0; i-- {
		m.Revert(bacon[i])
	}
	assertEqualInt(t, "n", m.N(), 0)
	assertEqual(t, "mean", m.Mean(), 0.0)
	assertNaN(t, "variance", m.Variance())
}

func TestCentralMomentsKleinKBN_RevertOldestAndMiddle(t *testing.T) {
	t.Parallel()

	data := []float64{0.0, 1.0, 2.0, 4.0, 8.0}
	m := feedCentral(NewCentralMomentsKleinKBN(0, true, true), data)
	for _, removed := range []float64{0.0, 2.0} {
		m.Revert(removed)
		// Remove the first occurrence, as Python's list.remove.
		for i, x := range data {
			if x == removed {
				data = append(data[:i:i], data[i+1:]...)
				break
			}
		}
		mean := fmean(data)
		d2 := make([]float64, len(data))
		d3 := make([]float64, len(data))
		d4 := make([]float64, len(data))
		for i, x := range data {
			d2[i] = math.Pow(x-mean, 2)
			d3[i] = math.Pow(x-mean, 3)
			d4[i] = math.Pow(x-mean, 4)
		}
		fn := float64(len(data))
		mu2 := fsum(d2) / fn
		mu3 := fsum(d3) / fn
		mu4 := fsum(d4) / fn
		label := fmt.Sprintf("removed=%v", removed)
		assertEqualInt(t, label+" n", m.N(), len(data))
		assertAlmostEqual(t, label+" mean", m.Mean(), mean, 14)
		assertAlmostEqual(t, label+" variance", m.Variance(), mu2, 14)
		assertAlmostEqual(t, label+" skewness", m.Skewness(), mu3/math.Pow(mu2, 1.5), 13)
		assertAlmostEqual(t, label+" kurtosis", m.Kurtosis(), mu4/math.Pow(mu2, 2)-3, 13)
	}
}

func TestCentralMomentsKleinKBN_FifoRollingWindow(t *testing.T) {
	t.Parallel()

	m := NewCentralMomentsKleinKBN(0, true, true)
	const width = 6
	for i, x := range bacon {
		m.Update(x)
		if i >= width {
			m.Revert(bacon[i-width])
		}
		win := window(bacon, i, width)
		label := fmt.Sprintf("step=%d", i)
		assertEqualInt(t, label+" n", m.N(), len(win))
		assertAlmostEqual(t, label+" mean", m.Mean(), fmean(win), 14)
		assertAlmostEqual(t, label+" variance", m.Variance(), pvariance(win), 14)
	}
}

func TestCentralMomentsKleinKBN_RevertEmptyRaises(t *testing.T) {
	t.Parallel()

	m := NewCentralMomentsKleinKBN(1, true, true)
	assertPanics(t, "Cannot revert from an empty accumulator", func() { m.Revert(1.0) })
}

func TestCentralMomentsKleinKBN_StandardDeviationIsRealAfterRevert(t *testing.T) {
	t.Parallel()

	// Reverting to two equal samples can leave a tiny negative M2.
	m := NewCentralMomentsKleinKBN(0, true, true)
	for _, x := range []float64{0.1, 0.1, 0.7} {
		m.Update(x)
	}
	m.Revert(0.7)
	if math.IsNaN(m.StandardDeviation()) {
		t.Errorf("standard_deviation = NaN, want a real number")
	}
	if !(m.Variance() >= 0.0) {
		t.Errorf("variance = %v, want >= 0", m.Variance())
	}
	assertAlmostEqual(t, "standard_deviation", m.StandardDeviation(), 0.0, 15)
}

func TestCentralMomentsKleinKBN_Reset(t *testing.T) {
	t.Parallel()

	m := feedCentral(NewCentralMomentsKleinKBN(1, true, true), bacon)
	m.Reset()
	assertEqualInt(t, "n", m.N(), 0)
	assertEqual(t, "mean", m.Mean(), 0.0)
	assertNaN(t, "variance", m.Variance())
	feedCentral(m, []float64{1.0, 2.0, 3.0})
	assertAlmostEqual(t, "variance", m.Variance(), 1.0, 15)
}
