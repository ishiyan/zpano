package performance

import (
	"math"

	"zpano/performance/core"
	"zpano/streamingkbn"
)

// TailRatio returns the tail ratio Q(cutoff)/|Q(1 - cutoff)|, the upper
// percentile of the returns divided by the magnitude of the corresponding
// lower percentile (linear percentiles). It describes the tail asymmetry
// of the return distribution. Not available in PerformanceAnalytics. The
// Python default cutoff is 0.95. Not annualized; O(n log n).
//
// NaN with fewer than two observations or a zero lower percentile.
// Returns an error if cutoff is not in (0.5, 1).
func (m *Measures) TailRatio(cutoff float64) (float64, error) {
	if !(0.5 < cutoff && cutoff < 1.0) {
		return math.NaN(), errCutoff
	}
	w := m.returns
	if len(w) < 2 {
		return math.NaN(), nil
	}

	rightTail, err := core.Percentile(w, cutoff)
	if err != nil {
		return math.NaN(), err
	}
	leftTail, err := core.Percentile(w, 1-cutoff)
	if err != nil {
		return math.NaN(), err
	}
	if leftTail != 0 {
		return rightTail / math.Abs(leftTail), nil
	}
	return math.NaN(), nil
}

// KellyRatioFull returns the full Kelly criterion fraction
// E[R - R_f]/Var(R - R_f) based on the mean and sample variance (ddof=1)
// of the excess returns. A value of 1 corresponds to 100% capital
// allocation; values above 1 imply leverage.
//
// NaN when the variance of the excess returns is zero or undefined.
func (m *Measures) KellyRatioFull() float64 {
	meanExcess := m.excessReturnsKBN.Mean()
	varExcess := m.excessReturnsKBN.Variance()
	return divOrNaN(meanExcess, varExcess)
}

// KellyRatio returns the half-Kelly criterion fraction, one half of
// [Measures.KellyRatioFull], a more conservative alternative to full
// Kelly.
func (m *Measures) KellyRatio() float64 {
	return m.KellyRatioFull() / 2
}

// HurstExponent returns the single-scale rescaled range (R/S) estimate of
// the Hurst exponent, log(R/S)/log(N), where S is the sample standard
// deviation of the returns and R the range of the cumulative demeaned
// returns. H > 0.5 indicates persistence, H ≈ 0.5 a random walk and
// H < 0.5 anti-persistence. Not annualized; O(n).
//
// NaN with fewer than two observations, zero volatility or a
// non-positive rescaled range.
func (m *Measures) HurstExponent() float64 {
	n := m.returnsKBN.N()
	if n < 2 {
		return math.NaN()
	}

	mean := m.returnsKBN.Mean()
	std := m.returnsKBN.StandardDeviationDdof1()
	if std == 0 {
		return math.NaN()
	}
	var cumSum streamingkbn.KleinKBNAccumulator
	cumMin := math.Inf(1)
	cumMax := math.Inf(-1)
	for _, x := range m.returns {
		cumSum.Update(x - mean) // Demean.
		val := cumSum.Value()
		if cumMin > val {
			cumMin = val
		}
		if cumMax < val {
			cumMax = val
		}
	}
	delta := cumMax - cumMin
	rescaledRange := delta / std
	if rescaledRange <= 0 {
		return math.NaN()
	}
	return math.Log(rescaledRange) / math.Log(float64(n))
}

// BiasRatio returns the bias ratio (Abdulali, 2006), a diagnostic for
// stale pricing and return smoothing,
//
//	count(0 <= r <= t) / (1 + count(-t <= r < 0)),  t = k·σ
//
// where σ is the sample standard deviation of the returns and k the
// stdDevMultiplier (Python default 1). Recomputed over the window (O(n)).
//
// NaN when the standard deviation is undefined or zero. Returns an error
// if stdDevMultiplier is not positive.
func (m *Measures) BiasRatio(stdDevMultiplier float64) (float64, error) {
	if !(stdDevMultiplier > 0) {
		return math.NaN(), errStdDevMultiplier
	}

	std := m.returnsKBN.StandardDeviationDdof1()
	if math.IsNaN(std) || std == 0 {
		return math.NaN(), nil
	}
	threshold := stdDevMultiplier * std

	countPositive := 0
	countNegative := 0
	for _, x := range m.returns {
		if 0 <= x && x <= threshold {
			countPositive++
		} else if -threshold <= x && x < 0 {
			countNegative++
		}
	}

	return float64(countPositive) / float64(1+countNegative), nil
}

// KRatio returns the K-ratio (Lars Kestner), slope/(SE(slope)·√N) of the
// least-squares regression of the cumulative log return Σlog(1 + r) on
// time. It measures the consistency of the equity curve growth. Not
// annualized; recomputed over the window (O(n), plain sums).
//
// NaN with fewer than three observations or a zero slope standard error.
func (m *Measures) KRatio() float64 {
	n := m.returnsKBN.N()
	if n < 3 {
		return math.NaN()
	}
	fn := float64(n)

	// Equity curve (cumulative log returns).
	equity := make([]float64, 0, len(m.returns))
	cumSum := 0.0
	for _, x := range m.returns {
		cumSum += math.Log1p(x)
		equity = append(equity, cumSum)
	}

	// Linear regression statistics equity = a + b·t, t = 0, 1, ..., n-1.
	sumT := 0.0
	sumT2 := 0.0
	sumEq := 0.0
	sumTE := 0.0
	for i, eqVal := range equity {
		tVal := float64(i)
		sumT += tVal
		sumT2 += tVal * tVal
		sumEq += eqVal
		sumTE += tVal * eqVal
	}

	tMean := sumT / fn
	equityMean := sumEq / fn

	// S_tt = Σ(t - t_mean)² = Σt² - n·t_mean².
	sTT := sumT2 - fn*(tMean*tMean)

	// S_te = Σ(t - t_mean)(eq - eq_mean) = Σt·eq - n·t_mean·eq_mean.
	sTE := sumTE - fn*tMean*equityMean
	if sTT == 0 {
		return math.NaN()
	}

	slope := sTE / sTT

	intercept := equityMean - slope*tMean

	// Residual variance Σ(eq - (a + b·t))²/(n - 2).
	sumSqResiduals := 0.0
	for i, eqVal := range equity {
		tVal := float64(i)
		predicted := intercept + slope*tVal
		residual := eqVal - predicted
		sumSqResiduals += residual * residual
	}
	residualVar := sumSqResiduals / float64(n-2)

	// Standard error of the slope.
	if residualVar < 0 { // Handle floating point noise.
		residualVar = 0.0
	}
	seSlope := math.Sqrt(residualVar / sTT)
	if seSlope == 0 {
		return math.NaN()
	}

	return slope / (seSlope * math.Sqrt(fn))
}

// GainToPainRatio returns Jack Schwager's net sum of returns divided by
// the sum of loss magnitudes: Σr / Σmax(-r, 0).
// NaN when there are no losses.
func (m *Measures) GainToPainRatio() float64 {
	lpm1 := m.rawPartialMoments.LowerPartialMoment1()
	if math.IsNaN(lpm1) || lpm1 == 0 {
		return math.NaN()
	}
	return m.returnsKBN.X1Sum() / lpm1
}
