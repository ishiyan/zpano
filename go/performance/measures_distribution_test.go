package performance

import (
	"fmt"
	"math"
	"testing"

	"zpano/performance/referencedata"
	"zpano/streamingkbn"
)

func TestAutocorrelationPenaltyMetamorphicProperties(t *testing.T) {
	t.Parallel()
	// Since we don't have data for it, we test only its metamorphic
	// properties.
	penalty := (*Measures).AutocorrelationPenalty
	n := bacon2023PortfolioLen

	// Constant returns.
	returns := repeat(0.01, n)
	actual := runStream(t, stream{daily: true, returns: returns, benchmark: returns}, penalty)
	assertFloatEqual(t, actual[n-1], 1.0, places(15), "autocorrelation penalty (constant)")

	// Too few observations.
	assertFloatEqual(t, actual[0], 1.0, places(15), "autocorrelation penalty (len=0)")
	assertFloatEqual(t, actual[1], 1.0, places(15), "autocorrelation penalty (len=1)")

	// Positive autocorrelation.
	returns = make([]float64, n)
	for i := range returns {
		returns[i] = 0.01 * float64(i)
	}
	actual = runStream(t, stream{daily: true, returns: returns, benchmark: returns}, penalty)
	assertFloatEqual(t, actual[n-1], 2.722393904531189, places(15), "autocorrelation penalty (positive)")

	// Negative autocorrelation.
	returns = make([]float64, n)
	for i := range returns {
		if i%2 == 0 {
			returns[i] = 0.01
		} else {
			returns[i] = -0.01
		}
	}
	actual = runStream(t, stream{daily: true, returns: returns, benchmark: returns}, penalty)
	assertFloatEqual(t, actual[n-1], 0.16903085094570597, places(15), "autocorrelation penalty (negative)")

	// Scale and translation invariance.
	expected := runStream(t, stream{daily: true}, penalty)
	for _, scale := range []float64{4.2, -4.2} {
		for _, shift := range []float64{0.042, -0.042} {
			transformed := make([]float64, baconPortfolioLen)
			for i, r := range baconPortfolioReturns {
				transformed[i] = scale*r + shift
			}
			actual = runStream(t, stream{daily: true, returns: transformed, benchmark: transformed}, penalty)
			assertSeriesEqual(t, actual, expected, places(15),
				fmt.Sprintf("autocorrelation penalty (transform) scale %v shift %v", scale, shift))
		}
	}
}

func TestCumulativeGeometricReturnMatchesPerformanceAnalyticsOutput(t *testing.T) {
	t.Parallel()
	expected := referencedata.CumulativeGeometricReturnExpectedValues
	f := (*Measures).CumulativeGeometricReturn
	assertSeriesEqual(t, runStream(t, stream{}, f), expected, places(14), "cumulative geometric return (yearly)")
	assertSeriesEqual(t, runStream(t, stream{monthly: true}, f), expected, places(14), "cumulative geometric return (monthly)")
	assertSeriesEqual(t, runStream(t, stream{daily: true}, f), expected, places(14), "cumulative geometric return (daily)")
}

func TestGeometricMeanReturnMatchesPerformanceAnalyticsOutput(t *testing.T) {
	t.Parallel()
	expected := referencedata.GeometricMeanReturnExpectedValuesGeometric
	f := (*Measures).GeometricMeanReturn
	assertSeriesEqual(t, runStream(t, stream{}, f), expected, places(15), "geometric mean return (yearly)")
	assertSeriesEqual(t, runStream(t, stream{monthly: true}, f), expected, places(15), "geometric mean return (monthly)")
	assertSeriesEqual(t, runStream(t, stream{daily: true}, f), expected, places(15), "geometric mean return (daily)")
}

func TestCompoundAnnualGrowthRateAnnualizedReturnDefinition(t *testing.T) {
	t.Parallel()
	calculate := func(m *Measures, ppa float64) float64 {
		growth := prod(onePlus(m.returns))
		return math.Pow(growth, ppa/float64(len(m.returns))) - 1
	}
	f := (*Measures).CompoundAnnualGrowthRate

	expected := runStream(t, stream{}, func(m *Measures) float64 { return calculate(m, 1) })
	assertSeriesEqual(t, runStream(t, stream{}, f), expected, places(15), "compound annual growth rate (yearly)")

	expected = runStream(t, stream{monthly: true}, func(m *Measures) float64 { return calculate(m, 12) })
	assertSeriesEqual(t, runStream(t, stream{monthly: true}, f), expected, places(14), "compound annual growth rate (monthly)")

	expected = runStream(t, stream{daily: true}, func(m *Measures) float64 { return calculate(m, 252) })
	assertSeriesEqual(t, runStream(t, stream{daily: true}, f), expected, places(11), "compound annual growth rate (daily)")
}

func skewnessByMethod(method string) func(*Measures) float64 {
	switch method {
	case "moment":
		return (*Measures).SkewnessMoment
	case "fisher":
		return (*Measures).SkewnessFisher
	case "sample":
		return (*Measures).SkewnessSample
	}
	panic("unknown skewness method " + method)
}

func TestSkewnessMatchesPerformanceAnalyticsOutput(t *testing.T) {
	t.Parallel()
	// RawMomentsKleinKBN is created with bias=true, so only the 'moment'
	// method matches Skewness. Calculation doesn't depend on periods per
	// annum, so we use the yearly default.
	for method, expected := range referencedata.SkewnessExpectedValuesByMethod {
		name := "skewness_" + method
		assertSeriesEqual(t, runStream(t, stream{}, skewnessByMethod(method)), expected, places(14), name)
		if method == "moment" {
			assertSeriesEqual(t, runStream(t, stream{}, (*Measures).Skewness), expected, places(14), "skewness")
		}
	}
}

func TestSkewnessRawMomentsKleinKBN(t *testing.T) {
	t.Parallel()
	// Mapping between PerformanceAnalytics naming and RawMomentsKleinKBN:
	// moment: bias=true; fisher: bias=false; sample: SkewnessSample.
	for method, expected := range referencedata.SkewnessExpectedValuesByMethod {
		bias := method != "fisher"
		kbn := streamingkbn.NewRawMomentsKleinKBN(1, bias, true)
		for i := range baconPortfolioLen {
			kbn.Update(baconPortfolioReturns[i])
			switch method {
			case "moment":
				assertFloatEqual(t, kbn.SkewnessMoment(), kbn.Skewness(), places(15),
					fmt.Sprintf("step %d skewness_%s vs. skewness", i, method))
			case "fisher":
				assertFloatEqual(t, kbn.SkewnessFisher(), kbn.Skewness(), places(15),
					fmt.Sprintf("step %d skewness_%s vs. skewness", i, method))
			default: // sample
				assertFloatEqual(t, kbn.SkewnessSample(), expected[i], places(14),
					fmt.Sprintf("step %d skewness_%s", i, method))
			}
		}
	}
}

func kurtosisByMethod(method string) func(*Measures) float64 {
	switch method {
	case "excess":
		return (*Measures).KurtosisExcess
	case "moment":
		return (*Measures).KurtosisMoment
	case "sample_excess":
		return (*Measures).KurtosisSampleExcess
	case "sample_corrected":
		return (*Measures).KurtosisSampleCorrected
	}
	panic("unknown kurtosis method " + method)
}

func TestKurtosisMatchesPerformanceAnalyticsOutput(t *testing.T) {
	t.Parallel()
	// RawMomentsKleinKBN is created with bias=true, fisher=true, so Kurtosis
	// is always 'excess'.
	for method, expected := range referencedata.KurtosisExpectedValuesByMethod {
		name := "kurtosis_" + method
		assertSeriesEqual(t, runStream(t, stream{}, kurtosisByMethod(method)), expected, places(13), name)
		if method == "excess" {
			assertSeriesEqual(t, runStream(t, stream{}, (*Measures).Kurtosis), expected, places(13), "kurtosis")
		}
	}
}

func TestKurtosisRawMomentsKleinKBN(t *testing.T) {
	t.Parallel()
	// Mapping between R naming and RawMomentsKleinKBN parameters:
	// excess: bias=true, fisher=true; moment: bias=true, fisher=false;
	// sample_excess: bias=false, fisher=true; sample: KurtosisSampleCorrected
	// (Kurtosis with bias=false, fisher=false returns KurtosisSample).
	for method, expected := range referencedata.KurtosisExpectedValuesByMethod {
		bias, fisher := true, true
		switch method {
		case "moment":
			fisher = false
		case "sample_corrected":
			bias, fisher = false, false
		case "sample_excess":
			bias = false
		}
		kbn := streamingkbn.NewRawMomentsKleinKBN(1, bias, fisher)
		for i := range baconPortfolioLen {
			kbn.Update(baconPortfolioReturns[i])
			var actual float64
			switch method {
			case "excess":
				actual = kbn.KurtosisExcess()
			case "moment":
				actual = kbn.KurtosisMoment()
			case "sample_corrected":
				actual = kbn.KurtosisSampleCorrected()
			default: // sample_excess
				actual = kbn.KurtosisSampleExcess()
			}
			dispatched := actual
			if method == "sample_corrected" {
				dispatched = kbn.KurtosisSample()
			}
			assertFloatEqual(t, dispatched, kbn.Kurtosis(), places(15),
				fmt.Sprintf("step %d kurtosis_%s / kurtosis", i, method))
			assertFloatEqual(t, actual, expected[i], places(13),
				fmt.Sprintf("step %d kurtosis_%s", i, method))
		}
	}
}

func TestSkewnessKurtosisRatioMatchesPerformanceAnalyticsOutput(t *testing.T) {
	t.Parallel()
	expected := referencedata.SkewnessKurtosisRatioExpectedValues
	actual := runStream(t, stream{}, (*Measures).SkewnessKurtosisRatio)
	assertSeriesEqual(t, actual, expected, places(14), "skewness-kurtosis ratio")
}

func TestJarqueBeraNrmalityTestStatisticMatchesBacon3Output(t *testing.T) {
	t.Parallel()
	actual := runStream(t, stream{returns: bacon2023PortfolioReturns, benchmark: bacon2023PortfolioReturns},
		(*Measures).JarqueBeraNormalityTestStatistic)
	expected := 0.34 // Chapter 5, exhibit 5.4.
	assertFloatEqual(t, actual[bacon2023PortfolioLen-1], expected, places(2), "Jarque-Bera normality (bacon3)")
}

func TestJarqueBeraNrmalityTestStatisticMatchesScipyOutput(t *testing.T) {
	t.Parallel()
	expected := referencedata.JarqueBeraNormalityTestStatisticExpectedValues
	actual := runStream(t, stream{}, (*Measures).JarqueBeraNormalityTestStatistic)
	assertSeriesEqual(t, actual, expected, places(14), "Jarque-Bera normality (scipy)")
}

func TestIsNormalDistributionMockedJB(t *testing.T) {
	t.Parallel()
	// The Python test mocks the Jarque-Bera statistic; here the decision
	// rule used by IsNormalDistribution is tested directly.
	check := func(jb, confidence float64, want bool, msg string) {
		t.Helper()
		got, err := isNormalFromJB(jb, confidence)
		if err != nil {
			t.Errorf("%s: unexpected error %v", msg, err)
		} else if got != want {
			t.Errorf("%s: expected %v, got %v", msg, want, got)
		}
	}

	// Normality accepted: 5.0 < 5.991...
	check(5.0, 0.95, true, "normality accepted")
	// Normality rejected.
	check(10.0, 0.95, false, "normality rejected")
	// NaN statistic.
	check(math.NaN(), 0.95, false, "NaN statistic")
	// Invalid confidence.
	for _, c := range []float64{1.0, 0.0} {
		if _, err := isNormalFromJB(0.0, c); err == nil {
			t.Errorf("confidence %v: expected an error", c)
		}
	}
	// Custom confidence.
	check(8.0, 0.99, true, "custom confidence 8")
	check(10.0, 0.99, false, "custom confidence 10")

	// Invalid confidence is rejected even with insufficient data.
	m := newTestMeasures(t, 1, 0, 0, 0)
	if got, err := m.IsNormalDistribution(2.0); got || err == nil {
		t.Errorf("empty: expected false, error; got %v, %v", got, err)
	}
	addBacon(m, nil, nil)
	if _, err := m.IsNormalDistribution(1.0); err == nil || err.Error() != "confidence must be between 0 and 1" {
		t.Errorf("invalid confidence: expected error, got %v", err)
	}
	jb := m.JarqueBeraNormalityTestStatistic()
	want, _ := isNormalFromJB(jb, 0.95)
	if got := must(m.IsNormalDistribution(0.95)); got != want {
		t.Errorf("bacon: expected %v, got %v", want, got)
	}
}

func TestVarCornishFisherMatchesPerformanceAnalyticsOutput(t *testing.T) {
	t.Parallel()
	for p, expected := range referencedata.VarExpectedValuesByPCornishFisher {
		actual := runStream(t, stream{}, func(m *Measures) float64 { return m.VarCornishFisher(p) })
		assertSeriesEqual(t, actual, expected, places(9), fmt.Sprintf("var cornish-fisher p %v", p))
	}
}

func TestVarGaussianMatchesPerformanceAnalyticsOutput(t *testing.T) {
	t.Parallel()
	for p, expected := range referencedata.VarExpectedValuesByPGaussian {
		actual := runStream(t, stream{}, func(m *Measures) float64 { return m.VarGaussian(p) })
		assertSeriesEqual(t, actual, expected, places(9), fmt.Sprintf("var gaussian p %v", p))
	}
}

func TestVarHistoricalMatchesPerformanceAnalyticsOutput(t *testing.T) {
	t.Parallel()
	for p, expected := range referencedata.VarExpectedValuesByPHistorical {
		actual := runStream(t, stream{}, func(m *Measures) float64 { return m.VarHistorical(p) })
		n := 15
		if p == 0.999 {
			n = 4
		}
		assertSeriesEqual(t, actual, expected, places(n), fmt.Sprintf("var historical p %v", p))
	}
}

func TestEsCornishFisherMatchesPerformanceAnalyticsOutput(t *testing.T) {
	t.Parallel()
	for p, expected := range referencedata.EsExpectedValuesByPCornishFisher {
		actual := runStream(t, stream{}, func(m *Measures) float64 { return m.EsCornishFisher(p) })
		n := 7
		if p < 0.995 {
			n = 9
		} else if p < 0.999 {
			n = 8
		}
		assertSeriesEqual(t, actual, expected, places(n), fmt.Sprintf("es cornish-fisher p %v", p))
	}
}

func TestEsGaussianMatchesPerformanceAnalyticsOutput(t *testing.T) {
	t.Parallel()
	for p, expected := range referencedata.EsExpectedValuesByPGaussian {
		actual := runStream(t, stream{}, func(m *Measures) float64 { return m.EsGaussian(p) })
		n := 8
		if p < 0.995 {
			n = 9
		}
		assertSeriesEqual(t, actual, expected, places(n), fmt.Sprintf("es gaussian p %v", p))
	}
}

func TestEsHistoricalMatchesPerformanceAnalyticsOutput(t *testing.T) {
	t.Parallel()
	for p, expected := range referencedata.EsExpectedValuesByPHistorical {
		actual := runStream(t, stream{}, func(m *Measures) float64 { return m.EsHistorical(p) })
		assertSeriesEqual(t, actual, expected, places(15), fmt.Sprintf("es historical p %v", p))
	}
}

// rewardToRiskCase pairs reward_to_* ratios with the corresponding
// sharpe_ratio_* variant and risk measure.
type rewardToRiskCase struct {
	name   string
	reward func(*Measures, float64) float64
	sharpe func(*Measures, float64) float64
	risk   func(*Measures, float64) float64
}

var rewardToRiskCases = []rewardToRiskCase{
	{"var_historical", (*Measures).RewardToVarRatioHistorical, (*Measures).SharpeRatioVarHistorical, (*Measures).VarHistorical},
	{"var_gaussian", (*Measures).RewardToVarRatioGaussian, (*Measures).SharpeRatioVarGaussian, (*Measures).VarGaussian},
	{"var_cornish_fisher", (*Measures).RewardToVarRatioCornishFisher, (*Measures).SharpeRatioVarCornishFisher, (*Measures).VarCornishFisher},
	{"es_historical", (*Measures).RewardToEsRatioHistorical, (*Measures).SharpeRatioEsHistorical, (*Measures).EsHistorical},
	{"es_gaussian", (*Measures).RewardToEsRatioGaussian, (*Measures).SharpeRatioEsGaussian, (*Measures).EsGaussian},
	{"es_cornish_fisher", (*Measures).RewardToEsRatioCornishFisher, (*Measures).SharpeRatioEsCornishFisher, (*Measures).EsCornishFisher},
}

func TestRewardToVarEsRatiosZeroRiskFreeRateEqualsSharpeVariants(t *testing.T) {
	t.Parallel()
	// With a zero risk-free rate, they equal sharpe_ratio_var_* /
	// sharpe_ratio_es_*.
	for _, c := range rewardToRiskCases {
		reward := runStream(t, stream{}, func(m *Measures) float64 { return c.reward(m, 0.95) })
		sharpe := runStream(t, stream{}, func(m *Measures) float64 { return c.sharpe(m, 0.95) })
		assertSeriesEqualSkip(t, reward, sharpe, places(14), 1, "reward_to "+c.name)
	}
}

func TestRewardToVarEsRatiosDefinition(t *testing.T) {
	t.Parallel()
	annualRf := 0.05
	for _, c := range rewardToRiskCases {
		for _, confidence := range []float64{0.9, 0.95} {
			m := makeMeasures(t, 0, annualRf, 0, false, true)
			for i := range baconPortfolioLen {
				m.AddReturn(baconPortfolioReturns[i], baconBenchmarkReturns[i])
				excess := make([]float64, i+1)
				for j, r := range baconPortfolioReturns[:i+1] {
					excess[j] = r - m.RiskFreeRate()
				}
				excessMean := fsum(excess) / float64(i+1)
				denom := c.risk(m, confidence)
				expected := math.NaN()
				if denom != 0 {
					expected = excessMean / denom
				}
				actual := c.reward(m, confidence)
				assertFloatEqual(t, actual, expected, places(14),
					fmt.Sprintf("reward_to %s confidence %v step %d", c.name, confidence, i))
			}
		}
	}
}

func TestMeanAbsoluteDeviationRatioExactValues(t *testing.T) {
	t.Parallel()
	// mean / (Σ|r - mean| / n) on the Bacon portfolio returns, computed
	// with exact rational arithmetic.
	expected := []float64{
		math.NaN(), 1.2608695652173914, 1.5789473684210527, 0.6818181818181818,
		0.9, 1.129032258064516, 1.308695652173913, 1.2618556701030927,
		0.9623076923076923, 1.0358796296296295, 0.9166666666666666, 0.96045197740113,
		1.0325794291868604, 0.7659033078880407, 0.480644111906311, 0.5178463399879009,
		0.3424072265625, 0.276536312849162, 0.3738222796970257, 0.4428828239908482,
		0.29573420836751435, 0.3247753530166881, 0.3100659077291792, 0.289544235924933,
	}
	actual := runStream(t, stream{}, (*Measures).MeanAbsoluteDeviationRatio)
	assertSeriesEqual(t, actual, expected, places(15), "mean absolute deviation ratio")
}
