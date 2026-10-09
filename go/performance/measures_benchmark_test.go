package performance

import (
	"fmt"
	"math"
	"testing"

	"zpano/performance/referencedata"
)

func dailyRf(rf float64) stream {
	return stream{daily: true, annualRiskFreeRate: annualize(rf, 252)}
}

func monthlyRf(rf float64) stream {
	return stream{monthly: true, annualRiskFreeRate: annualize(rf, 12)}
}

// assertYearlyAndDailyRf checks yearly data with the rate as is and daily
// data with the rate annualized over 252 periods.
func assertYearlyAndDailyRf(t *testing.T, refs map[float64][]float64, f func(*Measures) float64,
	tol tolerance, skip func(rf float64) int, name string,
) {
	t.Helper()
	for rf, expected := range refs {
		assertSeriesEqualSkip(t, runStream(t, yearlyRf(rf), f), expected, tol, skip(rf), fmt.Sprintf("%s (yearly, Rf %v)", name, rf))
		assertSeriesEqualSkip(t, runStream(t, dailyRf(rf), f), expected, tol, skip(rf), fmt.Sprintf("%s (daily, Rf %v)", name, rf))
	}
}

func noSkip(float64) int { return 0 }

func TestSfmRiskPremiumMatchesPerformanceAnalyticsOutput(t *testing.T) {
	t.Parallel()
	assertYearlyAndDailyRf(t, referencedata.SfmRiskPremiumExpectedValuesByRfPerfan, (*Measures).SfmRiskPremium, places(14), noSkip, "SFM risk premium")
}

func TestSfmAlphaMatchesPerformanceAnalyticsOutput(t *testing.T) {
	t.Parallel()
	assertYearlyAndDailyRf(t, referencedata.SfmAlphaExpectedValuesByRfPerfan, (*Measures).SfmAlpha, places(14), noSkip, "SFM alpha")
}

func TestSfmBetaMatchesPerformanceAnalyticsOutput(t *testing.T) {
	t.Parallel()
	assertYearlyAndDailyRf(t, referencedata.SfmBetaExpectedValuesByRfPerfan, (*Measures).SfmBeta, places(14), noSkip, "SFM beta")
}

func TestSfmBetaBullMatchesPerformanceAnalyticsOutput(t *testing.T) {
	t.Parallel()
	assertYearlyAndDailyRf(t, referencedata.SfmBetaBullExpectedValuesByRfPerfan, (*Measures).SfmBetaBull, places(14), noSkip, "SFM beta bull")
}

func TestSfmBetaBearMatchesReferenceImplementationOutput(t *testing.T) {
	t.Parallel()
	assertYearlyAndDailyRf(t, referencedata.SfmBetaBearExpectedValuesByRfReference, (*Measures).SfmBetaBear, places(14), noSkip, "SFM beta bear")
}

func TestTimingRatioMatchesPerformanceAnalyticsOutput(t *testing.T) {
	t.Parallel()
	assertYearlyAndDailyRf(t, referencedata.TimingRatioExpectedValuesByRfPerfan, (*Measures).TimingRatio, places(14), noSkip, "timing ratio")
}

func TestSfmR2MatchesPerformanceAnalyticsOutput(t *testing.T) {
	t.Parallel()
	assertYearlyAndDailyRf(t, referencedata.SfmR2ExpectedValuesByRfPerfan, (*Measures).SfmR2, places(14), func(rf float64) int {
		if rf < 0.05 {
			return 15
		}
		return 18
	}, "SFM R^2")
}

func TestJensenAlphaHighDailyRiskFreeRateDefinition(t *testing.T) {
	t.Parallel()
	// The R fixtures lose precision after compounding a 10% or 30% periodic
	// rate over 252 periods. Check the documented formula.
	for _, rf := range []float64{0.1, 0.3} {
		annualRf := annualize(rf, 252)
		m := newTestMeasures(t, 252, annualRf, 0, 0)
		for i := range baconPortfolioLen {
			m.AddReturn(baconPortfolioReturns[i], baconBenchmarkReturns[i])
			if i == 0 {
				continue
			}
			n := float64(i + 1)
			pAnn := math.Pow(prod(onePlus(baconPortfolioReturns[:i+1])), 252/n) - 1
			bAnn := math.Pow(prod(onePlus(baconBenchmarkReturns[:i+1])), 252/n) - 1
			beta := m.SfmBeta()
			expected := pAnn - (beta*bAnn + (1-beta)*annualRf)
			assertFloatEqual(t, m.JensenAlpha(), expected, relTol(1e-12, 1e-9),
				fmt.Sprintf("Jensen alpha rf=%v step=%d", rf, i))
			if beta != 0 {
				assertFloatEqual(t, m.JensenAlphaModified(), expected/beta, relTol(1e-12, 1e-9),
					fmt.Sprintf("Jensen alpha modified rf=%v step=%d", rf, i))
			}
		}
	}
}

func TestJensenAlphaMatchesPerformanceAnalyticsOutput(t *testing.T) {
	t.Parallel()
	for rf, expected := range referencedata.JensenAlphaExpectedValuesByRfDailyPerfan {
		actual := runStream(t, dailyRf(rf), (*Measures).JensenAlpha)
		if rf <= 0.05 {
			// At higher periodic rates the R reference loses precision
			// through cancellation; the formula test above covers them.
			assertSeriesEqual(t, actual, expected, relTol(1e-10, 9e-10), fmt.Sprintf("Jensen alpha (daily, Rf %v)", rf))
		}
	}
	for rf, expected := range referencedata.JensenAlphaExpectedValuesByRfMonthlyPerfan {
		actual := runStream(t, monthlyRf(rf), (*Measures).JensenAlpha)
		assertSeriesEqual(t, actual, expected, places(12), fmt.Sprintf("Jensen alpha (monthly, Rf %v)", rf))
	}
	for rf, expected := range referencedata.JensenAlphaExpectedValuesByRfYearlyPerfan {
		actual := runStream(t, yearlyRf(rf), (*Measures).JensenAlpha)
		assertSeriesEqual(t, actual, expected, places(14), fmt.Sprintf("Jensen alpha (yearly, Rf %v)", rf))
	}
}

func TestFamaBetaMatchesPerformanceAnalyticsOutput(t *testing.T) {
	t.Parallel()
	assertYearlyAndDaily(t, (*Measures).FamaBeta, referencedata.FamaBetaExpectedValuesPerfan, places(14), "fama beta")
}

func TestModiglianiMatchesPerformanceAnalyticsOutput(t *testing.T) {
	t.Parallel()
	assertYearlyAndDailyRf(t, referencedata.ModiglianiExpectedValuesByRfPerfan, (*Measures).Modigliani, places(15), noSkip, "Modigliani-Modigliani")
}

func TestTrackingErrorMatchesPerformanceAnalyticsOutput(t *testing.T) {
	t.Parallel()
	f := (*Measures).TrackingError
	assertSeriesEqual(t, runStream(t, stream{daily: true}, f), referencedata.TrackingErrorExpectedValuesDailyPerfan, places(15), "tracking error (daily)")
	assertSeriesEqual(t, runStream(t, stream{monthly: true}, f), referencedata.TrackingErrorExpectedValuesMonthlyPerfan, places(15), "tracking error (monthly)")
	assertSeriesEqual(t, runStream(t, stream{}, f), referencedata.TrackingErrorExpectedValuesAnnualPerfan, places(15), "tracking error (yearly)")
}

func TestActivePremiumMatchesPerformanceAnalyticsOutput(t *testing.T) {
	t.Parallel()
	// Skip the first element: PerformanceAnalytics uses it to determine
	// the periodicity.
	f := (*Measures).ActivePremium
	assertSeriesEqualSkip(t, runStream(t, stream{daily: true}, f), referencedata.ActivePremiumExpectedValuesDailyPerfan, places(11), 1, "active premium (daily)")
	assertSeriesEqualSkip(t, runStream(t, stream{monthly: true}, f), referencedata.ActivePremiumExpectedValuesMonthlyPerfan, places(14), 1, "active premium (monthly)")
	assertSeriesEqualSkip(t, runStream(t, stream{}, f), referencedata.ActivePremiumExpectedValuesAnnualPerfan, places(15), 1, "active premium (yearly)")
}

func TestInformationRatioMatchesPerformanceAnalyticsOutput(t *testing.T) {
	t.Parallel()
	f := (*Measures).InformationRatio
	assertSeriesEqualSkip(t, runStream(t, stream{daily: true}, f), referencedata.InformationRatioExpectedValuesDailyPerfan, places(10), 2, "information ratio (daily)")
	assertSeriesEqualSkip(t, runStream(t, stream{monthly: true}, f), referencedata.InformationRatioExpectedValuesMonthlyPerfan, places(12), 2, "information ratio (monthly)")
	assertSeriesEqualSkip(t, runStream(t, stream{}, f), referencedata.InformationRatioExpectedValuesAnnualPerfan, places(13), 2, "information ratio (yearly)")
}

func TestInformationRatioModifiedSignRule(t *testing.T) {
	t.Parallel()
	// Equals the information ratio when the geometric active premium is positive,
	// otherwise its negation.
	m := makeMeasures(t, 0, 0, 0, false, true)
	for i := range baconPortfolioLen {
		m.AddReturn(baconPortfolioReturns[i], baconBenchmarkReturns[i])
		active := m.ActivePremium()
		ir := m.InformationRatio()
		expected := ir
		if math.IsNaN(ir) {
			expected = math.NaN()
		} else if !(active > 0) {
			expected = -ir
		}
		assertFloatEqual(t, m.InformationRatioModified(), expected, places(15), fmt.Sprintf("step %d", i))
	}
}

// assertDailyMonthlyAnnual checks the daily, monthly and annual reference
// maps with the periodic rate annualized accordingly.
func assertDailyMonthlyAnnual(t *testing.T, daily, monthly, annual map[float64][]float64,
	f func(*Measures) float64, tolDaily, tolMonthly, tolAnnual func(rf float64) tolerance,
	skip int, include func(rf float64) bool, name string,
) {
	t.Helper()
	for rf, expected := range daily {
		actual := runStream(t, dailyRf(rf), f)
		if include(rf) {
			assertSeriesEqualSkip(t, actual, expected, tolDaily(rf), skip, fmt.Sprintf("%s (daily, Rf %v)", name, rf))
		}
	}
	for rf, expected := range monthly {
		assertSeriesEqualSkip(t, runStream(t, monthlyRf(rf), f), expected, tolMonthly(rf), skip, fmt.Sprintf("%s (monthly, Rf %v)", name, rf))
	}
	for rf, expected := range annual {
		assertSeriesEqualSkip(t, runStream(t, yearlyRf(rf), f), expected, tolAnnual(rf), skip, fmt.Sprintf("%s (yearly, Rf %v)", name, rf))
	}
}

func always(float64) bool { return true }

func TestSystematicRiskMatchesPerformanceAnalyticsOutput(t *testing.T) {
	t.Parallel()
	assertDailyMonthlyAnnual(t, referencedata.SystematicRiskExpectedValuesByRfDailyPerfan,
		referencedata.SystematicRiskExpectedValuesByRfMonthlyPerfan, referencedata.SystematicRiskExpectedValuesByRfAnnualPerfan,
		(*Measures).SystematicRisk, fixed(places(14)), fixed(places(15)), fixed(places(15)), 0, always, "systematic risk")
}

func TestTreynorRatioMatchesPerformanceAnalyticsOutput(t *testing.T) {
	t.Parallel()
	assertDailyMonthlyAnnual(t, referencedata.TreynorRatioExpectedValuesByRfDailyPerfan,
		referencedata.TreynorRatioExpectedValuesByRfMonthlyPerfan, referencedata.TreynorRatioExpectedValuesByRfAnnualPerfan,
		(*Measures).TreynorRatio, fixed(places(10)), fixed(places(13)), fixed(places(14)), 0, always, "treynor ratio")
}

func TestTreynorRatioModifiedMatchesPerformanceAnalyticsOutput(t *testing.T) {
	t.Parallel()
	assertDailyMonthlyAnnual(t, referencedata.TreynorRatioModifiedExpectedValuesByRfDailyPerfan,
		referencedata.TreynorRatioModifiedExpectedValuesByRfMonthlyPerfan, referencedata.TreynorRatioModifiedExpectedValuesByRfAnnualPerfan,
		(*Measures).TreynorRatioModified, fixed(places(10)), fixed(places(12)), fixed(places(12)), 0, always, "treynor ratio modified")
}

func TestSpecificRiskMatchesPerformanceAnalyticsOutput(t *testing.T) {
	t.Parallel()
	assertDailyMonthlyAnnual(t, referencedata.SpecificRiskExpectedValuesByRfDailyPerfan,
		referencedata.SpecificRiskExpectedValuesByRfMonthlyPerfan, referencedata.SpecificRiskExpectedValuesByRfAnnualPerfan,
		(*Measures).SpecificRisk, fixed(places(14)), fixed(places(15)), fixed(places(15)), 0, always, "specific risk")
}

func TestTotalRiskMatchesPerformanceAnalyticsOutput(t *testing.T) {
	t.Parallel()
	assertDailyMonthlyAnnual(t, referencedata.TotalRiskExpectedValuesByRfDailyPerfan,
		referencedata.TotalRiskExpectedValuesByRfMonthlyPerfan, referencedata.TotalRiskExpectedValuesByRfAnnualPerfan,
		(*Measures).TotalRisk, fixed(places(14)), fixed(places(14)), fixed(places(15)), 0, always, "total_risk")
}

func TestAppraisalRatioMatchesPerformanceAnalyticsOutput(t *testing.T) {
	t.Parallel()
	byRf := func(lo, hi float64) func(float64) tolerance {
		return func(rf float64) tolerance {
			if rf < 0.05 {
				return delta(lo)
			}
			return delta(hi)
		}
	}
	assertDailyMonthlyAnnual(t, referencedata.AppraisalRatioExpectedValuesByRfDailyPerfan,
		referencedata.AppraisalRatioExpectedValuesByRfMonthlyPerfan, referencedata.AppraisalRatioExpectedValuesByRfAnnualPerfan,
		(*Measures).AppraisalRatio, byRf(1e-8, 1e-4), byRf(1e-11, 1e-9), byRf(1e-13, 1e-11), 2,
		func(rf float64) bool { return rf < 0.1 }, "appraisal ratio")
}

func TestJensenAlphaModifiedDefinition(t *testing.T) {
	t.Parallel()
	m := makeMeasures(t, 0, 0, 0, false, false)
	for i := range baconPortfolioLen {
		m.AddReturn(baconPortfolioReturns[i], baconBenchmarkReturns[i])
		beta := m.SfmBeta()
		expected := math.NaN()
		if beta != 0 {
			expected = m.JensenAlpha() / beta
		}
		assertFloatEqual(t, m.JensenAlphaModified(), expected, places(14), fmt.Sprintf("Jensen alpha modified n=%d", i+1))
	}
}

func TestJensenAlphaModifiedMatchesPerformanceAnalyticsOutput(t *testing.T) {
	t.Parallel()
	// Annualizing 24 daily observations amplifies the reference's
	// floating-point error, even when the formula agrees.
	assertDailyMonthlyAnnual(t, referencedata.JensenAlphaModifiedExpectedValuesByRfDailyPerfan,
		referencedata.JensenAlphaModifiedExpectedValuesByRfMonthlyPerfan, referencedata.JensenAlphaModifiedExpectedValuesByRfAnnualPerfan,
		(*Measures).JensenAlphaModified, fixed(delta(1e-8)), fixed(delta(1e-10)), fixed(delta(1e-13)), 0,
		func(rf float64) bool { return rf < 0.05 }, "Jensen alpha modified")
}

func TestJensenAlphaAlternativeMatchesPerformanceAnalyticsOutput(t *testing.T) {
	t.Parallel()
	assertDailyMonthlyAnnual(t, referencedata.JensenAlphaAlternativeExpectedValuesByRfDailyPerfan,
		referencedata.JensenAlphaAlternativeExpectedValuesByRfMonthlyPerfan, referencedata.JensenAlphaAlternativeExpectedValuesByRfAnnualPerfan,
		(*Measures).JensenAlphaAlternative,
		func(rf float64) tolerance {
			if rf < 0.1 {
				return delta(0.1826)
			}
			return delta(0.707)
		},
		func(rf float64) tolerance {
			if rf < 0.3 {
				return places(10)
			}
			return places(9)
		},
		fixed(places(12)), 0, func(rf float64) bool { return rf < 0.3 }, "Jensen alpha alternative")
}

func TestMSquaredMatchesPerformanceAnalyticsOutput(t *testing.T) {
	t.Parallel()
	assertDailyMonthlyAnnual(t, referencedata.MSquaredExpectedValuesByRfDailyPerfan,
		referencedata.MSquaredExpectedValuesByRfMonthlyPerfan, referencedata.MSquaredExpectedValuesByRfAnnualPerfan,
		(*Measures).MSquared,
		func(rf float64) tolerance {
			if rf < 0.01 {
				return delta(0.1849)
			}
			return delta(0.82956)
		},
		func(rf float64) tolerance {
			if rf < 0.05 {
				return delta(0.00861)
			}
			return delta(1.621)
		},
		fixed(places(14)), 0, func(rf float64) bool { return rf < 0.05 }, "M squared")
}

func TestMSquaredExcessMatchesPerformanceAnalyticsOutput(t *testing.T) {
	t.Parallel()
	assertDailyMonthlyAnnual(t, referencedata.MSquaredExcessExpectedValuesByRfDailyPerfan,
		referencedata.MSquaredExcessExpectedValuesByRfMonthlyPerfan, referencedata.MSquaredExcessExpectedValuesByRfAnnualPerfan,
		(*Measures).MSquaredExcess,
		func(rf float64) tolerance {
			if rf < 0.01 {
				return delta(0.02244)
			}
			return delta(0.101)
		},
		func(rf float64) tolerance {
			if rf < 0.05 {
				return delta(0.007782)
			}
			return delta(1.466)
		},
		fixed(places(15)), 0, func(rf float64) bool { return rf < 0.05 }, "M squared excess")
}

func TestMSquaredSortinoMatchesPerformanceAnalyticsOutput(t *testing.T) {
	t.Parallel()
	f := (*Measures).MSquaredSortino
	for mar, expected := range referencedata.MSquaredSortinoExpectedValuesByMarDailyPerfan {
		actual := runStream(t, stream{daily: true, annualTargetReturn: annualize(mar, 252)}, f)
		assertSeriesEqualSkip(t, actual, expected, places(11), 3, fmt.Sprintf("M squared Sortino (daily, MAR %v)", mar))
	}
	for mar, expected := range referencedata.MSquaredSortinoExpectedValuesByMarMonthlyPerfan {
		actual := runStream(t, stream{monthly: true, annualTargetReturn: annualize(mar, 12)}, f)
		assertSeriesEqualSkip(t, actual, expected, places(14), 3, fmt.Sprintf("M squared Sortino (monthly, MAR %v)", mar))
	}
	for mar, expected := range referencedata.MSquaredSortinoExpectedValuesByMarAnnualPerfan {
		actual := runStream(t, yearlyMar(mar), f)
		assertSeriesEqualSkip(t, actual, expected, places(15), 3, fmt.Sprintf("M squared Sortino (yearly, MAR %v)", mar))
	}
}

func TestDocumentedFormulasMSquaredAndJensenAlphaAlternative(t *testing.T) {
	t.Parallel()
	for _, c := range []struct{ periods, periodicRf float64 }{{1, 0.05}, {12, 0.05}, {252, 0.01}} {
		annualRf := annualize(c.periodicRf, c.periods)
		m := newTestMeasures(t, c.periods, annualRf, 0, 0)
		for i := range baconPortfolioLen {
			m.AddReturn(baconPortfolioReturns[i], baconBenchmarkReturns[i])
			n := i + 1
			if n < 2 {
				continue
			}
			portfolio := baconPortfolioReturns[:n]
			benchmark := baconBenchmarkReturns[:n]
			pAnn := math.Pow(prod(onePlus(portfolio)), c.periods/float64(n)) - 1
			bAnn := math.Pow(prod(onePlus(benchmark)), c.periods/float64(n)) - 1
			scale := pstdev(benchmark) / pstdev(portfolio)
			expectedM2 := pAnn*scale + annualRf*(1-scale)
			prefix := fmt.Sprintf("periods=%v n=%d", c.periods, n)
			assertFloatEqual(t, m.MSquared(), expectedM2, relTol(1e-11, 1e-10), "M squared "+prefix)
			expectedExcess := (1+expectedM2)/(1+bAnn) - 1
			assertFloatEqual(t, m.MSquaredExcess(), expectedExcess, relTol(1e-11, 1e-10), "M squared excess "+prefix)
			systematicRisk := math.Abs(m.SfmBeta()) * stdev(benchmark) * math.Sqrt(c.periods)
			if systematicRisk > 0 && !math.IsInf(systematicRisk, 0) {
				assertFloatEqual(t, m.JensenAlphaAlternative(), m.JensenAlpha()/systematicRisk,
					relTol(1e-11, 1e-10), "Jensen alpha alternative "+prefix)
			}
		}
	}
}

func TestDocumentedFormulasMSquaredEqualVolatilityWithExtremeRate(t *testing.T) {
	t.Parallel()
	m := newTestMeasures(t, 252, math.Pow(1.3, 252)-1, 0, 0)
	for _, pb := range [][2]float64{{0.125, 0.25}, {0.375, 0.5}} {
		m.AddReturn(pb[0], pb[1])
	}
	// Binary-exact inputs give exactly equal portfolio and benchmark
	// volatility, so the annual risk-free terms must cancel.
	a, b := 1.125, 1.375
	expected := math.Pow(a*b, 126) - 1
	assertFloatEqual(t, m.MSquared(), expected, relTol(1e-14, 0), "M squared equal volatility")
}

func TestUpsideCaptureRatioMatchesPerformanceAnalyticsOutput(t *testing.T) {
	t.Parallel()
	for geom, expected := range referencedata.UpsideCaptureRatioExpectedValuesByGeometricPerfan {
		actual := runStream(t, stream{}, func(m *Measures) float64 { return m.UpsideCaptureRatio(geom) })
		assertSeriesEqualSkip(t, actual, expected, places(13), 1, fmt.Sprintf("Upside capture ratio (yearly, geometric %v)", geom))
	}
}

func TestDownsideCaptureRatioMatchesPerformanceAnalyticsOutput(t *testing.T) {
	t.Parallel()
	for geom, expected := range referencedata.DownsideCaptureRatioExpectedValuesByGeometricPerfan {
		actual := runStream(t, stream{}, func(m *Measures) float64 { return m.DownsideCaptureRatio(geom) })
		assertSeriesEqual(t, actual, expected, places(14), fmt.Sprintf("Downside capture ratio (yearly, geometric %v)", geom))
	}
}

func TestOverallCaptureRatioMatchesReferenceImplementationOutput(t *testing.T) {
	t.Parallel()
	for geom, expected := range referencedata.OverallCaptureRatioExpectedValuesByGeometricReference {
		actual := runStream(t, stream{}, func(m *Measures) float64 { return m.OverallCaptureRatio(geom) })
		assertSeriesEqual(t, actual, expected, places(13), fmt.Sprintf("Overall capture ratio (yearly, geometric %v)", geom))
	}
}

func TestUpNumberRatioMatchesPerformanceAnalyticsOutput(t *testing.T) {
	t.Parallel()
	actual := runStream(t, stream{}, (*Measures).UpNumberRatio)
	assertSeriesEqual(t, actual, referencedata.UpNumberRatioExpectedValuesPerfan, places(15), "Up number ratio (yearly)")
}

func TestDownNumberRatioMatchesPerformanceAnalyticsOutput(t *testing.T) {
	t.Parallel()
	actual := runStream(t, stream{}, (*Measures).DownNumberRatio)
	assertSeriesEqual(t, actual, referencedata.DownNumberRatioExpectedValuesPerfan, places(15), "Down number ratio (yearly)")
}

func TestUpPercentageRatioMatchesPerformanceAnalyticsOutput(t *testing.T) {
	t.Parallel()
	actual := runStream(t, stream{}, (*Measures).UpPercentageRatio)
	assertSeriesEqual(t, actual, referencedata.UpPercentageRatioExpectedValuesPerfan, places(15), "Up percentage ratio (yearly)")
}

func TestDownPercentageRatioMatchesPerformanceAnalyticsOutput(t *testing.T) {
	t.Parallel()
	actual := runStream(t, stream{}, (*Measures).DownPercentageRatio)
	assertSeriesEqual(t, actual, referencedata.DownPercentageRatioExpectedValuesPerfan, places(15), "Down percentage ratio (yearly)")
}
