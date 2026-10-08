package performance

import (
	"math"
	"slices"

	"zpano/performance/core"
	"zpano/streamingkbn"
)

// DrawdownsCumulative returns the drawdown series of cumulative geometric
// returns,
//
//	D_t = W_t / max(W_0, W_1, ..., W_t) - 1
//
// where W_t is the cumulative wealth at observation t and W_0 = 1 is the
// starting wealth, so a first negative return already produces a drawdown
// (as in PerformanceAnalytics Drawdowns()). In a rolling window, W_0 is
// the wealth just before the first observation of the window.
//
// The result is a newly allocated slice.
func (m *Measures) DrawdownsCumulative() []float64 {
	return slices.Clone(m.drawdownHighWatermark.Drawdowns())
}

// MinDrawdownsCumulative returns the minimum (most negative) value of the
// cumulative drawdown series, the worst peak-to-valley decline. NaN when
// there are no observations.
func (m *Measures) MinDrawdownsCumulative() float64 {
	return m.drawdownHighWatermark.MaximumDrawdown()
}

// WorstDrawdownsCumulative returns the magnitude |min(D_t)| of the worst
// cumulative peak-to-valley drawdown. NaN when there are no observations.
func (m *Measures) WorstDrawdownsCumulative() float64 {
	return math.Abs(m.drawdownHighWatermark.MaximumDrawdown())
}

// DrawdownsHighWatermark returns the drawdown series measured from the
// high-water mark, D_t = W_t/H_t - 1 with H_t the highest wealth reached
// within the observation history or rolling window. Drawdowns are
// non-positive; 0 means the portfolio is at a high-water mark. This is
// the series used by the pain index, pain ratio, ulcer index and Martin
// ratio, and equals [Measures.DrawdownsCumulative].
//
// The result is a newly allocated slice.
func (m *Measures) DrawdownsHighWatermark() []float64 {
	return slices.Clone(m.drawdownHighWatermark.Drawdowns())
}

// DrawdownsContinuousRuns returns the drawdowns of continuous losing-return
// runs, one (negative) value for each maximal uninterrupted sequence of
// negative returns, oldest first. Following PerformanceAnalytics, the run
// drawdowns use R's percent convention (as in BurkeRatio):
//
//	DD_j = (Π(1 + r_i/100) - 1)·100
//
// If maxRuns > 0, the runs are sorted from most negative to least
// negative and truncated to the worst maxRuns. maxRuns <= 0 (Python None,
// the default) returns all runs unsorted.
//
// The result is a newly allocated slice (empty when there are no runs).
func (m *Measures) DrawdownsContinuousRuns(maxRuns int) []float64 {
	// One (negative) value per losing run.
	drawdowns := m.drawdownContinuousRuns.Drawdowns()
	if len(drawdowns) < 1 {
		return []float64{}
	}
	if maxRuns > 0 {
		// Ascending: most negative (worst) first.
		slices.Sort(drawdowns)
		// Keep only the worst maxRuns.
		if maxRuns < len(drawdowns) {
			drawdowns = drawdowns[:maxRuns]
		}
	}
	return drawdowns
}

// CalmarRatio returns the Calmar ratio R_g/|MDD|, the per-period geometric
// mean return divided by the worst cumulative peak-to-valley drawdown.
// Not annualized.
//
// NaN when the worst drawdown is zero or there are no returns.
func (m *Measures) CalmarRatio() float64 {
	wdd := m.WorstDrawdownsCumulative()
	if wdd == 0 {
		return math.NaN()
	}
	cagr := m.cumulativeReturn.GeometricMeanReturn()
	if math.IsNaN(cagr) {
		return math.NaN()
	}
	return cagr / wdd
}

// SterlingRatio returns the Sterling ratio R_g/(|MDD| + excess), the
// per-period geometric mean return divided by the worst cumulative
// drawdown plus the excess adjustment (Python default 0.1). Not
// annualized (PerformanceAnalytics scale = 1).
//
// NaN when the denominator is zero or there are no returns.
func (m *Measures) SterlingRatio(excess float64) float64 {
	wdd := m.WorstDrawdownsCumulative() + excess
	if wdd == 0 {
		return math.NaN()
	}
	cagr := m.cumulativeReturn.GeometricMeanReturn()
	if math.IsNaN(cagr) {
		return math.NaN()
	}
	return cagr / wdd
}

// BurkeRatio returns the Burke ratio (R_p - R_f)/√(Σ DD_j²), where R_p is
// the per-period geometric mean return, R_f the periodic risk-free rate
// and DD_j the drawdowns of the continuous losing runs (see
// [Measures.DrawdownsContinuousRuns], R's percent convention).
//
// NaN when there are no returns or no losing runs.
func (m *Measures) BurkeRatio() float64 {
	rate := m.cumulativeReturn.GeometricMeanReturn() - m.riskFreeRate
	if math.IsNaN(rate) {
		return math.NaN()
	}
	sqrtSumDrawdownsSquared := m.drawdownContinuousRuns.SqrtSumDrawdownsSquared()
	if sqrtSumDrawdownsSquared == 0 {
		return math.NaN()
	}
	return rate / sqrtSumDrawdownsSquared
}

// BurkeRatioModified returns the modified Burke ratio, the Burke ratio
// multiplied by √n. NaN when the Burke ratio is undefined.
func (m *Measures) BurkeRatioModified() float64 {
	burke := m.BurkeRatio()
	if math.IsNaN(burke) {
		return math.NaN()
	}
	return burke * math.Sqrt(float64(m.returnsKBN.N()))
}

// PainIndex returns the pain index -(1/n)·Σ D_t, the mean magnitude of
// the high-water-mark drawdowns. NaN when there are no observations.
func (m *Measures) PainIndex() float64 {
	// By construction all values are <= 0, so no abs() is needed.
	return -m.drawdownHighWatermark.DrawdownsMean()
}

// PainRatio returns (R_g - R_f)/PainIndex, the per-period geometric mean
// excess return over the risk-free rate divided by the pain index.
//
// NaN when there are no returns or the pain index is zero.
func (m *Measures) PainRatio() float64 {
	rate := m.cumulativeReturn.GeometricMeanReturn() - m.riskFreeRate
	if math.IsNaN(rate) {
		return math.NaN()
	}
	return divOrNaN(rate, m.PainIndex())
}

// UlcerIndex returns the ulcer index √((1/n)·Σ D_t²), the root mean square
// high-water-mark drawdown. NaN when there are no observations.
func (m *Measures) UlcerIndex() float64 {
	return math.Sqrt(m.drawdownHighWatermark.DrawdownsSquaredMean())
}

// MartinRatio returns (R_g - R_f)/UlcerIndex, the per-period geometric
// mean excess return over the risk-free rate divided by the ulcer index.
//
// NaN when there are no returns or the ulcer index is zero.
func (m *Measures) MartinRatio() float64 {
	rate := m.cumulativeReturn.GeometricMeanReturn() - m.riskFreeRate
	if math.IsNaN(rate) {
		return math.NaN()
	}
	return divOrNaN(rate, m.UlcerIndex())
}

// DrawdownAverage returns the average drawdown (ADD), the mean depth
// magnitude of the drawdown episodes (including an open episode). 0 when
// there are no episodes.
func (m *Measures) DrawdownAverage() float64 {
	return m.drawdownEpisodes.AverageEpisodeDrawdown()
}

// DrawdownAverageLength returns the average drawdown episode length in
// observations. 0 when there are no episodes.
func (m *Measures) DrawdownAverageLength() float64 {
	return m.drawdownEpisodes.AverageEpisodeLength()
}

// DrawdownAveragePeakToTrough returns the average drawdown episode
// peak-to-trough period in observations. 0 when there are no episodes.
func (m *Measures) DrawdownAveragePeakToTrough() float64 {
	return m.drawdownEpisodes.AverageEpisodePeakToTrough()
}

// DrawdownAverageRecovery returns the average drawdown episode recovery
// period in observations. 0 when there are no episodes.
func (m *Measures) DrawdownAverageRecovery() float64 {
	return m.drawdownEpisodes.AverageEpisodeRecovery()
}

// DrawdownDeviation returns the drawdown deviation √(Σ depth²/n), where
// the sum is over the episode depths and n is the number of observations.
// 0 when there are no observations.
func (m *Measures) DrawdownDeviation() float64 {
	return math.Sqrt(m.drawdownEpisodes.AverageEpisodeDrawdownSquared())
}

// CdarAverage returns the conditional drawdown at risk (CDaR) of the
// continuous high-water-mark drawdown path at the given confidence level
// (Python default 0.95): the mean magnitude of the drawdowns less than or
// equal to their (1 - confidence) linear percentile.
//
// Returns 0 when there are no observations or the percentile is not
// negative. Returns an error if confidence is not in (0, 1).
func (m *Measures) CdarAverage(confidence float64) (float64, error) {
	if !(0.0 < confidence && confidence < 1.0) {
		return math.NaN(), errConfidence
	}

	drawdowns := m.drawdownHighWatermark.Drawdowns()
	if len(drawdowns) == 0 {
		return 0, nil
	}

	q, err := core.Percentile(drawdowns, 1.0-confidence)
	if err != nil {
		return math.NaN(), err
	}
	if q >= 0.0 {
		return 0, nil
	}

	var tailSum streamingkbn.KleinKBNAccumulator
	tailLen := 0
	for _, dd := range drawdowns {
		if dd <= q {
			tailLen++
			tailSum.Update(dd)
		}
	}

	if tailLen > 0 {
		return -tailSum.Value() / float64(tailLen), nil
	}
	return 0, nil
}

// CdarDiscrete returns the conditional drawdown at risk (CDaR) of the
// discrete drawdown episodes at the given confidence level (Python
// default 0.95), the PerformanceAnalytics default: the mean magnitude of
// the episode depths less than or equal to their (1 - confidence) linear
// percentile.
//
// Returns 0 when there are no episodes. Returns an error if confidence is
// not in (0, 1).
func (m *Measures) CdarDiscrete(confidence float64) (float64, error) {
	if !(0.0 < confidence && confidence < 1.0) {
		return math.NaN(), errConfidence
	}

	depths := m.drawdownEpisodes.Depths()
	if len(depths) == 0 {
		return 0, nil
	}

	q, err := core.Percentile(depths, 1.0-confidence)
	if err != nil {
		return math.NaN(), err
	}

	var tailSum streamingkbn.KleinKBNAccumulator
	tailLen := 0
	for _, depth := range depths {
		if depth <= q {
			tailLen++
			tailSum.Update(depth)
		}
	}

	if tailLen > 0 {
		return -tailSum.Value() / float64(tailLen), nil
	}
	return 0, nil
}

// CdarBeta returns the conditional drawdown beta, the sensitivity of the
// portfolio returns to the worst benchmark drawdown episodes, following
// the legacy PerformanceAnalytics CDaR.beta definition:
//
//  1. Identify the discrete benchmark drawdown episodes.
//  2. Select the worst max(1, ceil(k·(1 - confidence))) of the k episodes;
//     the cutoff depth q is the depth of the last selected episode.
//  3. For each episode with depth <= q, compound the portfolio returns
//     from the beginning of the drawdown through its trough.
//  4. Divide the sum of these returns by (number of selected episodes)·q.
//
// The Python default confidence is 0.95.
//
// NaN when there are no benchmark episodes or q is zero. Returns an error
// if confidence is not in (0, 1).
func (m *Measures) CdarBeta(confidence float64) (float64, error) {
	if !(0.0 < confidence && confidence < 1.0) {
		return math.NaN(), errConfidence
	}

	w := m.returns

	episodes := m.drawdownEpisodesBenchmark.Episodes()
	depths := m.drawdownEpisodesBenchmark.Depths()
	if len(depths) == 0 {
		return math.NaN(), nil
	}

	tailCount := max(1, int(math.Ceil(float64(len(depths))*(1.0-confidence))))
	slices.Sort(depths)
	q := depths[tailCount-1]
	if q == 0.0 {
		return math.NaN(), nil
	}

	var sumRet, ret streamingkbn.KleinKBNAccumulator
	tailLen := 0
	for _, episode := range episodes {
		if episode.Depth <= q {
			tailLen++
			ret.Reset()
			for i := episode.FromIdx; i <= episode.TroughIdx; i++ {
				ret.Update(math.Log1p(w[i]))
			}
			sumRet.Update(math.Expm1(ret.Value()))
		}
	}

	if tailLen != 0 {
		return sumRet.Value() / (float64(tailLen) * q), nil
	}
	return math.NaN(), nil
}

// CdarAlpha returns the conditional drawdown alpha, the portfolio's
// annualized arithmetic mean return less the CDaR beta times the
// benchmark's annualized arithmetic mean return,
//
//	((1 + mean(r))^P - 1) - β_CDaR·((1 + mean(b))^P - 1)
//
// where P is periods per annum (PerformanceAnalytics CDaR.alpha hard-codes
// 12). The Python default confidence is 0.95.
//
// NaN when the CDaR beta is undefined. Returns an error if confidence is
// not in (0, 1).
func (m *Measures) CdarAlpha(confidence float64) (float64, error) {
	beta, err := m.CdarBeta(confidence)
	if err != nil {
		return math.NaN(), err
	}
	if math.IsNaN(beta) {
		return math.NaN(), nil
	}

	period := m.periodsPerAnnum

	rMean := m.returnsKBN.Mean()
	bMean := m.benchmarkReturnsKBN.Mean()

	rAnnual := math.Pow(1.0+rMean, period) - 1.0
	bAnnual := math.Pow(1.0+bMean, period) - 1.0

	return rAnnual - beta*bAnnual, nil
}

// RewardToConditionalDrawdown returns the per-period geometric mean return
// (not annualized) divided by the historical conditional drawdown, the
// mean magnitude of the worst max(1, int(n·(1 - confidence))) cumulative
// drawdowns. The Python default confidence is 0.95; it is not validated.
//
// NaN when there are no returns or the conditional drawdown is zero.
func (m *Measures) RewardToConditionalDrawdown(confidence float64) float64 {
	cagr := m.cumulativeReturn.GeometricMeanReturn()
	if math.IsNaN(cagr) {
		return math.NaN()
	}

	// Drawdowns from cumulative returns.
	dd := m.DrawdownsCumulative()
	if len(dd) < 1 {
		return math.NaN()
	}

	// Conditional drawdown: average of the worst (1-confidence) drawdowns.
	nTail := max(1, int(float64(len(dd))*(1-confidence)))
	slices.Sort(dd) // Most negative first.
	sortedTail := dd[:min(nTail, len(dd))]
	// Positive number.
	cdar := -pySum(sortedTail) / float64(len(sortedTail))

	return divOrNaN(cagr, cdar)
}
