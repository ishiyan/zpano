package performance

import (
	"fmt"
	"math"
	"slices"
	"testing"

	"zpano/performance/referencedata"
)

func TestDrawdownsCumulativeMatchesPerformanceAnalyticsOutput(t *testing.T) {
	t.Parallel()
	actual := runStreamAny(t, stream{}, (*Measures).DrawdownsCumulative)
	for i, expected := range referencedata.DrawdownsCumulativeExpectedValuesByIndex {
		assertSeriesEqual(t, actual[int(i)], expected, places(15), fmt.Sprintf("drawdowns cumulative (i %v)", i))
	}
}

func TestDrawdownsCumulativeMatchesBacon2023Output(t *testing.T) {
	t.Parallel()
	actual := runStreamAny(t, stream{returns: bacon2023PortfolioReturns, benchmark: bacon2023PortfolioReturns},
		(*Measures).DrawdownsCumulative)
	assertSeriesEqual(t, actual[len(actual)-1], bacon2023DrawdownFromPeak, places(4), "drawdowns cumulative (bacon 2023)")
}

func TestMinDrawdownsCumulativeMatchesPerformanceAnalyticsOutput(t *testing.T) {
	t.Parallel()
	actual := runStream(t, stream{}, (*Measures).MinDrawdownsCumulative)
	assertSeriesEqual(t, actual, referencedata.MinDrawdownsCumulativeExpectedValues, places(15), "min drawdowns cumulative")
}

func TestWorstDrawdownsCumulativeMatchesPerformanceAnalyticsOutput(t *testing.T) {
	t.Parallel()
	actual := runStream(t, stream{}, (*Measures).WorstDrawdownsCumulative)
	assertSeriesEqual(t, actual, referencedata.MinDrawdownsCumulativeExpectedValuesInverted, places(15), "worst drawdowns cumulative")
}

func TestDrawdownsHighWatermarkMatchesPerformanceAnalyticsOutput(t *testing.T) {
	t.Parallel()
	actual := runStreamAny(t, stream{}, (*Measures).DrawdownsHighWatermark)
	for i, expected := range referencedata.DrawdownsHighWatermarkExpectedValuesByIndex {
		assertSeriesEqual(t, actual[int(i)], expected, places(15), fmt.Sprintf("drawdowns high_watermark (i %v)", i))
	}
}

func TestDrawdownsContinuousRunsMatchesBacon2023Output(t *testing.T) {
	t.Parallel()
	// The book's four-decimal drawdown values are approximate; the core run
	// tracker has separate exact rolling-window tests.
	actual := runStreamAny(t, stream{returns: bacon2023PortfolioReturns, benchmark: bacon2023PortfolioReturns},
		func(m *Measures) []float64 { return m.DrawdownsContinuousRuns(0) })
	assertSeriesEqual(t, actual[len(actual)-1], bacon2023DrawdownContinuousWithoutZeroes, delta(0.002),
		"drawdowns continuous runs (bacon 2023)")
}

func TestDrawdownsContinuousRunsMaxRuns(t *testing.T) {
	t.Parallel()
	m := makeMeasures(t, 0, 0, 0, false, false)
	if got := m.DrawdownsContinuousRuns(0); got == nil || len(got) != 0 {
		t.Errorf("empty: expected an empty slice, got %v", got)
	}
	addBacon(m, nil, nil)
	all := m.DrawdownsContinuousRuns(0)
	sorted := slices.Clone(all)
	slices.Sort(sorted)
	assertSeriesEqual(t, m.DrawdownsContinuousRuns(2), sorted[:2], places(15), "max runs 2")
	assertSeriesEqual(t, m.DrawdownsContinuousRuns(100), sorted, places(15), "max runs 100")
}

func TestCalmarRatioMatchesPerformanceAnalyticsOutput(t *testing.T) {
	t.Parallel()
	expected := referencedata.CalmarRatioExpectedValues
	assertSeriesEqual(t, runStream(t, stream{}, (*Measures).CalmarRatio), expected, places(13), "calmar ratio (yearly)")
	assertSeriesEqual(t, runStream(t, stream{daily: true}, (*Measures).CalmarRatio), expected, places(13), "calmar ratio (daily)")
}

func TestSterlingRatioMatchesPerformanceAnalyticsOutput(t *testing.T) {
	t.Parallel()
	for excess, expected := range referencedata.SterlingRatioExpectedValuesByExcess {
		f := func(m *Measures) float64 { return m.SterlingRatio(excess) }
		assertSeriesEqual(t, runStream(t, stream{}, f), expected, places(13), fmt.Sprintf("sterling ratio (yearly, excess %v)", excess))
		assertSeriesEqual(t, runStream(t, stream{daily: true}, f), expected, places(13), fmt.Sprintf("sterling ratio (daily, excess %v)", excess))
	}
}

func TestBurkeRatioMatchesPerformanceAnalyticsOutput(t *testing.T) {
	t.Parallel()
	// Generated reference data is yearly.
	assertByKey(t, referencedata.BurkeRatioExpectedValuesByRf, yearlyRf,
		(*Measures).BurkeRatio, fixed(places(11)), 0, "burke ratio (yearly) Rf")
}

func TestBurkeRatioModifiedMatchesPerformanceAnalyticsOutput(t *testing.T) {
	t.Parallel()
	assertByKey(t, referencedata.BurkeRatioExpectedValuesByRfModified, yearlyRf,
		(*Measures).BurkeRatioModified, fixed(places(11)), 0, "burke ratio modified Rf")
}

func TestPainIndexMatchesPerformanceAnalyticsOutput(t *testing.T) {
	t.Parallel()
	// R's drawdown series differs slightly from the high-water-mark series
	// used here. TestDocumentedFormulasDrawdownRiskAndRatios checks the
	// exact formula.
	expected := referencedata.PainIndexExpectedValues
	assertSeriesEqual(t, runStream(t, stream{}, (*Measures).PainIndex), expected, delta(0.00098), "pain index (yearly)")
	assertSeriesEqual(t, runStream(t, stream{daily: true}, (*Measures).PainIndex), expected, delta(0.00098), "pain index (daily)")
}

func TestPainRatioMatchesPerformanceAnalyticsOutput(t *testing.T) {
	t.Parallel()
	for rf, expected := range referencedata.PainRatioExpectedValuesByRf {
		tol := delta(0.111)
		if rf < 0.04 {
			tol = delta(0.016)
		}
		assertSeriesEqual(t, runStream(t, yearlyRf(rf), (*Measures).PainRatio), expected, tol,
			fmt.Sprintf("pain ratio (yearly, Rf %v)", rf))
		assertSeriesEqual(t, runStream(t, stream{daily: true, annualRiskFreeRate: annualize(rf, 252)}, (*Measures).PainRatio),
			expected, tol, fmt.Sprintf("pain ratio (daily, Rf %v)", rf))
	}
}

func TestUlcerIndexMatchesPerformanceAnalyticsOutput(t *testing.T) {
	t.Parallel()
	expected := referencedata.UlcerIndexExpectedValues
	assertSeriesEqual(t, runStream(t, stream{}, (*Measures).UlcerIndex), expected, delta(0.00192), "ulcer index (yearly)")
	assertSeriesEqual(t, runStream(t, stream{daily: true}, (*Measures).UlcerIndex), expected, delta(0.00192), "ulcer index (daily)")
}

func TestMartinRatioMatchesPerformanceAnalyticsOutput(t *testing.T) {
	t.Parallel()
	for rf, expected := range referencedata.MartinRatioExpectedValuesByRf {
		assertSeriesEqual(t, runStream(t, yearlyRf(rf), (*Measures).MartinRatio), expected, delta(0.0630),
			fmt.Sprintf("martin ratio (yearly, Rf %v)", rf))
		assertSeriesEqual(t, runStream(t, stream{daily: true, annualRiskFreeRate: annualize(rf, 252)}, (*Measures).MartinRatio),
			expected, delta(0.0630), fmt.Sprintf("martin ratio (daily, Rf %v)", rf))
	}
}

func assertYearlyAndDaily(t *testing.T, f func(*Measures) float64, expected []float64, tol tolerance, name string) {
	t.Helper()
	assertSeriesEqual(t, runStream(t, stream{}, f), expected, tol, name+" (yearly)")
	assertSeriesEqual(t, runStream(t, stream{daily: true}, f), expected, tol, name+" (daily)")
}

func TestDrawdownAverageMatchesPerformanceAnalyticsOutput(t *testing.T) {
	t.Parallel()
	assertYearlyAndDaily(t, (*Measures).DrawdownAverage, referencedata.DrawdownAverageExpectedValues, places(15), "drawdown average")
}

func TestDrawdownAverageLengthMatchesPerformanceAnalyticsOutput(t *testing.T) {
	t.Parallel()
	assertYearlyAndDaily(t, (*Measures).DrawdownAverageLength, referencedata.DrawdownAverageLengthExpectedValues, places(15), "drawdown average length")
}

func TestDrawdownAveragePeakToTroughMatchesPerformanceAnalyticsOutput(t *testing.T) {
	t.Parallel()
	assertYearlyAndDaily(t, (*Measures).DrawdownAveragePeakToTrough, referencedata.DrawdownAveragePeakToTroughExpectedValues, places(15), "drawdown average peak-to-trough")
}

func TestDrawdownAverageRecoveryMatchesPerformanceAnalyticsOutput(t *testing.T) {
	t.Parallel()
	assertYearlyAndDaily(t, (*Measures).DrawdownAverageRecovery, referencedata.DrawdownAverageRecoveryExpectedValues, places(15), "drawdown average recovery")
}

func TestDrawdownDeviationMatchesPerformanceAnalyticsOutput(t *testing.T) {
	t.Parallel()
	assertYearlyAndDaily(t, (*Measures).DrawdownDeviation, referencedata.DrawdownDeviationExpectedValues, places(15), "drawdown deviation")
}

func TestCDaRAverageMatchesPerformanceAnalyticsOutput(t *testing.T) {
	t.Parallel()
	// R and this implementation select the continuous drawdown tail
	// differently. TestDocumentedFormulasCdarAverageAndAlpha checks the
	// linear quantile.
	for p, expected := range referencedata.CdarExpectedValuesByPAverageGeometricInverted {
		f := func(m *Measures) float64 { return must(m.CdarAverage(p)) }
		assertSeriesEqual(t, runStream(t, stream{}, f), expected, delta(0.02938), fmt.Sprintf("CDaR average geometric (yearly) p %v", p))
		assertSeriesEqual(t, runStream(t, stream{daily: true}, f), expected, delta(0.02938), fmt.Sprintf("CDaR average geometric (daily) p %v", p))
	}
}

func TestCDaRDiscreteMatchesPerformanceAnalyticsOutput(t *testing.T) {
	t.Parallel()
	for p, expected := range referencedata.CdarExpectedValuesByPDiscreteGeometricInverted {
		f := func(m *Measures) float64 { return must(m.CdarDiscrete(p)) }
		assertSeriesEqual(t, runStream(t, stream{}, f), expected, places(15), fmt.Sprintf("CDaR discrete geometric (yearly) p %v", p))
		assertSeriesEqual(t, runStream(t, stream{daily: true}, f), expected, places(15), fmt.Sprintf("CDaR discrete geometric (daily) p %v", p))
	}
}

func TestCDaRValidation(t *testing.T) {
	t.Parallel()
	m := makeMeasures(t, 0, 0, 0, false, false)
	addBacon(m, nil, nil)
	for _, c := range []float64{0.0, 1.0, math.NaN()} {
		for name, f := range map[string]func(float64) (float64, error){
			"cdar_average": m.CdarAverage, "cdar_discrete": m.CdarDiscrete,
			"cdar_beta": m.CdarBeta, "cdar_alpha": m.CdarAlpha,
		} {
			if _, err := f(c); err == nil || err.Error() != "confidence must be between 0 and 1" {
				t.Errorf("%s confidence %v: expected error, got %v", name, c, err)
			}
		}
	}
}

func TestCDaRBetaDiscreteTailSelection(t *testing.T) {
	t.Parallel()
	m := makeMeasures(t, 0, 0, 0, false, false)
	for _, pb := range [][2]float64{{0.1, 0.1}, {-0.05, -0.1}, {0.2, 0.2}, {-0.1, -0.2}, {0.3, 0.3}, {-0.15, -0.3}} {
		m.AddReturn(pb[0], pb[1])
	}
	// At 50% confidence, two of three episodes are selected. The
	// denominator is the second-worst depth, -0.2.
	a, b, c, d := -0.15, 0.1, 2.0, -0.2
	assertAlmostEqual(t, must(m.CdarBeta(0.5)), (a-b)/(c*d), 14, "CDaR beta 0.5")
	e := -0.3
	assertAlmostEqual(t, must(m.CdarBeta(0.8)), a/e, 14, "CDaR beta 0.8")
}

func TestCDaRBetaMatchesPerformanceAnalyticsOutput(t *testing.T) {
	t.Parallel()
	for p, expected := range referencedata.CdarBetaExpectedValuesByPGeometric {
		f := func(m *Measures) float64 { return must(m.CdarBeta(p)) }
		assertSeriesEqual(t, runStream(t, stream{}, f), expected, places(13), fmt.Sprintf("CDaR beta geometric (yearly) p %v", p))
		assertSeriesEqual(t, runStream(t, stream{daily: true}, f), expected, places(13), fmt.Sprintf("CDaR beta geometric (daily) p %v", p))
	}
}

func TestCDaRBetaMathematicalProperties(t *testing.T) {
	t.Parallel()
	// With one selected episode, identical portfolio and benchmark returns
	// give the same numerator and denominator.
	m := makeMeasures(t, 0, 0, 0, false, false)
	for _, ret := range []float64{0.05, -0.1} {
		m.AddReturn(ret, ret)
	}
	assertFloatEqual(t, must(m.CdarBeta(0.95)), 1.0, places(15), "CDaR beta (one episode) identity")

	m = makeMeasures(t, 0, 0, 0, false, false)
	addBacon(m, nil, baconPortfolioReturns)
	assertFloatEqual(t, must(m.CdarBeta(0.95)), 1.0, places(14), "CDaR beta (Bacon) identity")

	// No drawdowns.
	mt := makeMeasures(t, 0, 0, 0, false, false)
	addBacon(mt, repeat(0.01, baconPortfolioLen), repeat(0.02, baconPortfolioLen))
	assertFloatEqual(t, must(mt.CdarBeta(0.95)), math.NaN(), places(15), "CDaR beta (geometric) no drawdowns")

	// Zero portfolio returns.
	mt = makeMeasures(t, 0, 0, 0, false, false)
	addBacon(mt, repeat(0, baconPortfolioLen), nil)
	assertFloatEqual(t, must(mt.CdarBeta(0.95)), 0.0, places(15), "CDaR beta (geometric) zero returns")

	for _, confidence := range []float64{0.0, 1.0} {
		if _, err := m.CdarBeta(confidence); err == nil {
			t.Errorf("confidence %v: expected an error", confidence)
		}
	}
}

func TestCDaRAlphaMatchesPerformanceAnalyticsOutput(t *testing.T) {
	t.Parallel()
	// PerformanceAnalytics hard-codes 12 periods when annualizing the means;
	// this implementation uses periods per annum, so monthly data is used.
	for p, expected := range referencedata.CdarAlphaExpectedValuesByPGeometric {
		actual := runStream(t, stream{monthly: true}, func(m *Measures) float64 { return must(m.CdarAlpha(p)) })
		assertSeriesEqual(t, actual, expected, places(14), fmt.Sprintf("CDaR alpha geometric (monthly) p %v", p))
	}
}

func TestCDaRAlphaMathematicalProperties(t *testing.T) {
	t.Parallel()
	// Identity (portfolio == benchmark).
	m := makeMeasures(t, 0, 0, 0, false, true)
	addBacon(m, nil, baconPortfolioReturns)
	assertFloatEqual(t, must(m.CdarAlpha(0.95)), 0.0, places(14), "CDaR alpha (geometric) identity")

	// No drawdowns.
	mt := makeMeasures(t, 0, 0, 0, false, false)
	addBacon(mt, repeat(0.01, baconPortfolioLen), repeat(0.02, baconPortfolioLen))
	assertFloatEqual(t, must(mt.CdarAlpha(0.95)), math.NaN(), places(15), "CDaR alpha (geometric) no drawdowns")

	// Zero portfolio returns.
	mt = makeMeasures(t, 0, 0, 0, false, false)
	addBacon(mt, repeat(0, baconPortfolioLen), nil)
	assertFloatEqual(t, must(mt.CdarAlpha(0.95)), 0.0, places(15), "CDaR alpha (geometric) zero returns")
}

func TestRewardToConditionalDrawdownDefinition(t *testing.T) {
	t.Parallel()
	// Geometric mean return divided by the mean magnitude of the worst
	// max(1, int(n·(1 - confidence))) drawdowns.
	for _, confidence := range []float64{0.8, 0.95} {
		m := makeMeasures(t, 0, 0, 0, false, false)
		for i := range baconPortfolioLen {
			m.AddReturn(baconPortfolioReturns[i], baconBenchmarkReturns[i])
			dd := m.DrawdownsHighWatermark()
			slices.Sort(dd)
			tail := dd[:max(1, int(float64(len(dd))*(1-confidence)))]
			cdar := -pySum(tail) / float64(len(tail))
			expected := math.NaN()
			if cdar != 0 {
				expected = m.GeometricMeanReturn() / cdar
			}
			assertFloatEqual(t, must(m.RewardToConditionalDrawdown(confidence)), expected, places(15),
				fmt.Sprintf("confidence %v step %d", confidence, i))
		}
	}
}

func TestDocumentedFormulasDrawdownRiskAndRatios(t *testing.T) {
	t.Parallel()
	for _, rf := range []float64{0.0, 0.05} {
		m := makeMeasures(t, 0, rf, 0, false, false)
		for i := range baconPortfolioLen {
			m.AddReturn(baconPortfolioReturns[i], baconBenchmarkReturns[i])
			drawdowns := m.DrawdownsHighWatermark()
			n := float64(i + 1)
			pain := -fsum(drawdowns) / n
			sq := make([]float64, len(drawdowns))
			for j, x := range drawdowns {
				sq[j] = x * x
			}
			ulcer := math.Sqrt(fsum(sq) / n)
			geometricReturn := math.Pow(prod(onePlus(baconPortfolioReturns[:i+1])), 1/n) - 1
			prefix := fmt.Sprintf("rf %v step %d", rf, i)
			assertAlmostEqual(t, m.PainIndex(), pain, 14, prefix+" pain index")
			assertAlmostEqual(t, m.UlcerIndex(), ulcer, 14, prefix+" ulcer index")
			if pain > 0 {
				assertAlmostEqual(t, m.PainRatio(), (geometricReturn-rf)/pain, 12, prefix+" pain ratio")
			}
			if ulcer > 0 {
				assertAlmostEqual(t, m.MartinRatio(), (geometricReturn-rf)/ulcer, 12, prefix+" martin ratio")
			}
		}
	}
}

func TestDocumentedFormulasCdarAverageAndAlpha(t *testing.T) {
	t.Parallel()
	m := makeMeasures(t, 0, 0, 0, false, true)
	for i := range baconPortfolioLen {
		m.AddReturn(baconPortfolioReturns[i], baconBenchmarkReturns[i])
		drawdowns := m.DrawdownsHighWatermark()
		slices.Sort(drawdowns)
		for _, confidence := range []float64{0.9, 0.95} {
			position := (1 - confidence) * float64(len(drawdowns)-1)
			lo := int(position)
			q := drawdowns[lo] + (position-float64(lo))*(drawdowns[min(lo+1, len(drawdowns)-1)]-drawdowns[lo])
			tail := []float64{}
			for _, d := range drawdowns {
				if d <= q {
					tail = append(tail, d)
				}
			}
			expectedCdar := 0.0
			if q < 0 {
				expectedCdar = -fsum(tail) / float64(len(tail))
			}
			prefix := fmt.Sprintf("step %d confidence %v", i, confidence)
			assertAlmostEqual(t, must(m.CdarAverage(confidence)), expectedCdar, 14, prefix+" cdar average")
			beta := must(m.CdarBeta(confidence))
			if !math.IsNaN(beta) && !math.IsInf(beta, 0) {
				portfolioMean := fmean(baconPortfolioReturns[:len(drawdowns)])
				benchmarkMean := fmean(baconBenchmarkReturns[:len(drawdowns)])
				expectedAlpha := math.Pow(1+portfolioMean, 12) - 1 - beta*(math.Pow(1+benchmarkMean, 12)-1)
				assertAlmostEqual(t, must(m.CdarAlpha(confidence)), expectedAlpha, 13, prefix+" cdar alpha")
			}
		}
	}
}
