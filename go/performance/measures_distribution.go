package performance

import (
	"math"

	"zpano/performance/core"
	"zpano/streamingkbn"
)

// AutocorrelationPenalty returns the autocorrelation penalty factor for
// serially correlated returns, based on the variance inflation adjustment
// of Andrew W. Lo ("The Statistics of Sharpe Ratios", 2002):
//
//	√(1 + 2·Σ_{k=1}^{q-1} (1 - k/q)·ρ_k)
//
// where ρ_k is the sample autocorrelation at lag k and q is
// min(periods per annum, n - 1) (Lo's recommended aggregation period).
//
// Divide a Sharpe or Sortino ratio by this value to obtain an
// autocorrelation-adjusted ratio. A value of 1 indicates no detected
// serial correlation, values above 1 positive autocorrelation, values
// below 1 negative autocorrelation.
//
// Returns 1 if there are fewer than two observations or the return
// variance is zero. The value depends on periods per annum. O(n·q).
func (m *Measures) AutocorrelationPenalty() float64 {
	n := m.returnsKBN.N()
	if n < 2 {
		return 1.0
	}
	mean := m.returnsKBN.Mean()
	denom := m.returnsKBN.VarianceDdof0() * float64(n)
	if math.IsNaN(denom) || denom == 0 {
		return 1.0
	}

	// Lo's recommended aggregation period:
	// daily 252, weekly 52, monthly 12, quarterly 4.
	q := int(math.Min(m.periodsPerAnnum, float64(n-1)))

	w := m.returns
	s := 0.0
	for k := 1; k < q; k++ {
		numer := 0.0
		for t := k; t < n; t++ {
			numer += (w[t] - mean) * (w[t-k] - mean)
		}
		rho := numer / denom
		s += (1.0 - float64(k)/float64(q)) * rho
	}

	return math.Sqrt(math.Max(0.0, 1.0+2.0*s))
}

// CumulativeGeometricReturn returns the cumulative geometric return
// Π(1 + r) - 1 (0 when there are no returns).
func (m *Measures) CumulativeGeometricReturn() float64 {
	return m.cumulativeReturn.CumulativeGeometricReturn()
}

// GeometricMeanReturn returns the geometric mean return per observation:
// the constant per-period return that would produce the same cumulative
// growth. NaN when there are no returns.
func (m *Measures) GeometricMeanReturn() float64 {
	return m.cumulativeReturn.GeometricMeanReturn()
}

// CompoundAnnualGrowthRate returns the compound annual growth rate (CAGR),
// the geometric mean return annualized with periods per annum. NaN when
// there are no returns.
func (m *Measures) CompoundAnnualGrowthRate() float64 {
	return m.cumulativeReturn.AnnualizedGeometricMeanReturn(m.periodsPerAnnum)
}

// Skewness returns the scipy-default ('moment', bias=True) skewness of the
// returns, the degree of asymmetry of the distribution around its mean.
func (m *Measures) Skewness() float64 {
	return m.returnsKBN.Skewness()
}

// SkewnessMoment returns the 'moment' skewness g₁ = μ₃/μ₂^(3/2) of the
// returns.
func (m *Measures) SkewnessMoment() float64 {
	return m.returnsKBN.SkewnessMoment()
}

// SkewnessFisher returns the 'fisher' skewness g₁·√(n(n-1))/(n-2) of the
// returns (NaN for n < 3).
func (m *Measures) SkewnessFisher() float64 {
	return m.returnsKBN.SkewnessFisher()
}

// SkewnessSample returns the 'sample' skewness g₁·n²/((n-1)(n-2)) of the
// returns (NaN for n < 3).
func (m *Measures) SkewnessSample() float64 {
	return m.returnsKBN.SkewnessSample()
}

// Kurtosis returns the scipy-default biased excess kurtosis of the
// returns, the degree to which the distribution peaks compared to a
// normal distribution.
func (m *Measures) Kurtosis() float64 {
	return m.returnsKBN.Kurtosis()
}

// KurtosisExcess returns the biased excess kurtosis β₂ - 3 of the returns,
// where β₂ = μ₄/μ₂².
func (m *Measures) KurtosisExcess() float64 {
	return m.returnsKBN.KurtosisExcess()
}

// KurtosisMoment returns the biased Pearson (population) kurtosis
// β₂ = μ₄/μ₂² of the returns.
func (m *Measures) KurtosisMoment() float64 {
	return m.returnsKBN.KurtosisMoment()
}

// KurtosisSampleExcess returns the unbiased excess kurtosis
// ((n²-1)β₂ - 3(n-1)²)/((n-2)(n-3)) of the returns (NaN for n < 4).
func (m *Measures) KurtosisSampleExcess() float64 {
	return m.returnsKBN.KurtosisSampleExcess()
}

// KurtosisSampleCorrected returns the PerformanceAnalytics-compatible
// 'sample' kurtosis (n²-1)β₂/((n-2)(n-3)) of the returns.
func (m *Measures) KurtosisSampleCorrected() float64 {
	return m.returnsKBN.KurtosisSampleCorrected()
}

// KurtosisSample returns the unbiased Pearson kurtosis
// ((n²-1)β₂ - 3(n-1)²)/((n-2)(n-3)) + 3 of the returns.
func (m *Measures) KurtosisSample() float64 {
	return m.returnsKBN.KurtosisSample()
}

// SkewnessKurtosisRatio returns the ratio g₁/β₂ of the population
// ('moment') skewness to the population (Pearson, non-excess) kurtosis.
//
// Positive values indicate that positive asymmetry dominates relative to
// tail heaviness. Used in conjunction with the Sharpe ratio to rank
// portfolios; higher is better. NaN when the kurtosis is zero.
func (m *Measures) SkewnessKurtosisRatio() float64 {
	s := m.returnsKBN.SkewnessMoment()
	k := m.returnsKBN.KurtosisMoment()
	return divOrNaN(s, k)
}

// JarqueBeraNormalityTestStatistic returns the Jarque–Bera normality test
// statistic
//
//	JB = n/6·(g₁² + (β₂-3)²/4)
//
// using the population skewness and excess kurtosis. Under the normality
// null hypothesis JB is asymptotically χ²(2) distributed (critical values
// 4.61, 5.99, 9.21 at 90%, 95%, 99% confidence); normality is rejected
// when JB exceeds the critical value. NaN when the moments are
// unavailable.
func (m *Measures) JarqueBeraNormalityTestStatistic() float64 {
	// Population skewness and excess kurtosis (not sample corrected),
	// matching scipy.stats skew and kurtosis with bias=True, fisher=True.
	s := m.returnsKBN.SkewnessMoment()
	k := m.returnsKBN.KurtosisExcess()
	if math.IsNaN(s) || math.IsNaN(k) {
		return math.NaN()
	}
	// Bera-Jarque formula (Equation 5.17 from Bacon 3rd ed).
	n := m.returnsKBN.N()
	return (float64(n) / 6) * (s*s + (k*k)/4)
}

// IsNormalDistribution tests the null hypothesis that the returns are
// normally distributed using the Jarque–Bera test at the given confidence
// level in (0, 1) (Python default 0.95).
//
// Returns true if normality cannot be rejected, false if it is rejected
// or there is insufficient data to perform the test. Returns an error if
// confidence is not in (0, 1), including NaN, even without observations.
func (m *Measures) IsNormalDistribution(confidence float64) (bool, error) {
	return isNormalFromJB(m.JarqueBeraNormalityTestStatistic(), confidence)
}

// isNormalFromJB is the decision rule of IsNormalDistribution for a given
// Jarque–Bera statistic jb.
func isNormalFromJB(jb, confidence float64) (bool, error) {
	if !(0 < confidence && confidence < 1) {
		return false, errConfidence
	}
	if math.IsNaN(jb) {
		return false, nil
	}

	// The χ² distribution with 2 degrees of freedom is the exponential
	// distribution with λ = 0.5, so its inverse CDF has the closed form
	// chi2.ppf(p, 2) = -2·ln(1 - p):
	// -2·ln(1 - 0.95) = 5.991464547107982,
	// -2·ln(1 - 0.99) = 9.210340371976184.
	critical := -2.0 * math.Log1p(-confidence)
	return jb <= critical, nil
}

// VarHistorical returns the historical value at risk (VaR) at the given
// confidence level in (0, 1) (Python default 0.95), the negated
// (1 - confidence) percentile of the observed returns.
//
// The result is the estimated loss as a positive value; it is negative
// ("inverse risk") when the lower-tail percentile is positive. NaN when
// there are no returns or the confidence level is invalid. O(n log n).
func (m *Measures) VarHistorical(confidence float64) float64 {
	return core.VarHistorical(m.returns, 0.0, confidence)
}

// VarGaussian returns the Gaussian (parametric) value at risk
// -(μ + z·σ) at the given confidence level in (0, 1) (Python default
// 0.95), where z is the standard normal (1 - confidence) quantile and σ
// the population standard deviation.
//
// A negative result indicates "inverse risk". NaN when the standard
// deviation is unavailable or the confidence level is invalid.
func (m *Measures) VarGaussian(confidence float64) float64 {
	return core.VarGaussian(m.returnsKBN, confidence)
}

// VarCornishFisher returns the modified Cornish-Fisher value at risk at
// the given confidence level in (0, 1) (Python default 0.95): the
// Gaussian VaR with the normal quantile adjusted for the observed
// skewness and excess kurtosis (falling back to the Gaussian VaR when
// these are unavailable).
//
// A negative result indicates "inverse risk". NaN when the standard
// deviation is unavailable or the confidence level is invalid.
func (m *Measures) VarCornishFisher(confidence float64) float64 {
	return core.VarCornishFisher(m.returnsKBN, confidence)
}

// EsHistorical returns the historical expected shortfall (ES) at the given
// confidence level in (0, 1) (Python default 0.95): the negated mean of
// all returns less than or equal to the historical VaR threshold.
//
// The result may be negative when all observed returns are positive
// ("inverse risk"). NaN when there are no returns or the confidence level
// is invalid. O(n log n).
func (m *Measures) EsHistorical(confidence float64) float64 {
	return core.EsHistorical(m.returns, 0.0, confidence)
}

// EsGaussian returns the Gaussian expected shortfall at the given
// confidence level in (0, 1) (Python default 0.95), assuming normally
// distributed returns with the sample mean and population standard
// deviation.
//
// NaN when the standard deviation is unavailable or the confidence level
// is invalid.
func (m *Measures) EsGaussian(confidence float64) float64 {
	return core.EsGaussian(m.returnsKBN, confidence)
}

// EsCornishFisher returns the Cornish-Fisher expected shortfall at the
// given confidence level in (0, 1) (Python default 0.95), adjusting the
// Gaussian ES for the sample skewness and excess kurtosis.
//
// NaN when the standard deviation is unavailable or the confidence level
// is invalid.
func (m *Measures) EsCornishFisher(confidence float64) float64 {
	return core.EsCornishFisher(m.returnsKBN, confidence)
}

// RewardToVarRatioHistorical returns the mean excess return over the
// risk-free rate divided by the historical VaR of the raw returns at the
// given confidence level (Python default 0.95).
//
// NaN when the VaR is zero or NaN (including an invalid confidence level).
func (m *Measures) RewardToVarRatioHistorical(confidence float64) float64 {
	return divOrNaN(m.excessReturnsKBN.Mean(), m.VarHistorical(confidence))
}

// RewardToVarRatioGaussian returns the mean excess return over the
// risk-free rate divided by the Gaussian VaR of the raw returns at the
// given confidence level (Python default 0.95).
//
// NaN when the VaR is zero or NaN (including an invalid confidence level).
func (m *Measures) RewardToVarRatioGaussian(confidence float64) float64 {
	return divOrNaN(m.excessReturnsKBN.Mean(), m.VarGaussian(confidence))
}

// RewardToVarRatioCornishFisher returns the mean excess return over the
// risk-free rate divided by the Cornish-Fisher VaR of the raw returns at
// the given confidence level (Python default 0.95).
//
// NaN when the VaR is zero or NaN (including an invalid confidence level).
func (m *Measures) RewardToVarRatioCornishFisher(confidence float64) float64 {
	return divOrNaN(m.excessReturnsKBN.Mean(), m.VarCornishFisher(confidence))
}

// RewardToEsRatioHistorical returns the mean excess return over the
// risk-free rate divided by the historical ES of the raw returns at the
// given confidence level (Python default 0.95).
//
// NaN when the ES is zero or NaN (including an invalid confidence level).
func (m *Measures) RewardToEsRatioHistorical(confidence float64) float64 {
	return divOrNaN(m.excessReturnsKBN.Mean(), m.EsHistorical(confidence))
}

// RewardToEsRatioGaussian returns the mean excess return over the
// risk-free rate divided by the Gaussian ES of the raw returns at the
// given confidence level (Python default 0.95).
//
// NaN when the ES is zero or NaN (including an invalid confidence level).
func (m *Measures) RewardToEsRatioGaussian(confidence float64) float64 {
	return divOrNaN(m.excessReturnsKBN.Mean(), m.EsGaussian(confidence))
}

// RewardToEsRatioCornishFisher returns the mean excess return over the
// risk-free rate divided by the Cornish-Fisher ES of the raw returns at
// the given confidence level (Python default 0.95).
//
// NaN when the ES is zero or NaN (including an invalid confidence level).
func (m *Measures) RewardToEsRatioCornishFisher(confidence float64) float64 {
	return divOrNaN(m.excessReturnsKBN.Mean(), m.EsCornishFisher(confidence))
}

// MeanAbsoluteDeviationRatio returns the mean return divided by the mean
// absolute deviation Σ|r - mean|/n of the returns from their sample mean.
//
// NaN when there are no returns or the mean absolute deviation is zero.
// O(n).
func (m *Measures) MeanAbsoluteDeviationRatio() float64 {
	n := m.returnsKBN.N()
	if n < 1 {
		return math.NaN()
	}
	mean := m.returnsKBN.Mean()

	var sum streamingkbn.KleinKBNAccumulator
	for _, x := range m.returns {
		sum.Update(math.Abs(x - mean))
	}

	mad := sum.Value() / float64(n)
	if mad > 0 {
		return mean / mad
	}
	return math.NaN()
}
