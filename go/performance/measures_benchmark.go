package performance

import (
	"math"

	"zpano/streamingkbn"
)

// SfmRiskPremium returns the single-factor model (SFM) risk premium, the
// mean portfolio return in excess of the risk-free rate.
func (m *Measures) SfmRiskPremium() float64 {
	return m.excessReturnsKBN.Mean()
}

// SfmAlpha returns the single-factor model (SFM) alpha, the intercept of
// the least-squares regression
//
//	R_p - R_f = α + β·(R_b - R_f) + ε
//
// of the portfolio excess returns on the benchmark excess returns. The
// benchmark may be any return series. NaN when beta cannot be estimated.
func (m *Measures) SfmAlpha() float64 {
	return m.sfmRegression.Alpha()
}

// SfmBeta returns the single-factor model (SFM) beta, the least-squares
// slope Cov(R_p - R_f, R_b - R_f)/Var(R_b - R_f). When the benchmark is
// the market portfolio this is the CAPM beta. NaN when the benchmark
// excess returns have zero variance.
func (m *Measures) SfmBeta() float64 {
	return m.sfmRegression.Beta()
}

// SfmBetaBull returns the single-factor model bull beta, the SFM slope
// restricted to observations with a positive benchmark excess return.
// NaN when it cannot be estimated.
func (m *Measures) SfmBetaBull() float64 {
	return m.sfmRegression.BetaBull()
}

// SfmBetaBear returns the single-factor model bear beta, the SFM slope
// restricted to observations with a negative benchmark excess return.
// NaN when it cannot be estimated.
func (m *Measures) SfmBetaBear() float64 {
	return m.sfmRegression.BetaBear()
}

// TimingRatio returns the timing ratio β_bull/β_bear. A value above 1
// indicates greater benchmark sensitivity in rising than in falling
// markets. NaN when the bear beta is zero or either beta is undefined.
func (m *Measures) TimingRatio() float64 {
	return divOrNaN(m.sfmRegression.BetaBull(), m.sfmRegression.BetaBear())
}

// SfmR2 returns the coefficient of determination R² of the single-factor
// model, Corr(R_p - R_f, R_b - R_f)², in [0, 1]. NaN when it cannot be
// estimated.
func (m *Measures) SfmR2() float64 {
	return m.sfmRegression.R2()
}

// JensenAlpha returns the annualized Jensen's alpha
//
//	R_p - (β·R_b + (1 - β)·R_f)
//
// where R_p and R_b are the annualized geometric mean portfolio and
// benchmark returns, β the SFM beta and R_f the annual risk-free rate.
// This form retains R_b when β = 1, even for a huge risk-free rate.
func (m *Measures) JensenAlpha() float64 {
	rf := m.annualRiskFreeRate
	mean := m.cumulativeReturn.AnnualizedGeometricMeanReturn(m.periodsPerAnnum)
	meanB := m.benchmarkCumulativeReturn.AnnualizedGeometricMeanReturn(m.periodsPerAnnum)
	beta := m.SfmBeta()
	return mean - (beta*meanB + (1.0-beta)*rf)
}

// FamaBeta returns the Fama beta σ_P/σ_B, the ratio of the portfolio to
// the benchmark population standard deviation (total, not systematic,
// risk). Not annualized; the annualization factors would cancel. NaN when
// the benchmark standard deviation is zero.
func (m *Measures) FamaBeta() float64 {
	sigma := m.returnsKBN.StandardDeviationDdof0()
	sigmaB := m.benchmarkReturnsKBN.StandardDeviationDdof0()
	return divOrNaN(sigma, sigmaB)
}

// Modigliani returns the periodic Modigliani–Modigliani measure (M²)
//
//	R_f + mean(R_p - R_f)·σ_b/σ_p
//
// the return of the portfolio scaled to the benchmark's total risk, with
// population standard deviations σ of the excess portfolio returns and
// the benchmark returns. NaN when the portfolio standard deviation is
// zero.
func (m *Measures) Modigliani() float64 {
	sigma := m.excessReturnsKBN.StandardDeviationDdof0()
	if sigma == 0 {
		return math.NaN()
	}
	sigmaB := m.benchmarkReturnsKBN.StandardDeviationDdof0()

	// M² = Rf + (Rp - Rf)·(σ_b/σ).
	return m.riskFreeRate + m.excessReturnsKBN.Mean()*sigmaB/sigma
}

// TrackingError returns the annualized tracking error, the sample
// standard deviation (ddof=1) of the active returns r - b times
// √(periods per annum).
func (m *Measures) TrackingError() float64 {
	return m.activeReturnsKBN.StandardDeviationDdof1() * m.sqrtPeriodsPerAnnum
}

// ActivePremium returns the annualized active premium, the annualized
// geometric mean portfolio return less the annualized geometric mean
// benchmark return.
func (m *Measures) ActivePremium() float64 {
	mean := m.cumulativeReturn.AnnualizedGeometricMeanReturn(m.periodsPerAnnum)
	meanB := m.benchmarkCumulativeReturn.AnnualizedGeometricMeanReturn(m.periodsPerAnnum)
	return mean - meanB
}

// InformationRatio returns the annualized information ratio, the active
// premium divided by the tracking error. NaN when the tracking error is
// zero.
func (m *Measures) InformationRatio() float64 {
	return divOrNaN(m.ActivePremium(), m.TrackingError())
}

// InformationRatioModified returns the information ratio when the annualized
// geometric active premium is positive, and its negation otherwise.
func (m *Measures) InformationRatioModified() float64 {
	excess := m.ActivePremium()
	ir := m.InformationRatio()
	if math.IsNaN(excess) || math.IsNaN(ir) {
		return math.NaN()
	}
	if excess > 0 {
		return ir
	}
	return -ir
}

// SystematicRisk returns the annualized systematic risk
// |β|·σ(R_b - R_f)·√P, with the sample (ddof=1) standard deviation of the
// benchmark excess returns. NaN when beta or the standard deviation is
// undefined.
func (m *Measures) SystematicRisk() float64 {
	beta := m.SfmBeta()
	if math.IsNaN(beta) {
		return math.NaN()
	}
	benchmarkRisk := m.benchmarkExcessReturnsKBN.StandardDeviationDdof1()
	if math.IsNaN(benchmarkRisk) {
		return math.NaN()
	}
	return math.Abs(beta) * benchmarkRisk * m.sqrtPeriodsPerAnnum
}

// TreynorRatio returns the annualized Treynor ratio, the annualized
// geometric mean excess return over the risk-free rate divided by the SFM
// beta. NaN when beta is zero.
func (m *Measures) TreynorRatio() float64 {
	beta := m.SfmBeta()
	if beta == 0 {
		return math.NaN()
	}
	return m.cumulativeExcessReturn.AnnualizedGeometricMeanReturn(m.periodsPerAnnum) / beta
}

// TreynorRatioModified returns the modified Treynor ratio, the annualized
// geometric mean excess return divided by the systematic risk. NaN when
// the systematic risk is zero.
func (m *Measures) TreynorRatioModified() float64 {
	sr := m.SystematicRisk()
	if sr == 0 {
		return math.NaN()
	}
	return m.cumulativeExcessReturn.AnnualizedGeometricMeanReturn(m.periodsPerAnnum) / sr
}

// SpecificRisk returns the annualized specific (residual) risk, the
// population standard deviation of the SFM residuals
// (r - R_f) - α - β·(b - R_f) times √P. Recomputed over the window (O(n)).
// NaN when alpha or beta is undefined.
func (m *Measures) SpecificRisk() float64 {
	beta := m.SfmBeta()
	if math.IsNaN(beta) {
		return math.NaN()
	}
	alpha := m.SfmAlpha()
	if math.IsNaN(alpha) {
		return math.NaN()
	}
	epsilonKBN := streamingkbn.NewRawMomentsKleinKBN(0, true, true)
	rf := m.riskFreeRate
	rb := m.returnsBenchmark
	for i, r := range m.returns {
		epsilonKBN.Update(r - rf - alpha - beta*(rb[i]-rf))
	}
	return epsilonKBN.StandardDeviationDdof0() * m.sqrtPeriodsPerAnnum
}

// TotalRisk returns the annualized total risk √(systematic² + specific²).
// NaN when either component is undefined.
func (m *Measures) TotalRisk() float64 {
	syr := m.SystematicRisk()
	if math.IsNaN(syr) {
		return math.NaN()
	}
	spr := m.SpecificRisk()
	if math.IsNaN(spr) {
		return math.NaN()
	}
	return math.Sqrt(syr*syr + spr*spr)
}

// AppraisalRatio returns the appraisal ratio, the annualized Jensen's
// alpha divided by the specific risk. NaN when alpha is undefined or the
// specific risk is zero.
func (m *Measures) AppraisalRatio() float64 {
	alpha := m.JensenAlpha()
	if math.IsNaN(alpha) {
		return math.NaN()
	}
	return divOrNaN(alpha, m.SpecificRisk())
}

// JensenAlphaModified returns the modified Jensen's alpha, the annualized
// Jensen's alpha divided by the SFM beta. NaN when alpha is undefined or
// beta is zero.
func (m *Measures) JensenAlphaModified() float64 {
	alpha := m.JensenAlpha()
	if math.IsNaN(alpha) {
		return math.NaN()
	}
	return divOrNaN(alpha, m.SfmBeta())
}

// JensenAlphaAlternative returns the alternative Jensen's alpha, the
// annualized Jensen's alpha divided by the systematic risk. NaN when alpha
// is undefined or the systematic risk is zero.
func (m *Measures) JensenAlphaAlternative() float64 {
	alpha := m.JensenAlpha()
	if math.IsNaN(alpha) {
		return math.NaN()
	}
	return divOrNaN(alpha, m.SystematicRisk())
}

// MSquared returns the annualized M-squared (M²) risk-adjusted return
//
//	R_p·s + R_f·(1 - s),  s = σ_b/σ_p
//
// where R_p is the annualized geometric mean portfolio return, R_f the
// annual risk-free rate and σ the annualized population standard
// deviations of the portfolio and benchmark returns. This form keeps R_p
// when s = 1 instead of subtracting two large risk-free terms.
//
// NaN when there are no returns or the portfolio standard deviation is
// zero or undefined.
func (m *Measures) MSquared() float64 {
	pRet := m.cumulativeReturn.AnnualizedGeometricMeanReturn(m.periodsPerAnnum)
	if math.IsNaN(pRet) {
		return math.NaN()
	}
	pStd := m.returnsKBN.StandardDeviationDdof0() * m.sqrtPeriodsPerAnnum
	if math.IsNaN(pStd) || pStd == 0 {
		return math.NaN()
	}
	bStd := m.benchmarkReturnsKBN.StandardDeviationDdof0() * m.sqrtPeriodsPerAnnum
	if math.IsNaN(bStd) {
		return math.NaN()
	}

	scale := bStd / pStd
	// Keep pRet when scale == 1 instead of subtracting two large rf terms.
	return pRet*scale + m.annualRiskFreeRate*(1.0-scale)
}

// MSquaredExcess returns the geometric excess M-squared return
// (1 + M²)/(1 + R_b) - 1, where R_b is the annualized geometric mean
// benchmark return. NaN when M² or R_b is undefined.
func (m *Measures) MSquaredExcess() float64 {
	mSq := m.MSquared()
	if math.IsNaN(mSq) {
		return math.NaN()
	}
	bRet := m.benchmarkCumulativeReturn.AnnualizedGeometricMeanReturn(m.periodsPerAnnum)
	if math.IsNaN(bRet) {
		return math.NaN()
	}
	return (1.0+mSq)/(1.0+bRet) - 1.0
}

// MSquaredSortino returns the M-squared using downside risk (M² Sortino)
//
//	R_p + Sortino·√P·(σD_b - σD_p)
//
// where R_p is the annualized geometric mean portfolio return and σD_p,
// σD_b the periodic downside deviations of the portfolio and benchmark
// about the MAR. NaN when any component is undefined.
func (m *Measures) MSquaredSortino() float64 {
	sortino := m.SortinoRatio()
	if math.IsNaN(sortino) {
		return math.NaN()
	}
	pRet := m.cumulativeReturn.AnnualizedGeometricMeanReturn(m.periodsPerAnnum)
	if math.IsNaN(pRet) {
		return math.NaN()
	}
	pDD := m.DownsideDeviation()
	if math.IsNaN(pDD) {
		return math.NaN()
	}

	bCount := m.benchmarkTargetPartialMoments.TotalCount()
	if bCount == 0 {
		return math.NaN()
	}
	bDD := math.Sqrt(m.benchmarkTargetPartialMoments.LowerExcessMoment2Sum() / float64(bCount))
	if math.IsNaN(bDD) {
		return math.NaN()
	}

	return pRet + sortino*m.sqrtPeriodsPerAnnum*(bDD-pDD)
}

// UpsideCaptureRatio returns the upside capture ratio, the portfolio's
// performance relative to the benchmark over the periods with a positive
// benchmark return. If geometric is true (the Python default), the
// cumulative compounded returns of these periods are compared, otherwise
// the sums of the periodic returns. Higher is better.
//
// NaN when the benchmark has no positive periods or its aggregate upside
// return is zero.
func (m *Measures) UpsideCaptureRatio(geometric bool) float64 {
	if geometric {
		return m.capture.UpsideCaptureRatioGeometric()
	}
	return m.capture.UpsideCaptureRatioArithmetic()
}

// DownsideCaptureRatio returns the downside capture ratio, the portfolio's
// performance relative to the benchmark over the periods with a
// non-positive benchmark return. If geometric is true (the Python
// default), the cumulative compounded returns of these periods are
// compared, otherwise the sums of the periodic returns. Lower is better.
//
// NaN when the benchmark has no such periods or its aggregate downside
// return is zero.
func (m *Measures) DownsideCaptureRatio(geometric bool) float64 {
	if geometric {
		return m.capture.DownsideCaptureRatioGeometric()
	}
	return m.capture.DownsideCaptureRatioArithmetic()
}

// OverallCaptureRatio returns the upside capture ratio divided by the
// downside capture ratio (geometric if geometric is true, the Python
// default). NaN when either is undefined or the downside capture ratio
// is zero.
func (m *Measures) OverallCaptureRatio(geometric bool) float64 {
	up := m.UpsideCaptureRatio(geometric)
	down := m.DownsideCaptureRatio(geometric)
	if math.IsNaN(up) || math.IsNaN(down) || down == 0 {
		return math.NaN()
	}
	return up / down
}

// UpNumberRatio returns the proportion of periods with a positive
// benchmark return in which the portfolio return was also positive.
func (m *Measures) UpNumberRatio() float64 {
	return m.capture.UpNumberRatio()
}

// DownNumberRatio returns the proportion of periods with a non-positive
// benchmark return (b <= 0, the downside capture bucket) in which the
// portfolio return was negative. Lower is better.
func (m *Measures) DownNumberRatio() float64 {
	return m.capture.DownNumberRatio()
}

// UpPercentageRatio returns the proportion of periods with a positive
// benchmark return in which the portfolio outperformed the benchmark.
func (m *Measures) UpPercentageRatio() float64 {
	return m.capture.UpPercentageRatio()
}

// DownPercentageRatio returns the proportion of periods with a negative
// benchmark return in which the portfolio outperformed the benchmark.
func (m *Measures) DownPercentageRatio() float64 {
	return m.capture.DownPercentageRatio()
}
