package performance

import (
	"math"
	"slices"

	"zpano/performance/core"
	"zpano/streamingkbn"
)

// UpsidePotentialRatio returns the first-order upper partial moment about
// the MAR divided by the square root of the second-order lower partial
// moment about the MAR, HPM₁/√LPM₂ (both normalized by the total number of
// observations).
//
// NaN when the partial moments are unavailable or the downside risk is
// zero.
func (m *Measures) UpsidePotentialRatio() float64 {
	hpm1 := m.targetPartialMoments.HigherPartialMoment1()
	lpm2 := m.targetPartialMoments.LowerPartialMoment2()
	if math.IsNaN(hpm1) || math.IsNaN(lpm2) || lpm2 == 0 {
		return math.NaN()
	}
	return hpm1 / math.Sqrt(lpm2)
}

// UpsidePotentialRatioSubset returns the subset upside potential ratio:
// like [Measures.UpsidePotentialRatio], but the upside potential is
// normalized by the number of observations above the MAR and the downside
// risk by the number of observations below the MAR.
//
// NaN when there are no upside or no downside observations, or the
// downside risk is zero.
func (m *Measures) UpsidePotentialRatioSubset() float64 {
	n1 := m.targetPartialMoments.UpperExcessCount()
	n2 := m.targetPartialMoments.LowerExcessCount()
	if n1 == 0 || n2 == 0 {
		return math.NaN()
	}
	hpm1 := m.targetPartialMoments.UpperExcessMoment1Sum() / float64(n1)
	lpm2 := m.targetPartialMoments.LowerExcessMoment2Sum() / float64(n2)
	if math.IsNaN(hpm1) || math.IsNaN(lpm2) || lpm2 == 0 {
		return math.NaN()
	}
	return hpm1 / math.Sqrt(lpm2)
}

// UpsideFrequency returns the proportion of returns above the MAR, in
// [0, 1].
func (m *Measures) UpsideFrequency() float64 {
	return m.targetPartialMoments.UpsideFrequency()
}

// UpsidePotential returns the first-order upper partial moment about the
// MAR (the average upside above the MAR, observations below the MAR
// contributing zero). NaN when there are no returns.
func (m *Measures) UpsidePotential() float64 {
	return m.targetPartialMoments.HigherPartialMoment1()
}

// UpsidePotentialSubset returns the average excess return above the MAR,
// normalized only by the observations exceeding the MAR. 0 when no
// observation exceeds the MAR.
func (m *Measures) UpsidePotentialSubset() float64 {
	n := m.targetPartialMoments.UpperExcessCount()
	if n > 0 {
		return m.targetPartialMoments.UpperExcessMoment1Sum() / float64(n)
	}
	return 0
}

// UpsideVariance returns the second-order upper partial moment about the
// MAR, normalized by the total number of observations. NaN when there are
// no returns.
func (m *Measures) UpsideVariance() float64 {
	return m.targetPartialMoments.HigherPartialMoment2()
}

// UpsideVarianceSubset returns the second-order upper partial moment about
// the MAR, normalized only by the observations exceeding the MAR. 0 when
// no observation exceeds the MAR.
func (m *Measures) UpsideVarianceSubset() float64 {
	n := m.targetPartialMoments.UpperExcessCount()
	if n > 0 {
		return m.targetPartialMoments.UpperExcessMoment2Sum() / float64(n)
	}
	return 0
}

// UpsideRisk returns the square root of the upside variance. NaN when the
// upside variance is unavailable.
func (m *Measures) UpsideRisk() float64 {
	variance := m.UpsideVariance()
	if math.IsNaN(variance) {
		return math.NaN()
	}
	return math.Sqrt(variance)
}

// UpsideRiskSubset returns the square root of the subset upside variance.
func (m *Measures) UpsideRiskSubset() float64 {
	variance := m.UpsideVarianceSubset()
	if math.IsNaN(variance) {
		return math.NaN()
	}
	return math.Sqrt(variance)
}

// SemiDeviation returns the semi-deviation of the returns: the square root
// of the second-order lower partial moment about the sample mean,
// normalized by the total number of observations.
//
// NaN when there are no returns. O(n).
func (m *Measures) SemiDeviation() float64 {
	n := m.returnsKBN.N()
	if n == 0 {
		return math.NaN()
	}

	mean := m.returnsKBN.Mean()

	// Accumulate squared negative deviations from the mean using a
	// compensated floating-point summation.
	var sumSquared streamingkbn.KleinKBNAccumulator
	for _, r := range m.returns {
		deviation := r - mean
		if deviation < 0 {
			sumSquared.Update(deviation * deviation)
		}
	}
	// Divide by the "full" window length.
	return math.Sqrt(sumSquared.Value() / float64(n))
}

// DownsideDeviation returns the downside deviation of the returns: the
// square root of the second-order lower partial moment about the MAR,
// normalized by the total number of observations.
//
// NaN when there are no returns.
func (m *Measures) DownsideDeviation() float64 {
	denom := m.targetPartialMoments.TotalCount()
	if denom == 0 {
		return math.NaN()
	}
	return math.Sqrt(m.targetPartialMoments.LowerExcessMoment2Sum() / float64(denom))
}

// DownsideDeviationSubset returns the downside deviation normalized only
// by the observations below the MAR. 0 when no observation falls below
// the MAR.
func (m *Measures) DownsideDeviationSubset() float64 {
	denom := m.targetPartialMoments.LowerExcessCount()
	if denom == 0 {
		return 0
	}
	return math.Sqrt(m.targetPartialMoments.LowerExcessMoment2Sum() / float64(denom))
}

// DownsideFrequency returns the proportion of returns below the MAR, in
// [0, 1].
func (m *Measures) DownsideFrequency() float64 {
	return m.targetPartialMoments.DownsideFrequency()
}

// DownsidePotential returns the average shortfall below the MAR (the
// first-order lower partial moment). NaN when there are no returns.
func (m *Measures) DownsidePotential() float64 {
	return m.targetPartialMoments.DownsidePotential()
}

// SharpeRatio returns the mean excess return over the risk-free rate
// divided by the sample standard deviation (ddof=1) of the excess
// returns.
//
// NaN with fewer than two observations or a zero standard deviation.
func (m *Measures) SharpeRatio() float64 {
	std := m.excessReturnsKBN.StandardDeviationDdof1()
	if math.IsNaN(std) || std == 0 {
		return math.NaN()
	}
	return m.excessReturnsKBN.Mean() / std
}

// sharpeRatioDenominated returns the mean excess return divided by denom,
// or NaN with fewer than two observations or a zero or NaN denom.
func (m *Measures) sharpeRatioDenominated(denom func() float64) float64 {
	if m.excessReturnsKBN.N() < 2 {
		return math.NaN()
	}
	d := denom()
	if math.IsNaN(d) || d == 0 {
		return math.NaN()
	}
	return m.excessReturnsKBN.Mean() / d
}

// SharpeRatioVarHistorical returns the modified Sharpe ratio using the
// historical VaR of the excess returns at the given confidence level
// (Python default 0.95) as the measure of risk.
//
// NaN with fewer than two observations, a zero VaR or an invalid
// confidence level. O(n log n).
func (m *Measures) SharpeRatioVarHistorical(confidence float64) float64 {
	return m.sharpeRatioDenominated(func() float64 {
		return core.VarHistorical(m.returns, m.riskFreeRate, confidence)
	})
}

// SharpeRatioVarGaussian returns the modified Sharpe ratio using the
// Gaussian VaR of the excess returns at the given confidence level
// (Python default 0.95) as the measure of risk.
//
// NaN with fewer than two observations, a zero VaR or an invalid
// confidence level.
func (m *Measures) SharpeRatioVarGaussian(confidence float64) float64 {
	return m.sharpeRatioDenominated(func() float64 {
		return core.VarGaussian(m.excessReturnsKBN, confidence)
	})
}

// SharpeRatioVarCornishFisher returns the modified Sharpe ratio using the
// Cornish-Fisher VaR of the excess returns at the given confidence level
// (Python default 0.95) as the measure of risk.
//
// NaN with fewer than two observations, a zero VaR or an invalid
// confidence level.
func (m *Measures) SharpeRatioVarCornishFisher(confidence float64) float64 {
	return m.sharpeRatioDenominated(func() float64 {
		return core.VarCornishFisher(m.excessReturnsKBN, confidence)
	})
}

// SharpeRatioEsHistorical returns the modified Sharpe ratio using the
// historical expected shortfall of the excess returns at the given
// confidence level (Python default 0.95) as the measure of risk.
//
// NaN with fewer than two observations, a zero ES or an invalid
// confidence level. O(n log n).
func (m *Measures) SharpeRatioEsHistorical(confidence float64) float64 {
	return m.sharpeRatioDenominated(func() float64 {
		return core.EsHistorical(m.returns, m.riskFreeRate, confidence)
	})
}

// SharpeRatioEsGaussian returns the modified Sharpe ratio using the
// Gaussian expected shortfall of the excess returns at the given
// confidence level (Python default 0.95) as the measure of risk.
//
// NaN with fewer than two observations, a zero ES or an invalid
// confidence level.
func (m *Measures) SharpeRatioEsGaussian(confidence float64) float64 {
	return m.sharpeRatioDenominated(func() float64 {
		return core.EsGaussian(m.excessReturnsKBN, confidence)
	})
}

// SharpeRatioEsCornishFisher returns the modified Sharpe ratio using the
// Cornish-Fisher expected shortfall of the excess returns at the given
// confidence level (Python default 0.95) as the measure of risk.
//
// NaN with fewer than two observations, a zero ES or an invalid
// confidence level.
func (m *Measures) SharpeRatioEsCornishFisher(confidence float64) float64 {
	return m.sharpeRatioDenominated(func() float64 {
		return core.EsCornishFisher(m.excessReturnsKBN, confidence)
	})
}

// DownsideSharpeRatio returns the symmetric downside-risk Sharpe ratio
// (Ziemba, 2005): the mean excess return divided by √2 times the
// semi-deviation of the returns about their sample mean.
//
// NaN when there are no returns; ±Inf when the semi-deviation is zero
// (-Inf if the mean excess return is negative, +Inf otherwise).
func (m *Measures) DownsideSharpeRatio() float64 {
	semiDev := m.SemiDeviation()
	if math.IsNaN(semiDev) {
		return math.NaN()
	}
	if semiDev == 0 {
		if m.excessReturnsKBN.Mean() < 0 {
			return math.Inf(-1)
		}
		return math.Inf(1)
	}
	return m.excessReturnsKBN.Mean() / (sqrt2 * semiDev)
}

// AdjustedSharpeRatio returns the adjusted Sharpe ratio (Pezier & White,
// 2006), SR·(1 + S·SR/6 - K·SR²/24), where S is the skewness and K the
// excess kurtosis of the raw returns and SR the Sharpe ratio of the
// excess returns.
//
// NaN when the moments or the Sharpe ratio are unavailable.
func (m *Measures) AdjustedSharpeRatio() float64 {
	skewness := m.returnsKBN.SkewnessMoment()
	kurtosis := m.returnsKBN.KurtosisExcess()
	if math.IsNaN(skewness) || math.IsNaN(kurtosis) {
		return math.NaN()
	}
	sr := m.SharpeRatio()
	if math.IsNaN(sr) {
		return math.NaN()
	}
	return sr * (1 + skewness*sr/6 - kurtosis*sr*sr/24)
}

// AdjustedSharpeRatioSkewOnly returns the skewness-only adjusted Sharpe
// ratio SR·(1 + S·SR/6).
//
// NaN when the skewness or the Sharpe ratio are unavailable.
func (m *Measures) AdjustedSharpeRatioSkewOnly() float64 {
	s := m.returnsKBN.SkewnessMoment()
	if math.IsNaN(s) {
		return math.NaN()
	}
	sr := m.SharpeRatio()
	if math.IsNaN(sr) {
		return math.NaN()
	}
	return sr * (1 + (s*sr)/6)
}

// ProbabilisticSharpeRatio returns the probabilistic Sharpe ratio (PSR)
// with the sample skewness and normal kurtosis K = 3: the probability
// that the true Sharpe ratio exceeds referenceSR (Python default 0),
//
//	Φ((SR - SR*)·√(n-1) / √(1 - S·SR + (K-1)/4·SR²))
//
// (Bailey & López de Prado 2014, Opdyke 2007). NaN when it cannot be
// estimated.
func (m *Measures) ProbabilisticSharpeRatio(referenceSR float64) float64 {
	return core.ProbabilisticSharpeRatio(m.returnsKBN, m.SharpeRatio(), referenceSR, false, true)
}

// ProbabilisticSharpeRatioFull returns the probabilistic Sharpe ratio
// with the sample skewness and the sample (non-excess) kurtosis, for the
// reference Sharpe ratio referenceSR (Python default 0).
func (m *Measures) ProbabilisticSharpeRatioFull(referenceSR float64) float64 {
	return core.ProbabilisticSharpeRatio(m.returnsKBN, m.SharpeRatio(), referenceSR, false, false)
}

// ProbabilisticSharpeRatioSymmetric returns the probabilistic Sharpe ratio
// with zero skewness and the sample kurtosis, for the reference Sharpe
// ratio referenceSR (Python default 0).
func (m *Measures) ProbabilisticSharpeRatioSymmetric(referenceSR float64) float64 {
	return core.ProbabilisticSharpeRatio(m.returnsKBN, m.SharpeRatio(), referenceSR, true, false)
}

// ProbabilisticSharpeRatioGaussian returns the probabilistic Sharpe ratio
// under Gaussian moments (zero skewness, K = 3), for the reference Sharpe
// ratio referenceSR (Python default 0).
func (m *Measures) ProbabilisticSharpeRatioGaussian(referenceSR float64) float64 {
	return core.ProbabilisticSharpeRatio(m.returnsKBN, m.SharpeRatio(), referenceSR, true, true)
}

// SortinoRatio returns the arithmetic mean excess return over the MAR
// divided by the downside deviation √LPM₂ about the MAR.
//
// NaN when the downside partial moment is unavailable or zero.
func (m *Measures) SortinoRatio() float64 {
	lpm2 := m.targetPartialMoments.LowerPartialMoment2()
	if math.IsNaN(lpm2) || lpm2 == 0 {
		return math.NaN()
	}
	return m.targetReturnsKBN.Mean() / math.Sqrt(lpm2)
}

// SortinoRatioSqrt2 returns the adjusted Sortino ratio proposed by Jack
// Schwager, the Sortino ratio divided by √2, which makes it more directly
// comparable with the Sharpe ratio.
func (m *Measures) SortinoRatioSqrt2() float64 {
	return m.SortinoRatio() / sqrt2
}

// SortinoSatchellRatio returns the Sortino-Satchell ratio, the arithmetic
// mean excess return over the MAR divided by √LPM₂. It equals
// [Measures.SortinoRatio].
func (m *Measures) SortinoSatchellRatio() float64 {
	lpm2 := m.targetPartialMoments.LowerPartialMoment2()
	if math.IsNaN(lpm2) || lpm2 == 0 {
		return math.NaN()
	}
	return m.targetReturnsKBN.Mean() / math.Sqrt(lpm2)
}

// OmegaRatio returns the Omega ratio relative to the MAR computed with the
// simple empirical method, mean(r - MAR)/LPM₁ + 1 (= HPM₁/LPM₁).
//
// NaN when the lower partial moment is unavailable or zero.
func (m *Measures) OmegaRatio() float64 {
	lpm1 := m.targetPartialMoments.LowerPartialMoment1()
	if math.IsNaN(lpm1) || lpm1 == 0 {
		return math.NaN()
	}
	return m.targetReturnsKBN.Mean()/lpm1 + 1
}

// OmegaSharpeRatio returns the Omega-Sharpe ratio, the Omega ratio minus
// one.
func (m *Measures) OmegaSharpeRatio() float64 {
	return m.OmegaRatio() - 1
}

// OmegaExcessReturn returns the annualized Omega excess return
//
//	Rp - 3·σDp·σDb
//
// where Rp is the annualized geometric mean portfolio return and σDp, σDb
// are the annualized portfolio and benchmark downside deviations about
// the MAR. The benchmark downside deviation is recomputed over the window
// (O(n)).
//
// NaN when there are no returns.
func (m *Measures) OmegaExcessReturn() float64 {
	benchmarkDownsideDeviation := func() float64 {
		if len(m.returnsBenchmark) == 0 {
			return math.NaN()
		}
		lowerExcessKBN := streamingkbn.NewRawMomentsKleinKBN(1, true, true)
		for _, r := range m.returnsBenchmark {
			excess := r - m.targetReturn
			if excess < 0 {
				lowerExcessKBN.Update(-excess)
			}
		}
		return math.Sqrt(lowerExcessKBN.X2Sum() / float64(len(m.returnsBenchmark)))
	}

	period := m.periodsPerAnnum

	// Annualized portfolio return.
	rp := m.cumulativeReturn.GeometricMeanReturn()
	if math.IsNaN(rp) {
		return math.NaN()
	}
	rpAnnual := math.Pow(1+rp, period) - 1

	sqrtPeriod := math.Sqrt(period)

	// Annualized portfolio downside (full) deviation.
	sigmaDAnn := m.DownsideDeviation() * sqrtPeriod

	// Annualized benchmark downside (full) deviation.
	sigmaDAnnBench := benchmarkDownsideDeviation() * sqrtPeriod

	return rpAnnual - 3*sigmaDAnn*sigmaDAnnBench
}

// Kappa1Ratio returns the Kappa ratio of order 1, mean(r - MAR)/LPM₁.
// NaN when LPM₁ is unavailable or zero.
func (m *Measures) Kappa1Ratio() float64 {
	lpm := m.targetPartialMoments.LowerPartialMoment1()
	if math.IsNaN(lpm) || lpm == 0 {
		return math.NaN()
	}
	return m.targetReturnsKBN.Mean() / lpm
}

// Kappa2Ratio returns the Kappa ratio of order 2, mean(r - MAR)/√LPM₂.
// NaN when LPM₂ is unavailable or zero.
func (m *Measures) Kappa2Ratio() float64 {
	lpm := m.targetPartialMoments.LowerPartialMoment2()
	if math.IsNaN(lpm) || lpm == 0 {
		return math.NaN()
	}
	return m.targetReturnsKBN.Mean() / math.Sqrt(lpm)
}

// Kappa3Ratio returns the Kappa ratio of order 3, mean(r - MAR)/LPM₃^(1/3).
// NaN when LPM₃ is unavailable or zero.
func (m *Measures) Kappa3Ratio() float64 {
	lpm := m.targetPartialMoments.LowerPartialMoment3()
	if math.IsNaN(lpm) || lpm == 0 {
		return math.NaN()
	}
	return m.targetReturnsKBN.Mean() / math.Pow(lpm, 1.0/3)
}

// Kappa4Ratio returns the Kappa ratio of order 4, mean(r - MAR)/LPM₄^(1/4).
// NaN when LPM₄ is unavailable or zero.
func (m *Measures) Kappa4Ratio() float64 {
	lpm := m.targetPartialMoments.LowerPartialMoment4()
	if math.IsNaN(lpm) || lpm == 0 {
		return math.NaN()
	}
	return m.targetReturnsKBN.Mean() / math.Pow(lpm, 1.0/4)
}

// ProspectRatio returns the prospect ratio based on Watanabe's
// formulation described by Bacon,
//
//	((Σr⁺ + λ·Σr⁻)/n - MAR) / σD
//
// where Σr⁻ is the signed sum of negative returns, λ the loss-aversion
// coefficient lambdaLoss (Python default 2.25) and σD the downside
// deviation. Not annualized.
//
// NaN when the downside deviation is unavailable or zero.
func (m *Measures) ProspectRatio(lambdaLoss float64) float64 {
	ddev := m.DownsideDeviation()
	if math.IsNaN(ddev) || ddev == 0 {
		return math.NaN()
	}

	pm := m.rawPartialMoments
	n := pm.Count()
	if n == 0 {
		return math.NaN()
	}

	prospectReturn := (pm.SumPositive() + lambdaLoss*pm.SumNegative()) / float64(n)

	return (prospectReturn - m.targetReturn) / ddev
}

// ProspectRatioPerformanceAnalytics returns the prospect ratio as
// implemented by ProspectRatio in R's PerformanceAnalytics,
//
//	(Σr⁺ + 2.25·Σr⁻ - MAR) / (n·σD)
//
// which differs from Bacon's formula when the MAR is non-zero. Not
// annualized.
//
// NaN when the downside deviation is unavailable or zero.
func (m *Measures) ProspectRatioPerformanceAnalytics() float64 {
	lambdaLoss := 2.25
	ddev := m.DownsideDeviation()
	if math.IsNaN(ddev) || ddev == 0 {
		return math.NaN()
	}

	pm := m.rawPartialMoments
	n := pm.Count()
	if n == 0 {
		return math.NaN()
	}

	return (pm.SumPositive() + lambdaLoss*pm.SumNegative() - m.targetReturn) / (ddev * float64(n))
}

// BernardoLedoitRatio returns the Bernardo-Ledoit ratio Σmax(r, 0) /
// Σmax(-r, 0). NaN when there are no negative returns.
func (m *Measures) BernardoLedoitRatio() float64 {
	lpm1 := m.rawPartialMoments.LowerPartialMoment1()
	hpm1 := m.rawPartialMoments.HigherPartialMoment1()
	return divOrNaN(hpm1, lpm1)
}

// DRatio returns the D-ratio
//
//	(n_d·Σmax(-r, 0)) / (n_u·Σmax(r, 0))
//
// where n_d and n_u are the numbers of negative and positive returns
// (zero returns are excluded). Lower is better.
//
// Returns +Inf when there are no positive returns (including no returns
// at all) and 0 when there are no negative returns.
func (m *Measures) DRatio() float64 {
	nUp := m.winLoss.WinningReturnsCount()
	if nUp == 0 {
		return math.Inf(1) // No positive returns.
	}
	nDown := m.winLoss.LosingReturnsCount()
	if nDown == 0 {
		return 0.0 // No negative returns.
	}

	sumUp := m.winLoss.WinningReturnsSum()
	sumDown := m.winLoss.LosingReturnsSum()
	return (float64(-nDown) * sumDown) / (float64(nUp) * sumUp)
}

// GainLossRatio returns the sum of positive returns divided by the
// absolute sum of negative returns. NaN when there are no negative
// returns.
func (m *Measures) GainLossRatio() float64 {
	sumLosses := math.Abs(m.winLoss.LosingReturnsSum())
	return divOrNaN(m.winLoss.WinningReturnsSum(), sumLosses)
}

// MeanNonZeroReturn returns the arithmetic mean of the non-zero returns.
// NaN when there are none.
func (m *Measures) MeanNonZeroReturn() float64 {
	return m.winLoss.NonZeroReturnsMean()
}

// MeanWinReturn returns the arithmetic mean of the strictly positive
// returns. NaN when there are none.
func (m *Measures) MeanWinReturn() float64 {
	return m.winLoss.WinningReturnsMean()
}

// MeanLossReturn returns the arithmetic mean of the strictly negative
// returns (a negative value). NaN when there are none.
func (m *Measures) MeanLossReturn() float64 {
	return m.winLoss.LosingReturnsMean()
}

// WinRate returns the proportion of non-zero returns that are positive,
// in [0, 1]. NaN when there are no non-zero returns.
func (m *Measures) WinRate() float64 {
	nonZeroCount := m.winLoss.NonZeroReturnsCount()
	if nonZeroCount <= 0 {
		return math.NaN()
	}
	return float64(m.winLoss.WinningReturnsCount()) / float64(nonZeroCount)
}

// LossRate returns the proportion of non-zero returns that are negative,
// in [0, 1]. NaN when there are no non-zero returns.
func (m *Measures) LossRate() float64 {
	nonZeroCount := m.winLoss.NonZeroReturnsCount()
	if nonZeroCount <= 0 {
		return math.NaN()
	}
	return float64(m.winLoss.LosingReturnsCount()) / float64(nonZeroCount)
}

// VariabilitySkewness returns the upside-to-downside variability skewness
// HPM₂/LPM₂ about the MAR. NaN when the moments are unavailable or LPM₂ is
// zero.
func (m *Measures) VariabilitySkewness() float64 {
	upMoment := m.targetPartialMoments.HigherPartialMoment2()
	downMoment := m.targetPartialMoments.LowerPartialMoment2()

	if math.IsNaN(upMoment) || math.IsNaN(downMoment) || downMoment == 0 {
		return math.NaN()
	}

	return upMoment / downMoment
}

// VolatilitySkewness returns the upside-to-downside volatility skewness
// √(HPM₂/LPM₂) about the MAR.
func (m *Measures) VolatilitySkewness() float64 {
	varSkew := m.VariabilitySkewness()
	if math.IsNaN(varSkew) {
		return math.NaN()
	}
	return math.Sqrt(varSkew)
}

// FarinelliTibilettiRatio returns the Farinelli-Tibiletti ratio
//
//	FT(u, l) = HPM_u(MAR)^(1/u) / LPM_l(MAR)^(1/l)
//
// with partial moments normalized by the total number of observations.
// FT(1, 1) is the Omega ratio, FT(1, 2) the upside potential ratio and
// FT(2, 2) the volatility skewness. The Python defaults are
// upperOrder = 2 and lowerOrder = 2.
//
// NaN when the moments are unavailable or the denominator is zero.
// Returns an error if either order is not 1, 2, 3 or 4.
func (m *Measures) FarinelliTibilettiRatio(upperOrder, lowerOrder int) (float64, error) {
	if upperOrder < 1 || upperOrder > 4 {
		return math.NaN(), errUpperOrder
	}
	if lowerOrder < 1 || lowerOrder > 4 {
		return math.NaN(), errLowerOrder
	}

	pm := m.targetPartialMoments
	var denom float64
	switch lowerOrder {
	case 1:
		denom = pm.LowerPartialMoment1()
	case 2:
		denom = math.Sqrt(pm.LowerPartialMoment2())
	case 3:
		denom = math.Pow(pm.LowerPartialMoment3(), 1.0/3)
	default:
		denom = math.Pow(pm.LowerPartialMoment4(), 1.0/4)
	}

	var num float64
	switch upperOrder {
	case 1:
		num = pm.HigherPartialMoment1()
	case 2:
		num = math.Sqrt(pm.HigherPartialMoment2())
	case 3:
		num = math.Pow(pm.HigherPartialMoment3(), 1.0/3)
	default:
		num = math.Pow(pm.HigherPartialMoment4(), 1.0/4)
	}

	if math.IsNaN(num) || math.IsNaN(denom) || denom == 0 {
		return math.NaN(), nil
	}
	return num / denom, nil
}

// RachevRatio returns the Rachev ratio ES_upper/ES_lower, the average
// return in the upper beta tail divided by the magnitude of the average
// return in the lower alpha tail, following the non-parametric definition
// of PerformanceAnalytics (the lower threshold is the alpha percentile,
// the upper threshold the observation at position floor((1 - beta)·n) of
// the sorted sample). Returns are not adjusted by the risk-free rate. The
// Python defaults are alpha = 0.1 and beta = 0.1. O(n log n).
//
// NaN with fewer than two observations, when a tail is empty or the lower
// tail expected loss is zero. Returns an error (checked after the
// observation count) if alpha or beta is not in (0, 1).
func (m *Measures) RachevRatio(alpha, beta float64) (float64, error) {
	n := m.returnsKBN.N()
	returns := m.returns

	if n < 2 {
		return math.NaN(), nil
	}

	if !(0 < alpha && alpha < 1) {
		return math.NaN(), errAlpha
	}

	if !(0 < beta && beta < 1) {
		return math.NaN(), errBeta
	}

	// Lower-tail VaR and expected shortfall.
	lowerVar, err := core.Percentile(returns, alpha)
	if err != nil {
		return math.NaN(), err
	}
	lowerTail := make([]float64, 0, len(returns))
	for _, r := range returns {
		if r <= lowerVar {
			lowerTail = append(lowerTail, r)
		}
	}

	if len(lowerTail) == 0 {
		return math.NaN(), nil
	}

	esLower := -pySum(lowerTail) / float64(len(lowerTail))

	// Upper-tail VaR and expected shortfall.
	sortedReturns := slices.Clone(returns)
	slices.Sort(sortedReturns)

	// PerformanceAnalytics uses
	//   n.upper <- floor((1-beta) * n)
	//   VaR.hat.upper <- sorted.returns[n.upper]
	// R is 1-based, so convert the position to a 0-based index.
	upperPosition := int(math.Floor((1.0 - beta) * float64(n)))

	if upperPosition < 1 {
		upperPosition = 1
	} else if upperPosition > n {
		upperPosition = n
	}

	upperVar := sortedReturns[upperPosition-1]
	upperTail := make([]float64, 0, len(returns))
	for _, r := range returns {
		if r >= upperVar {
			upperTail = append(upperTail, r)
		}
	}

	if len(upperTail) == 0 || esLower == 0 {
		return math.NaN(), nil
	}

	esUpper := pySum(upperTail) / float64(len(upperTail))

	return esUpper / esLower, nil
}
