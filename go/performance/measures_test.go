package performance

import (
	"errors"
	"fmt"
	"math"
	"testing"
)

// publicMeasure is a public measure of Measures evaluated with the Python
// default arguments.
type publicMeasure struct {
	name string
	eval func(*Measures) (any, error)
}

func pf(name string, f func(*Measures) float64) publicMeasure {
	return publicMeasure{name, func(m *Measures) (any, error) { return f(m), nil }}
}

func pfe(name string, f func(*Measures) (float64, error)) publicMeasure {
	return publicMeasure{name, func(m *Measures) (any, error) { return f(m) }}
}

func ps(name string, f func(*Measures) []float64) publicMeasure {
	return publicMeasure{name, func(m *Measures) (any, error) { return f(m), nil }}
}

// publicMeasures is the counterpart of the inspect-based Python
// public_measures(): every public property, and every public method called
// with its default arguments, except the mutators Reset and AddReturn.
var publicMeasures = []publicMeasure{
	pf("active_premium", (*Measures).ActivePremium),
	pf("adjusted_sharpe_ratio", (*Measures).AdjustedSharpeRatio),
	pf("adjusted_sharpe_ratio_skew_only", (*Measures).AdjustedSharpeRatioSkewOnly),
	pf("appraisal_ratio", (*Measures).AppraisalRatio),
	pf("autocorrelation_penalty", (*Measures).AutocorrelationPenalty),
	pf("bernardo_ledoit_ratio", (*Measures).BernardoLedoitRatio),
	pfe("bias_ratio", func(m *Measures) (float64, error) { return m.BiasRatio(1.0) }),
	pf("burke_ratio", (*Measures).BurkeRatio),
	pf("burke_ratio_modified", (*Measures).BurkeRatioModified),
	pf("calmar_ratio", (*Measures).CalmarRatio),
	pfe("cdar_alpha", func(m *Measures) (float64, error) { return m.CdarAlpha(0.95) }),
	pfe("cdar_average", func(m *Measures) (float64, error) { return m.CdarAverage(0.95) }),
	pfe("cdar_beta", func(m *Measures) (float64, error) { return m.CdarBeta(0.95) }),
	pfe("cdar_discrete", func(m *Measures) (float64, error) { return m.CdarDiscrete(0.95) }),
	pf("compound_annual_growth_rate", (*Measures).CompoundAnnualGrowthRate),
	pf("cumulative_geometric_return", (*Measures).CumulativeGeometricReturn),
	pf("d_ratio", (*Measures).DRatio),
	pf("down_number_ratio", (*Measures).DownNumberRatio),
	pf("down_percentage_ratio", (*Measures).DownPercentageRatio),
	pf("downside_capture_ratio", func(m *Measures) float64 { return m.DownsideCaptureRatio(true) }),
	pf("downside_deviation", (*Measures).DownsideDeviation),
	pf("downside_deviation_subset", (*Measures).DownsideDeviationSubset),
	pf("downside_frequency", (*Measures).DownsideFrequency),
	pf("downside_potential", (*Measures).DownsidePotential),
	pf("downside_sharpe_ratio", (*Measures).DownsideSharpeRatio),
	pf("drawdown_average", (*Measures).DrawdownAverage),
	pf("drawdown_average_length", (*Measures).DrawdownAverageLength),
	pf("drawdown_average_peak_to_trough", (*Measures).DrawdownAveragePeakToTrough),
	pf("drawdown_average_recovery", (*Measures).DrawdownAverageRecovery),
	pf("drawdown_deviation", (*Measures).DrawdownDeviation),
	ps("drawdowns_continuous_runs", func(m *Measures) []float64 { return m.DrawdownsContinuousRuns(0) }),
	ps("drawdowns_cumulative", (*Measures).DrawdownsCumulative),
	ps("drawdowns_high_watermark", (*Measures).DrawdownsHighWatermark),
	pf("es_cornish_fisher", func(m *Measures) float64 { return m.EsCornishFisher(0.95) }),
	pf("es_gaussian", func(m *Measures) float64 { return m.EsGaussian(0.95) }),
	pf("es_historical", func(m *Measures) float64 { return m.EsHistorical(0.95) }),
	pf("fama_beta", (*Measures).FamaBeta),
	pfe("farinelli_tibiletti_ratio", func(m *Measures) (float64, error) { return m.FarinelliTibilettiRatio(2, 2) }),
	pf("gain_loss_ratio", (*Measures).GainLossRatio),
	pf("gain_to_pain_ratio", (*Measures).GainToPainRatio),
	pf("geometric_mean_return", (*Measures).GeometricMeanReturn),
	pf("hurst_exponent", (*Measures).HurstExponent),
	pf("information_ratio", (*Measures).InformationRatio),
	pf("information_ratio_modified", (*Measures).InformationRatioModified),
	{"is_normal_distribution", func(m *Measures) (any, error) { return m.IsNormalDistribution(0.95) }},
	pf("jarque_bera_normality_test_statistic", (*Measures).JarqueBeraNormalityTestStatistic),
	pf("jensen_alpha", (*Measures).JensenAlpha),
	pf("jensen_alpha_alternative", (*Measures).JensenAlphaAlternative),
	pf("jensen_alpha_modified", (*Measures).JensenAlphaModified),
	pf("k_ratio", (*Measures).KRatio),
	pf("kappa_1_ratio", (*Measures).Kappa1Ratio),
	pf("kappa_2_ratio", (*Measures).Kappa2Ratio),
	pf("kappa_3_ratio", (*Measures).Kappa3Ratio),
	pf("kappa_4_ratio", (*Measures).Kappa4Ratio),
	pf("kelly_ratio", (*Measures).KellyRatio),
	pf("kelly_ratio_full", (*Measures).KellyRatioFull),
	pf("kurtosis", (*Measures).Kurtosis),
	pf("kurtosis_excess", (*Measures).KurtosisExcess),
	pf("kurtosis_moment", (*Measures).KurtosisMoment),
	pf("kurtosis_sample", (*Measures).KurtosisSample),
	pf("kurtosis_sample_corrected", (*Measures).KurtosisSampleCorrected),
	pf("kurtosis_sample_excess", (*Measures).KurtosisSampleExcess),
	pf("loss_rate", (*Measures).LossRate),
	pf("m_squared", (*Measures).MSquared),
	pf("m_squared_excess", (*Measures).MSquaredExcess),
	pf("m_squared_sortino", (*Measures).MSquaredSortino),
	pf("martin_ratio", (*Measures).MartinRatio),
	pf("mean_absolute_deviation_ratio", (*Measures).MeanAbsoluteDeviationRatio),
	pf("mean_loss_return", (*Measures).MeanLossReturn),
	pf("mean_non_zero_return", (*Measures).MeanNonZeroReturn),
	pf("mean_win_return", (*Measures).MeanWinReturn),
	pf("min_drawdowns_cumulative", (*Measures).MinDrawdownsCumulative),
	pf("modigliani", (*Measures).Modigliani),
	pf("omega_excess_return", (*Measures).OmegaExcessReturn),
	pf("omega_ratio", (*Measures).OmegaRatio),
	pf("omega_sharpe_ratio", (*Measures).OmegaSharpeRatio),
	pf("overall_capture_ratio", func(m *Measures) float64 { return m.OverallCaptureRatio(true) }),
	pf("pain_index", (*Measures).PainIndex),
	pf("pain_ratio", (*Measures).PainRatio),
	pf("probabilistic_sharpe_ratio", func(m *Measures) float64 { return m.ProbabilisticSharpeRatio(0.0) }),
	pf("probabilistic_sharpe_ratio_full", func(m *Measures) float64 { return m.ProbabilisticSharpeRatioFull(0.0) }),
	pf("probabilistic_sharpe_ratio_gaussian", func(m *Measures) float64 { return m.ProbabilisticSharpeRatioGaussian(0.0) }),
	pf("probabilistic_sharpe_ratio_symmetric", func(m *Measures) float64 { return m.ProbabilisticSharpeRatioSymmetric(0.0) }),
	pf("prospect_ratio", func(m *Measures) float64 { return m.ProspectRatio(2.25) }),
	pf("prospect_ratio_performance_analytics", (*Measures).ProspectRatioPerformanceAnalytics),
	pfe("rachev_ratio", func(m *Measures) (float64, error) { return m.RachevRatio(0.1, 0.1) }),
	pf("reward_to_conditional_drawdown", func(m *Measures) float64 { return m.RewardToConditionalDrawdown(0.95) }),
	pf("reward_to_es_ratio_cornish_fisher", func(m *Measures) float64 { return m.RewardToEsRatioCornishFisher(0.95) }),
	pf("reward_to_es_ratio_gaussian", func(m *Measures) float64 { return m.RewardToEsRatioGaussian(0.95) }),
	pf("reward_to_es_ratio_historical", func(m *Measures) float64 { return m.RewardToEsRatioHistorical(0.95) }),
	pf("reward_to_var_ratio_cornish_fisher", func(m *Measures) float64 { return m.RewardToVarRatioCornishFisher(0.95) }),
	pf("reward_to_var_ratio_gaussian", func(m *Measures) float64 { return m.RewardToVarRatioGaussian(0.95) }),
	pf("reward_to_var_ratio_historical", func(m *Measures) float64 { return m.RewardToVarRatioHistorical(0.95) }),
	pf("semi_deviation", (*Measures).SemiDeviation),
	pf("sfm_alpha", (*Measures).SfmAlpha),
	pf("sfm_beta", (*Measures).SfmBeta),
	pf("sfm_beta_bear", (*Measures).SfmBetaBear),
	pf("sfm_beta_bull", (*Measures).SfmBetaBull),
	pf("sfm_r2", (*Measures).SfmR2),
	pf("sfm_risk_premium", (*Measures).SfmRiskPremium),
	pf("sharpe_ratio", (*Measures).SharpeRatio),
	pf("sharpe_ratio_es_cornish_fisher", func(m *Measures) float64 { return m.SharpeRatioEsCornishFisher(0.95) }),
	pf("sharpe_ratio_es_gaussian", func(m *Measures) float64 { return m.SharpeRatioEsGaussian(0.95) }),
	pf("sharpe_ratio_es_historical", func(m *Measures) float64 { return m.SharpeRatioEsHistorical(0.95) }),
	pf("sharpe_ratio_var_cornish_fisher", func(m *Measures) float64 { return m.SharpeRatioVarCornishFisher(0.95) }),
	pf("sharpe_ratio_var_gaussian", func(m *Measures) float64 { return m.SharpeRatioVarGaussian(0.95) }),
	pf("sharpe_ratio_var_historical", func(m *Measures) float64 { return m.SharpeRatioVarHistorical(0.95) }),
	pf("skewness", (*Measures).Skewness),
	pf("skewness_fisher", (*Measures).SkewnessFisher),
	pf("skewness_kurtosis_ratio", (*Measures).SkewnessKurtosisRatio),
	pf("skewness_moment", (*Measures).SkewnessMoment),
	pf("skewness_sample", (*Measures).SkewnessSample),
	pf("sortino_ratio", (*Measures).SortinoRatio),
	pf("sortino_ratio_sqrt2", (*Measures).SortinoRatioSqrt2),
	pf("sortino_satchell_ratio", (*Measures).SortinoSatchellRatio),
	pf("specific_risk", (*Measures).SpecificRisk),
	pf("sterling_ratio", func(m *Measures) float64 { return m.SterlingRatio(0.1) }),
	pf("systematic_risk", (*Measures).SystematicRisk),
	pfe("tail_ratio", func(m *Measures) (float64, error) { return m.TailRatio(0.95) }),
	pf("timing_ratio", (*Measures).TimingRatio),
	pf("total_risk", (*Measures).TotalRisk),
	pf("tracking_error", (*Measures).TrackingError),
	pf("treynor_ratio", (*Measures).TreynorRatio),
	pf("treynor_ratio_modified", (*Measures).TreynorRatioModified),
	pf("ulcer_index", (*Measures).UlcerIndex),
	pf("up_number_ratio", (*Measures).UpNumberRatio),
	pf("up_percentage_ratio", (*Measures).UpPercentageRatio),
	pf("upside_capture_ratio", func(m *Measures) float64 { return m.UpsideCaptureRatio(true) }),
	pf("upside_frequency", (*Measures).UpsideFrequency),
	pf("upside_potential", (*Measures).UpsidePotential),
	pf("upside_potential_ratio", (*Measures).UpsidePotentialRatio),
	pf("upside_potential_ratio_subset", (*Measures).UpsidePotentialRatioSubset),
	pf("upside_potential_subset", (*Measures).UpsidePotentialSubset),
	pf("upside_risk", (*Measures).UpsideRisk),
	pf("upside_risk_subset", (*Measures).UpsideRiskSubset),
	pf("upside_variance", (*Measures).UpsideVariance),
	pf("upside_variance_subset", (*Measures).UpsideVarianceSubset),
	pf("var_cornish_fisher", func(m *Measures) float64 { return m.VarCornishFisher(0.95) }),
	pf("var_gaussian", func(m *Measures) float64 { return m.VarGaussian(0.95) }),
	pf("var_historical", func(m *Measures) float64 { return m.VarHistorical(0.95) }),
	pf("variability_skewness", (*Measures).VariabilitySkewness),
	pf("volatility_skewness", (*Measures).VolatilitySkewness),
	pf("win_rate", (*Measures).WinRate),
	pf("worst_drawdowns_cumulative", (*Measures).WorstDrawdownsCumulative),
}

// evaluateAll evaluates every public measure and fails on an error.
func evaluateAll(t *testing.T, m *Measures) {
	t.Helper()
	for _, pm := range publicMeasures {
		if _, err := pm.eval(m); err != nil {
			t.Errorf("%s: unexpected error %v", pm.name, err)
		}
	}
}

func TestPublicMeasuresCount(t *testing.T) {
	t.Parallel()
	// 106 properties and 38 methods in the Python reference.
	if len(publicMeasures) != 144 {
		t.Errorf("expected 144 public measures, got %d", len(publicMeasures))
	}
	seen := map[string]bool{}
	for _, pm := range publicMeasures {
		if seen[pm.name] {
			t.Errorf("duplicate public measure %s", pm.name)
		}
		seen[pm.name] = true
	}
}

func TestSeriesAssertionsRejectsTruncatedSeries(t *testing.T) {
	t.Parallel()
	if checkSeriesEqual([]float64{1.0}, []float64{1.0, 2.0}, places(15), "", 0) == nil {
		t.Error("truncated actual series accepted")
	}
	if checkSeriesEqual([]float64{1.0, 2.0}, []float64{1.0}, places(15), "", 0) == nil {
		t.Error("truncated expected series accepted")
	}
}

func TestSeriesAssertionsAcceptsGenerator(t *testing.T) {
	t.Parallel()
	// Go has no generators; a computed slice stands in for one.
	gen := []float64{}
	for _, x := range []float64{1.0, 2.0} {
		gen = append(gen, x)
	}
	assertSeriesEqual(t, gen, []float64{1.0, 2.0}, places(15), "generator")
}

func TestSeriesAssertionsRelativeToleranceForLargeReferenceValues(t *testing.T) {
	t.Parallel()
	if err := checkSeriesEqual([]float64{1e12 + 0.1}, []float64{1e12}, relTol(1e-12, 0), "", 0); err != nil {
		t.Error(err)
	}
	if checkSeriesEqual([]float64{1e12 + 2}, []float64{1e12}, relTol(1e-12, 0), "", 0) == nil {
		t.Error("relative tolerance exceeded but accepted")
	}
}

func TestNewMeasuresRejectsNonPositivePeriodsPerAnnum(t *testing.T) {
	t.Parallel()
	for _, ppa := range []float64{0, -1} {
		if _, err := NewMeasures(ppa, 0, 0, 0); err == nil || err.Error() != "periods_per_annum must be positive" {
			t.Errorf("periods per annum %v: expected error, got %v", ppa, err)
		}
	}
}

func TestEdgeCasesEmpty(t *testing.T) {
	t.Parallel()
	// No measure returns an error before the first return.
	m := makeMeasures(t, 0, 0, 0, false, false)
	evaluateAll(t, m)
}

func TestEdgeCasesSingleReturn(t *testing.T) {
	t.Parallel()
	m := makeMeasures(t, 0, 0, 0, false, false)
	m.AddReturn(0.01, 0.02)
	evaluateAll(t, m)
	if !math.IsNaN(m.SharpeRatio()) {
		t.Errorf("sharpe ratio: expected NaN, got %v", m.SharpeRatio())
	}
	assertAlmostEqual(t, m.CumulativeGeometricReturn(), 0.01, 15, "cumulative geometric return")
}

func TestEdgeCasesFirstNegativeReturnIsADrawdown(t *testing.T) {
	t.Parallel()
	// As in PerformanceAnalytics Drawdowns(), the high-water mark starts at
	// the initial equity 1.
	m := makeMeasures(t, 0, 0, 0, false, false)
	m.AddReturn(-0.05, -0.02)
	m.AddReturn(0.02, 0.01)
	assertSeriesEqual(t, m.DrawdownsHighWatermark(), []float64{-0.05, -0.031}, places(15), "drawdowns high watermark")
	assertSeriesEqual(t, m.DrawdownsCumulative(), []float64{-0.05, -0.031}, places(15), "drawdowns cumulative")
	assertAlmostEqual(t, m.WorstDrawdownsCumulative(), 0.05, 15, "worst drawdowns cumulative")
	a, b := 0.05, 0.031
	assertAlmostEqual(t, m.PainIndex(), (a+b)/2, 15, "pain index")
	assertAlmostEqual(t, m.DrawdownAverage(), 0.05, 15, "drawdown average")
}

func TestEdgeCasesLongDailySeries(t *testing.T) {
	t.Parallel()
	// Every measure works with more observations than periods per annum.
	rng := newRNG(1)
	m := newTestMeasures(t, 252.0, 0, 0, 0)
	for range 300 {
		r := gauss(rng, 0.0005, 0.01)
		b := gauss(rng, 0.0004, 0.01)
		m.AddReturn(r, b)
	}
	evaluateAll(t, m)
	if v := m.AutocorrelationPenalty(); math.IsNaN(v) || math.IsInf(v, 0) {
		t.Errorf("autocorrelation penalty: expected finite, got %v", v)
	}
}

func TestRollingWindowRollingMatchesFresh(t *testing.T) {
	t.Parallel()
	// At every step, including while the window is still filling, every
	// public measure of a rolling-window instance equals that of a fresh
	// instance fed only the returns in the window.
	rng := newRNG(42)
	randomReturns := make([]float64, 150)
	for i := range randomReturns {
		randomReturns[i] = gauss(rng, 0.002, 0.03)
	}
	randomBenchmark := make([]float64, 150)
	for i := range randomBenchmark {
		randomBenchmark[i] = gauss(rng, 0.001, 0.025)
	}
	configs := []struct {
		window                   int
		periodsPerAnnum, rf, mar float64
		returns, benchmark       []float64
	}{
		{10, 1, 0.0, 0.0, baconPortfolioReturns, baconBenchmarkReturns},
		{30, 12, 0.05, 0.03, randomReturns, randomBenchmark},
	}
	for _, cfg := range configs {
		window := cfg.window
		rolling := newTestMeasures(t, cfg.periodsPerAnnum, cfg.rf, cfg.mar, window)
		for i := range cfg.returns {
			rolling.AddReturn(cfg.returns[i], cfg.benchmark[i])
			fresh := newTestMeasures(t, cfg.periodsPerAnnum, cfg.rf, cfg.mar, 0)
			for j := max(0, i-window+1); j < i+1; j++ {
				fresh.AddReturn(cfg.returns[j], cfg.benchmark[j])
			}
			for _, pm := range publicMeasures {
				prefix := fmt.Sprintf("window %d step %d %s", window, i, pm.name)
				actual, errA := pm.eval(rolling)
				expected, errE := pm.eval(fresh)
				if !errors.Is(errA, errE) {
					t.Errorf("%s: errors differ: %v vs %v", prefix, errA, errE)
					continue
				}
				switch e := expected.(type) {
				case []float64:
					a := actual.([]float64)
					assertSeriesEqual(t, a, e, places(12), prefix)
				case bool:
					if actual.(bool) != e {
						t.Errorf("%s: expected %v, got %v", prefix, e, actual)
					}
				case float64:
					a := actual.(float64)
					tol := places(15)
					if !math.IsNaN(e) && !math.IsInf(e, 0) {
						tol = delta(1e-12 * math.Max(1.0, math.Abs(e)))
					}
					assertFloatEqual(t, a, e, tol, prefix)
				}
			}
		}
	}
}
