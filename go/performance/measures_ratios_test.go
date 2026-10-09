package performance

import (
	"fmt"
	"math"
	"testing"

	"zpano/performance/referencedata"
)

// assertByKey streams f for every key of the reference map, configured by
// cfg, and compares with the expected series.
func assertByKey(t *testing.T, refs map[float64][]float64, cfg func(k float64) stream,
	f func(*Measures) float64, tol func(k float64) tolerance, skip int, prefix string,
) {
	t.Helper()
	for k, expected := range refs {
		actual := runStream(t, cfg(k), f)
		assertSeriesEqualSkip(t, actual, expected, tol(k), skip, fmt.Sprintf("%s %v", prefix, k))
	}
}

func fixed(tol tolerance) func(float64) tolerance { return func(float64) tolerance { return tol } }

func yearlyMar(mar float64) stream { return stream{annualTargetReturn: mar} }

func yearlyRf(rf float64) stream { return stream{annualRiskFreeRate: rf} }

func TestUpsidePotentialRatioMatchesPerformanceAnalyticsOutput(t *testing.T) {
	t.Parallel()
	assertByKey(t, referencedata.UpsidePotentialRatioExpectedValuesByMarFull, yearlyMar,
		(*Measures).UpsidePotentialRatio, fixed(places(14)), 0, "upside potential ratio (full) MAR")
}

func TestUpsidePotentialRatioSubsetMatchesPerformanceAnalyticsOutput(t *testing.T) {
	t.Parallel()
	assertByKey(t, referencedata.UpsidePotentialRatioExpectedValuesByMarSubset, yearlyMar,
		(*Measures).UpsidePotentialRatioSubset, fixed(places(14)), 0, "upside potential ratio (subset) MAR")
}

func TestUpsideFrequencyMatchesPerformanceAnalyticsOutput(t *testing.T) {
	t.Parallel()
	assertByKey(t, referencedata.UpsideFrequencyExpectedValuesByMar, yearlyMar,
		(*Measures).UpsideFrequency, fixed(places(15)), 0, "upside frequency MAR")
}

func TestUpsidePotentialMatchesPerformanceAnalyticsOutput(t *testing.T) {
	t.Parallel()
	refs := referencedata.UpsideRiskExpectedValuesByMarPotentialFull
	assertByKey(t, refs, yearlyMar, (*Measures).UpsidePotential, fixed(places(15)), 0,
		"upside potential (full) MAR")
	assertByKey(t, refs, func(mar float64) stream {
		return stream{daily: true, annualTargetReturn: annualize(mar, 252)}
	}, (*Measures).UpsidePotential, fixed(places(15)), 0, "upside potential (full) daily MAR")
}

func TestUpsidePotentialSubsetMatchesPerformanceAnalyticsOutput(t *testing.T) {
	t.Parallel()
	assertByKey(t, referencedata.UpsideRiskExpectedValuesByMarPotentialSubset, yearlyMar,
		(*Measures).UpsidePotentialSubset, fixed(places(15)), 0, "upside potential (subset) MAR")
}

func TestUpsideVarianceMatchesPerformanceAnalyticsOutput(t *testing.T) {
	t.Parallel()
	assertByKey(t, referencedata.UpsideRiskExpectedValuesByMarVarianceFull, yearlyMar,
		(*Measures).UpsideVariance, fixed(places(15)), 0, "upside variance (full) MAR")
}

func TestUpsideVarianceSubsetMatchesPerformanceAnalyticsOutput(t *testing.T) {
	t.Parallel()
	assertByKey(t, referencedata.UpsideRiskExpectedValuesByMarVarianceSubset, yearlyMar,
		(*Measures).UpsideVarianceSubset, fixed(places(15)), 0, "upside variance (subset) MAR")
}

func TestUpsideRiskMatchesPerformanceAnalyticsOutput(t *testing.T) {
	t.Parallel()
	assertByKey(t, referencedata.UpsideRiskExpectedValuesByMarRiskFull, yearlyMar,
		(*Measures).UpsideRisk, fixed(places(15)), 0, "upside risk (full) MAR")
}

func TestUpsideRiskSubsetMatchesPerformanceAnalyticsOutput(t *testing.T) {
	t.Parallel()
	refs := referencedata.UpsideRiskExpectedValuesByMarRiskSubset
	assertByKey(t, refs, yearlyMar, (*Measures).UpsideRiskSubset, fixed(places(15)), 0,
		"upside risk (subset) yearly MAR")
	assertByKey(t, refs, func(mar float64) stream {
		return stream{daily: true, annualTargetReturn: annualize(mar, 252)}
	}, (*Measures).UpsideRiskSubset, fixed(places(15)), 0, "upside risk (subset) daily MAR")
}

func TestSemiDeviationMatchesPerformanceAnalyticsOutput(t *testing.T) {
	t.Parallel()
	actual := runStream(t, stream{}, (*Measures).SemiDeviation)
	assertSeriesEqual(t, actual, referencedata.SemiDeviationExpectedValues, places(15), "semi-deviation")
}

func TestDownsideDeviationMatchesPerformanceAnalyticsOutput(t *testing.T) {
	t.Parallel()
	assertByKey(t, referencedata.DownsideDeviationExpectedValuesByMarFull, yearlyMar,
		(*Measures).DownsideDeviation, fixed(places(15)), 0, "downside deviation MAR")
}

func TestDownsideDeviationSubsetMatchesPerformanceAnalyticsOutput(t *testing.T) {
	t.Parallel()
	assertByKey(t, referencedata.DownsideDeviationExpectedValuesByMarSubset, yearlyMar,
		(*Measures).DownsideDeviationSubset, fixed(places(15)), 0, "downside deviation subset MAR")
}

func TestDownsideFrequencyMatchesPerformanceAnalyticsOutput(t *testing.T) {
	t.Parallel()
	assertByKey(t, referencedata.DownsideFrequencyExpectedValuesByMar, yearlyMar,
		(*Measures).DownsideFrequency, fixed(places(15)), 0, "downside frequency MAR")
}

func TestDownsidePotentialMatchesPerformanceAnalyticsOutput(t *testing.T) {
	t.Parallel()
	assertByKey(t, referencedata.DownsidePotentialExpectedValuesByMar, yearlyMar,
		(*Measures).DownsidePotential, fixed(places(15)), 0, "downside potential MAR")
}

func TestSharpeRatioMatchesPerformanceAnalyticsOutput(t *testing.T) {
	t.Parallel()
	assertByKey(t, referencedata.SharpeRatioExpectedValuesByRfStdev, yearlyRf,
		(*Measures).SharpeRatio, func(rf float64) tolerance {
			if rf < 0.25 {
				return places(13)
			}
			return places(12)
		}, 0, "Sharpe ratio (stdev) Rf")
}

// assertByConfidenceRf streams f(m, p) for the nested p -> rf -> expected
// reference data with the yearly default and annual risk-free rate rf.
func assertByConfidenceRf(t *testing.T, refs map[float64]map[float64][]float64,
	f func(*Measures, float64) float64, tol tolerance, prefix string,
) {
	t.Helper()
	for p, rfPack := range refs {
		for rf, expected := range rfPack {
			actual := runStream(t, yearlyRf(rf), func(m *Measures) float64 { return f(m, p) })
			assertSeriesEqual(t, actual, expected, tol, fmt.Sprintf("%s conf %v Rf %v", prefix, p, rf))
		}
	}
}

func TestSharpeRatioVarHistoricalMatchesPerformanceAnalyticsOutput(t *testing.T) {
	t.Parallel()
	assertByConfidenceRf(t, referencedata.SharpeRatioExpectedValuesByPRfVarHistorical,
		(*Measures).SharpeRatioVarHistorical, places(12), "Sharpe ratio (VaR historical)")
}

func TestSharpeRatioVarGaussianMatchesPerformanceAnalyticsOutput(t *testing.T) {
	t.Parallel()
	assertByConfidenceRf(t, referencedata.SharpeRatioExpectedValuesByPRfVarGaussian,
		(*Measures).SharpeRatioVarGaussian, places(5), "Sharpe ratio (VaR Gaussian)")
}

func TestSharpeRatioVarCornishFisherMatchesPerformanceAnalyticsOutput(t *testing.T) {
	t.Parallel()
	assertByConfidenceRf(t, referencedata.SharpeRatioExpectedValuesByPRfVarCornishFisher,
		(*Measures).SharpeRatioVarCornishFisher, places(6), "Sharpe ratio (VaR Cornish-Fisher)")
}

func TestSharpeRatioEsHistoricalMatchesPerformanceAnalyticsOutput(t *testing.T) {
	t.Parallel()
	for p, rfPack := range referencedata.SharpeRatioExpectedValuesByPRfEsHistorical {
		for rf, expected := range rfPack {
			actual := runStream(t, yearlyRf(rf), func(m *Measures) float64 { return m.SharpeRatioEsHistorical(p) })
			if len(actual) != len(expected) {
				t.Fatalf("length %d, expected %d", len(actual), len(expected))
			}
			for i := range actual {
				a, e := actual[i], expected[i]
				prefix := fmt.Sprintf("Sharpe ratio (ES historical) conf %v Rf %v step %d", p, rf, i)
				if p == 0.9 && rf == 0.001 && i == 10 {
					// The quantile is exactly the second-worst return. This
					// implementation includes both tied-to-tail
					// observations; the R reference includes only one.
					excess := make([]float64, 11)
					for j, r := range baconPortfolioReturns[:11] {
						excess[j] = r - rf
					}
					excessMean := fsum(excess) / 11
					assertAlmostEqual(t, a, excessMean/0.013, 13, prefix)
				} else {
					assertFloatEqual(t, a, e, delta(1e-12), prefix)
				}
			}
		}
	}
}

func TestSharpeRatioEsGaussianMatchesPerformanceAnalyticsOutput(t *testing.T) {
	t.Parallel()
	assertByConfidenceRf(t, referencedata.SharpeRatioExpectedValuesByPRfEsGaussian,
		(*Measures).SharpeRatioEsGaussian, places(7), "Sharpe ratio (ES Gaussian)")
}

func TestSharpeRatioEsCornishFisherMatchesPerformanceAnalyticsOutput(t *testing.T) {
	t.Parallel()
	assertByConfidenceRf(t, referencedata.SharpeRatioExpectedValuesByPRfEsCornishFisher,
		(*Measures).SharpeRatioEsCornishFisher, places(5), "Sharpe ratio (ES Cornish-Fisher)")
}

func TestDownsideSharpeRatioMatchesPerformanceAnalyticsOutput(t *testing.T) {
	t.Parallel()
	assertByKey(t, referencedata.DownsideSharpeRatioExpectedValuesByRf, yearlyRf,
		(*Measures).DownsideSharpeRatio, fixed(places(13)), 0, "downside Sharpe ratio Rf")
}

func TestAdjustedSharpeRatioMatchesPerformanceAnalyticsOutput(t *testing.T) {
	t.Parallel()
	assertByKey(t, referencedata.AdjustedSharpeRatioExpectedValuesByRf, yearlyRf,
		(*Measures).AdjustedSharpeRatio, func(rf float64) tolerance {
			if rf < 0.2 {
				return places(12)
			}
			return places(11)
		}, 0, "adjusted Sharpe ratio (stdev) Rf")
}

func TestAdjustedSharpeRatioSkewOnlyScaleAndTranslationInvariance(t *testing.T) {
	t.Parallel()
	// Since we don't have data for it, we test only its metamorphic
	// properties.
	rf := 0.0042
	expected := runStream(t, yearlyRf(rf), (*Measures).AdjustedSharpeRatioSkewOnly)

	for _, scale := range []float64{4.2, -4.2} {
		for _, shift := range []float64{0.042, -0.042} {
			transformed := make([]float64, baconPortfolioLen)
			for i, r := range baconPortfolioReturns {
				transformed[i] = scale*r + shift
			}
			m := newTestMeasures(t, 1, scale*rf+shift, 0, 0)
			m.Reset()
			for i := range baconPortfolioLen {
				m.AddReturn(transformed[i], transformed[i])
				a := m.AdjustedSharpeRatioSkewOnly()
				if scale <= 0 {
					a = -a
				}
				assertFloatEqual(t, a, expected[i], places(14),
					fmt.Sprintf("ASR skew-only (scale %v shift %v) step %d", scale, shift, i))
			}
		}
	}
}

// assertByReferenceSRRf streams f(m, referenceSR) for the nested
// referenceSR -> rf -> expected reference data.
func assertByReferenceSRRf(t *testing.T, refs map[float64]map[float64][]float64,
	f func(*Measures, float64) float64, prefix string,
) {
	t.Helper()
	for refSR, rfPack := range refs {
		for rf, expected := range rfPack {
			actual := runStream(t, yearlyRf(rf), func(m *Measures) float64 { return f(m, refSR) })
			assertSeriesEqual(t, actual, expected, places(14),
				fmt.Sprintf("%s reference_sr %v Rf %v", prefix, refSR, rf))
		}
	}
}

func TestProbabilisticSharpeRatioMatchesPerformanceAnalyticsOutput(t *testing.T) {
	t.Parallel()
	assertByReferenceSRRf(t, referencedata.ProbabilisticSharpeRatioExpectedValuesByRefsrRf,
		(*Measures).ProbabilisticSharpeRatio, "probabilistic Sharpe ratio")
}

func TestProbabilisticSharpeRatioFullMatchesPerformanceAnalyticsOutput(t *testing.T) {
	t.Parallel()
	assertByReferenceSRRf(t, referencedata.ProbabilisticSharpeRatioExpectedValuesByRefsrRfFull,
		(*Measures).ProbabilisticSharpeRatioFull, "probabilistic Sharpe ratio (full)")
}

func TestProbabilisticSharpeRatioSymmetricMatchesPerformanceAnalyticsOutput(t *testing.T) {
	t.Parallel()
	assertByReferenceSRRf(t, referencedata.ProbabilisticSharpeRatioExpectedValuesByRefsrRfSymmetric,
		(*Measures).ProbabilisticSharpeRatioSymmetric, "probabilistic Sharpe ratio (symmetric)")
}

func TestProbabilisticSharpeRatioGaussianMatchesPerformanceAnalyticsOutput(t *testing.T) {
	t.Parallel()
	assertByReferenceSRRf(t, referencedata.ProbabilisticSharpeRatioExpectedValuesByRefsrRfGaussian,
		(*Measures).ProbabilisticSharpeRatioGaussian, "probabilistic Sharpe ratio (Gaussian)")
}

func TestSortinoRatioMatchesPerformanceAnalyticsOutput(t *testing.T) {
	t.Parallel()
	assertByKey(t, referencedata.SortinoRatioExpectedValuesByMar, yearlyMar,
		(*Measures).SortinoRatio, fixed(places(14)), 0, "sortino ratio MAR")
}

func TestSortinoRatioJackSchagerSqrt2Version(t *testing.T) {
	t.Parallel()
	for mar, expected := range referencedata.SortinoRatioExpectedValuesByMar {
		scaled := make([]float64, len(expected))
		for i, x := range expected {
			scaled[i] = x / math.Sqrt(2)
		}
		actual := runStream(t, yearlyMar(mar), (*Measures).SortinoRatioSqrt2)
		assertSeriesEqual(t, actual, scaled, places(14), fmt.Sprintf("sortino ratio (sqrt2) MAR %v", mar))
	}
}

func TestSortinoSatchellRatioShouldBeComputable(t *testing.T) {
	t.Parallel()
	actual := runStream(t, stream{}, (*Measures).SortinoSatchellRatio)
	assertFloatEqual(t, actual[len(actual)-1], 0.3923720287950653, places(15), "Sortino-Satchell ratio")
}

func TestOmegaRatioMatchesPerformanceAnalyticsOutput(t *testing.T) {
	t.Parallel()
	assertByKey(t, referencedata.OmegaRatioExpectedValuesByMar, yearlyMar,
		(*Measures).OmegaRatio, fixed(places(13)), 0, "omega ratio MAR")
}

func TestOmegaSharpeRatioMatchesPerformanceAnalyticsOutput(t *testing.T) {
	t.Parallel()
	assertByKey(t, referencedata.OmegaSharpeRatioExpectedValuesByMar, yearlyMar,
		(*Measures).OmegaSharpeRatio, fixed(places(13)), 0, "omega Sharpe ratio MAR")
}

func TestOmegaExcessReturnMatchesPerformanceAnalyticsOutput(t *testing.T) {
	t.Parallel()
	assertByKey(t, referencedata.OmegaExcessReturnExpectedValuesByMarWithBenchmark, func(mar float64) stream {
		return stream{daily: true, annualTargetReturn: annualize(mar, 252)}
	}, (*Measures).OmegaExcessReturn, fixed(places(11)), 0, "omega excess return (with benchmark) MAR")
	assertByKey(t, referencedata.OmegaExcessReturnExpectedValuesByMarWithSelf, func(mar float64) stream {
		return stream{daily: true, annualTargetReturn: annualize(mar, 252), benchmark: baconPortfolioReturns}
	}, (*Measures).OmegaExcessReturn, fixed(places(11)), 0, "omega excess return (with self) MAR")
}

func TestKappaRatioMatchesPerformanceAnalyticsOutput(t *testing.T) {
	t.Parallel()
	// Generated data is yearly.
	assertByKey(t, referencedata.KappaRatioExpectedValuesByMarOrder1, yearlyMar,
		(*Measures).Kappa1Ratio, fixed(places(13)), 0, "kappa 1 ratio MAR")
	assertByKey(t, referencedata.KappaRatioExpectedValuesByMarOrder2, yearlyMar,
		(*Measures).Kappa2Ratio, fixed(places(14)), 0, "kappa 2 ratio MAR")
	assertByKey(t, referencedata.KappaRatioExpectedValuesByMarOrder3, yearlyMar,
		(*Measures).Kappa3Ratio, fixed(places(14)), 0, "kappa 3 ratio MAR")
	assertByKey(t, referencedata.KappaRatioExpectedValuesByMarOrder4, yearlyMar,
		(*Measures).Kappa4Ratio, fixed(places(14)), 0, "kappa 4 ratio MAR")
}

func TestProspectRatioMatchesPerformanceAnalyticsOutput(t *testing.T) {
	t.Parallel()
	assertByKey(t, referencedata.ProspectRatioExpectedValuesByMarPerfan, yearlyMar,
		(*Measures).ProspectRatioPerformanceAnalytics, fixed(places(13)), 0,
		"Prospect ratio PerformanceAnalytics version (yearly) MAR")
}

func TestProspectRatioMatchesReferenceImplementationOutput(t *testing.T) {
	t.Parallel()
	assertByKey(t, referencedata.ProspectRatioExpectedValuesByMarReference, yearlyMar,
		func(m *Measures) float64 { return m.ProspectRatio(2.25) }, fixed(places(15)), 0,
		"Prospect ratio (yearly) MAR")
}

func TestBernardoLedoitRatioMatchesPerformanceAnalyticsOutput(t *testing.T) {
	t.Parallel()
	actual := runStream(t, stream{}, (*Measures).BernardoLedoitRatio)
	assertSeriesEqual(t, actual, referencedata.BernardoLedoitRatioExpectedValues, places(13), "Bernardo-Ledoit ratio")
}

func TestDRatioMatchesPerformanceAnalyticsOutput(t *testing.T) {
	t.Parallel()
	actual := runStream(t, stream{}, (*Measures).DRatio)
	assertSeriesEqual(t, actual, referencedata.DRatioExpectedValuesPerfan, places(15), "d-ratio")
}

func TestGainLossRatioMatchesPerformanceAnalyticsOutput(t *testing.T) {
	t.Parallel()
	actual := runStream(t, stream{}, (*Measures).GainLossRatio)
	assertSeriesEqual(t, actual, referencedata.BernardoLedoitRatioExpectedValues, places(13), "gain-loss ratio")
}

func TestMeanNonZeroReturnCalculatedByHand(t *testing.T) {
	t.Parallel()
	actual := runStream(t, stream{}, (*Measures).MeanNonZeroReturn)
	assertSeriesEqual(t, actual, referencedata.MeanNonZeroReturnExpectedValues, places(15), "mean non-zero return")
}

func TestMeanWinReturnCalculatedByHand(t *testing.T) {
	t.Parallel()
	actual := runStream(t, stream{}, (*Measures).MeanWinReturn)
	assertSeriesEqual(t, actual, referencedata.MeanWinReturnExpectedValues, places(15), "mean win return")
}

func TestMeanLossReturnCalculatedByHand(t *testing.T) {
	t.Parallel()
	actual := runStream(t, stream{}, (*Measures).MeanLossReturn)
	assertSeriesEqual(t, actual, referencedata.MeanLossReturnExpectedValues, places(15), "mean loss return")
}

func TestWinRateCalculatedByHand(t *testing.T) {
	t.Parallel()
	actual := runStream(t, stream{}, (*Measures).WinRate)
	assertSeriesEqual(t, actual, referencedata.WinRateExpectedValues, places(15), "win rate")
}

func TestLossRateCalculatedByHand(t *testing.T) {
	t.Parallel()
	actual := runStream(t, stream{}, (*Measures).LossRate)
	assertSeriesEqual(t, actual, referencedata.LossRateExpectedValues, places(15), "loss rate")
}

func TestVolatilitySkewnessMatchesPerformanceAnalyticsOutput(t *testing.T) {
	t.Parallel()
	for mar, expected := range referencedata.VolatilitySkewnessExpectedValuesByMarVolatility {
		actual := runStream(t, yearlyMar(mar), (*Measures).VolatilitySkewness)
		assertSeriesEqual(t, actual, expected, places(13), fmt.Sprintf("volatility skewness MAR %v", mar))
	}
	for mar, expected := range referencedata.VolatilitySkewnessExpectedValuesByMarVariability {
		actual := runStream(t, yearlyMar(mar), (*Measures).VariabilitySkewness)
		assertSeriesEqual(t, actual, expected, places(13), fmt.Sprintf("variability skewness MAR %v", mar))
	}
}

func TestFarinelliTibilettiRatioShouldBeComputable(t *testing.T) {
	t.Parallel()
	// Calculation doesn't depend on periods per annum, so we use the
	// yearly default.
	mar := 0.005

	verify := func(upper, lower int, related string, relatedF func(*Measures) float64,
		transform func(float64) float64, n int,
	) {
		t.Helper()
		expected := runStream(t, yearlyMar(mar), relatedF)
		actual := runStream(t, yearlyMar(mar), func(m *Measures) float64 {
			return transform(must(m.FarinelliTibilettiRatio(upper, lower)))
		})
		assertSeriesEqual(t, actual, expected, places(n),
			fmt.Sprintf("Farinelli-Tibiletti ratio (u %d, l %d) vs %s", upper, lower, related))
	}

	identity := func(r float64) float64 { return r }
	verify(1, 1, "omega_ratio", (*Measures).OmegaRatio, identity, 14)
	verify(1, 1, "kappa_1_ratio", (*Measures).Kappa1Ratio, func(r float64) float64 { return r - 1 }, 14)
	verify(1, 2, "upside_potential_ratio", (*Measures).UpsidePotentialRatio, identity, 15)
	verify(2, 2, "volatility_skewness", (*Measures).VolatilitySkewness, identity, 14)
	verify(2, 2, "variability_skewness", (*Measures).VariabilitySkewness, func(r float64) float64 { return r * r }, 13)

	verifyManual := func(upper, lower int, n int) {
		t.Helper()
		upms := make([]float64, baconPortfolioLen)
		lpms := make([]float64, baconPortfolioLen)
		for i, r := range baconPortfolioReturns {
			upms[i] = math.Pow(math.Max(r-mar, 0.0), float64(upper))
			lpms[i] = math.Pow(math.Max(mar-r, 0.0), float64(lower))
		}
		upm := pySum(upms) / float64(baconPortfolioLen)
		lpm := pySum(lpms) / float64(baconPortfolioLen)
		expected := math.Pow(upm, 1.0/float64(upper)) / math.Pow(lpm, 1.0/float64(lower))
		actual := runStream(t, yearlyMar(mar), func(m *Measures) float64 {
			return must(m.FarinelliTibilettiRatio(upper, lower))
		})
		assertFloatEqual(t, actual[len(actual)-1], expected, places(n),
			fmt.Sprintf("Farinelli-Tibiletti ratio (u %d, l %d) vs manual calculation", upper, lower))
	}

	for _, i := range []int{1, 2, 3, 4} {
		for _, j := range []int{1, 2, 3, 4} {
			verifyManual(i, j, 15)
		}
	}
}

func TestFarinelliTibilettiRatioRejectsInvalidOrders(t *testing.T) {
	t.Parallel()
	m := makeMeasures(t, 0, 0, 0, false, false)
	addBacon(m, nil, nil)
	for _, c := range []struct{ upper, lower int }{{0, 2}, {5, 2}, {2, 0}, {2, 5}} {
		if _, err := m.FarinelliTibilettiRatio(c.upper, c.lower); err == nil {
			t.Errorf("orders (%d, %d): expected an error", c.upper, c.lower)
		}
	}
}

func TestRachevRatioMatchesPerformanceAnalyticsOutput(t *testing.T) {
	t.Parallel()
	check := func(alpha float64, refs map[float64]map[float64][]float64) {
		for beta, bundle := range refs {
			for rf, expected := range bundle {
				actual := runStream(t, yearlyRf(rf), func(m *Measures) float64 {
					return must(m.RachevRatio(alpha, beta))
				})
				assertSeriesEqual(t, actual, expected, places(14),
					fmt.Sprintf("Rachev ratio (alpha %v beta %v Rf %v)", alpha, beta, rf))
			}
		}
	}
	check(0.05, referencedata.RachevRatioExpectedValuesByBetaRfAlfa0_05)
	check(0.1, referencedata.RachevRatioExpectedValuesByBetaRfAlfa0_1)
}

func TestRachevRatioValidation(t *testing.T) {
	t.Parallel()
	m := makeMeasures(t, 0, 0, 0, false, false)
	// With fewer than two observations NaN is returned before validation.
	m.AddReturn(0.01, 0.01)
	if v, err := m.RachevRatio(0, 2); err != nil || !math.IsNaN(v) {
		t.Errorf("n < 2: expected NaN, nil; got %v, %v", v, err)
	}
	m.AddReturn(-0.01, 0.01)
	for _, c := range []struct{ alpha, beta float64 }{{0, 0.1}, {1, 0.1}, {0.1, 0}, {0.1, 1}} {
		if _, err := m.RachevRatio(c.alpha, c.beta); err == nil {
			t.Errorf("alpha %v beta %v: expected an error", c.alpha, c.beta)
		}
	}
}
