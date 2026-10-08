//! VaR / ES, reward-to-risk, partial moments and Sharpe family tests.

use super::*;
use crate::performance::reference_data as rd;

type Method = fn(&Measures, f64) -> f64;

// VaR / ES (calculation doesn't depend on periods per annum)

fn check_by_p(table: &[(f64, &[f64])], method: Method, places: impl Fn(f64) -> i32, label: &str) {
    for &(p, expected) in table {
        let actual = run(cfg(), |m| method(m, p));
        assert_series(&actual, expected, Places(places(p)), 0, &format!("{label} p {p}"));
    }
}

#[test]
fn test_var_cornish_fisher_matches_performance_analytics_output() {
    check_by_p(rd::var::EXPECTED_VALUES_BY_P_CORNISH_FISHER, Measures::var_cornish_fisher, |_| 9, "var cornish-fisher");
}

#[test]
fn test_var_gaussian_matches_performance_analytics_output() {
    check_by_p(rd::var::EXPECTED_VALUES_BY_P_GAUSSIAN, Measures::var_gaussian, |_| 9, "var gaussian");
}

#[test]
fn test_var_historical_matches_performance_analytics_output() {
    check_by_p(
        rd::var::EXPECTED_VALUES_BY_P_HISTORICAL,
        Measures::var_historical,
        |p| if p == 0.999 { 4 } else { 15 },
        "var historical",
    );
}

#[test]
fn test_es_cornish_fisher_matches_performance_analytics_output() {
    check_by_p(
        rd::es::EXPECTED_VALUES_BY_P_CORNISH_FISHER,
        Measures::es_cornish_fisher,
        |p| if p < 0.995 { 9 } else if p < 0.999 { 8 } else { 7 },
        "es cornish-fisher",
    );
}

#[test]
fn test_es_gaussian_matches_performance_analytics_output() {
    check_by_p(
        rd::es::EXPECTED_VALUES_BY_P_GAUSSIAN,
        Measures::es_gaussian,
        |p| if p < 0.995 { 9 } else { 8 },
        "es gaussian",
    );
}

#[test]
fn test_es_historical_matches_performance_analytics_output() {
    check_by_p(rd::es::EXPECTED_VALUES_BY_P_HISTORICAL, Measures::es_historical, |_| 15, "es historical");
}

#[test]
fn test_var_es_invalid_confidence_mirrors_python() {
    // NaN exactly where Python raises (indirectly, from percentile or
    // norm_ppf); Python's value everywhere else.
    let mut m = make_measures(0, 0.05, 0.0, YEARLY);
    add_bacon(&mut m, &BACON_PORTFOLIO_RETURNS, &BACON_BENCHMARK_RETURNS);

    // Historical: percentile q = 1 - c must be in [0, 1]; c = 0 and c = 1
    // are valid. Expected values from Python.
    let historical: [(Method, f64, f64); 6] = [
        (Measures::var_historical, -0.081, 0.065),
        (Measures::es_historical, -0.009000000000000003, 0.065),
        (Measures::reward_to_var_ratio_historical, 0.5061728395061729, -0.6307692307692307),
        (Measures::reward_to_es_ratio_historical, 4.5555555555555545, -0.6307692307692307),
        (Measures::sharpe_ratio_var_historical, 1.3225806451612905, -0.3565217391304348),
        (Measures::sharpe_ratio_es_historical, -1.0, -0.3565217391304348),
    ];
    for (i, (method, at_0, at_1)) in historical.iter().enumerate() {
        assert_float(method(&m, 0.0), *at_0, Places(15), &format!("historical {i} c=0"));
        assert_float(method(&m, 1.0), *at_1, Places(15), &format!("historical {i} c=1"));
        for confidence in [-0.5, 1.5, f64::NAN] {
            assert!(method(&m, confidence).is_nan(), "historical {i} confidence {confidence}");
        }
        assert!(!method(&m, 0.95).is_nan(), "historical {i} confidence 0.95");
    }

    // Gaussian / Cornish-Fisher: norm_ppf(1 - c) or norm_ppf(c) raises for
    // c <= 0 or c >= 1; a NaN c passes norm_ppf and yields NaN.
    let parametric: [Method; 12] = [
        Measures::var_gaussian,
        Measures::var_cornish_fisher,
        Measures::es_gaussian,
        Measures::es_cornish_fisher,
        Measures::reward_to_var_ratio_gaussian,
        Measures::reward_to_var_ratio_cornish_fisher,
        Measures::reward_to_es_ratio_gaussian,
        Measures::reward_to_es_ratio_cornish_fisher,
        Measures::sharpe_ratio_var_gaussian,
        Measures::sharpe_ratio_var_cornish_fisher,
        Measures::sharpe_ratio_es_gaussian,
        Measures::sharpe_ratio_es_cornish_fisher,
    ];
    for (i, method) in parametric.iter().enumerate() {
        for confidence in [0.0, 1.0, -0.5, 1.5, f64::NAN] {
            assert!(method(&m, confidence).is_nan(), "parametric {i} confidence {confidence}");
        }
        assert!(!method(&m, 0.95).is_nan(), "parametric {i} confidence 0.95");
    }
}

// TestRewardToVarEsRatios: reward_to_var_ratio_* and reward_to_es_ratio_*
// divide the mean excess return by the VaR/ES of the raw returns.

/// (name, risk measure, reward-to ratio, Sharpe variant)
const REWARD_METHODS: [(&str, Method, Method, Method); 6] = [
    (
        "var_historical",
        Measures::var_historical,
        Measures::reward_to_var_ratio_historical,
        Measures::sharpe_ratio_var_historical,
    ),
    ("var_gaussian", Measures::var_gaussian, Measures::reward_to_var_ratio_gaussian, Measures::sharpe_ratio_var_gaussian),
    (
        "var_cornish_fisher",
        Measures::var_cornish_fisher,
        Measures::reward_to_var_ratio_cornish_fisher,
        Measures::sharpe_ratio_var_cornish_fisher,
    ),
    (
        "es_historical",
        Measures::es_historical,
        Measures::reward_to_es_ratio_historical,
        Measures::sharpe_ratio_es_historical,
    ),
    ("es_gaussian", Measures::es_gaussian, Measures::reward_to_es_ratio_gaussian, Measures::sharpe_ratio_es_gaussian),
    (
        "es_cornish_fisher",
        Measures::es_cornish_fisher,
        Measures::reward_to_es_ratio_cornish_fisher,
        Measures::sharpe_ratio_es_cornish_fisher,
    ),
];

#[test]
fn test_reward_to_var_es_zero_risk_free_rate_equals_sharpe_variants() {
    for (name, _, reward, sharpe) in REWARD_METHODS {
        let r = run(cfg(), |m| reward(m, 0.95));
        let s = run(cfg(), |m| sharpe(m, 0.95));
        assert_series(&r, &s, Places(14), 1, &format!("reward_to {name}"));
    }
}

#[test]
fn test_reward_to_var_es_definition() {
    let annual_rf = 0.05;
    for (name, risk, reward, _) in REWARD_METHODS {
        for confidence in [0.9, 0.95] {
            let mut m = make_measures(0, annual_rf, 0.0, MONTHLY);
            for i in 0..BACON_PORTFOLIO_LEN {
                m.add_return(BACON_PORTFOLIO_RETURNS[i], BACON_BENCHMARK_RETURNS[i]);
                let excess: Vec<f64> = BACON_PORTFOLIO_RETURNS[..=i].iter().map(|r| r - m.risk_free_rate()).collect();
                let excess_mean = fsum(&excess) / (i + 1) as f64;
                let denom = risk(&m, confidence);
                let expected = if denom != 0.0 { excess_mean / denom } else { f64::NAN };
                assert_float(
                    reward(&m, confidence),
                    expected,
                    Places(14),
                    &format!("reward_to {name} confidence {confidence} step {i}"),
                );
            }
        }
    }
}

// TestMeanAbsoluteDeviationRatio

#[test]
fn test_mean_absolute_deviation_ratio_exact_values() {
    // mean / (sum|r - mean| / n), computed with exact rational arithmetic.
    let expected = [
        f64::NAN,
        1.2608695652173914,
        1.5789473684210527,
        0.6818181818181818,
        0.9,
        1.129032258064516,
        1.308695652173913,
        1.2618556701030927,
        0.9623076923076923,
        1.0358796296296295,
        0.9166666666666666,
        0.96045197740113,
        1.0325794291868604,
        0.7659033078880407,
        0.480644111906311,
        0.5178463399879009,
        0.3424072265625,
        0.276536312849162,
        0.3738222796970257,
        0.4428828239908482,
        0.29573420836751435,
        0.3247753530166881,
        0.3100659077291792,
        0.289544235924933,
    ];
    let actual = run(cfg(), Measures::mean_absolute_deviation_ratio);
    assert_series(&actual, &expected, Places(15), 0, "mean absolute deviation ratio");
}

// Upside / downside partial moments (calculation doesn't depend on
// periods per annum, so the yearly default is used)

fn check_by_mar(table: &[(f64, &[f64])], f: fn(&Measures) -> f64, places: i32, label: &str) {
    for &(mar, expected) in table {
        let actual = run(cfg().mar(mar), f);
        assert_series(&actual, expected, Places(places), 0, &format!("{label} MAR {mar}"));
    }
}

/// Daily variant: the periodic MAR is compounded to an annual MAR.
fn check_by_mar_daily(table: &[(f64, &[f64])], f: fn(&Measures) -> f64, places: i32, label: &str) {
    for &(mar, expected) in table {
        let actual = run(cfg().daily().mar(compound(mar, 252.0)), f);
        assert_series(&actual, expected, Places(places), 0, &format!("{label} daily MAR {mar}"));
    }
}

#[test]
fn test_upside_potential_ratio_matches_performance_analytics_output() {
    check_by_mar(
        rd::upside_potential_ratio::EXPECTED_VALUES_BY_MAR_FULL,
        Measures::upside_potential_ratio,
        14,
        "upside potential ratio (full)",
    );
}

#[test]
fn test_upside_potential_ratio_subset_matches_performance_analytics_output() {
    check_by_mar(
        rd::upside_potential_ratio::EXPECTED_VALUES_BY_MAR_SUBSET,
        Measures::upside_potential_ratio_subset,
        14,
        "upside potential ratio (subset)",
    );
}

#[test]
fn test_upside_frequency_matches_performance_analytics_output() {
    check_by_mar(rd::upside_frequency::EXPECTED_VALUES_BY_MAR, Measures::upside_frequency, 15, "upside frequency");
}

#[test]
fn test_upside_potential_matches_performance_analytics_output() {
    let table = rd::upside_risk::EXPECTED_VALUES_BY_MAR_POTENTIAL_FULL;
    check_by_mar(table, Measures::upside_potential, 15, "upside potential (full)");
    check_by_mar_daily(table, Measures::upside_potential, 15, "upside potential (full)");
}

#[test]
fn test_upside_potential_subset_matches_performance_analytics_output() {
    check_by_mar(
        rd::upside_risk::EXPECTED_VALUES_BY_MAR_POTENTIAL_SUBSET,
        Measures::upside_potential_subset,
        15,
        "upside potential (subset)",
    );
}

#[test]
fn test_upside_variance_matches_performance_analytics_output() {
    check_by_mar(
        rd::upside_risk::EXPECTED_VALUES_BY_MAR_VARIANCE_FULL,
        Measures::upside_variance,
        15,
        "upside variance (full)",
    );
}

#[test]
fn test_upside_variance_subset_matches_performance_analytics_output() {
    check_by_mar(
        rd::upside_risk::EXPECTED_VALUES_BY_MAR_VARIANCE_SUBSET,
        Measures::upside_variance_subset,
        15,
        "upside variance (subset)",
    );
}

#[test]
fn test_upside_risk_matches_performance_analytics_output() {
    check_by_mar(rd::upside_risk::EXPECTED_VALUES_BY_MAR_RISK_FULL, Measures::upside_risk, 15, "upside risk (full)");
}

#[test]
fn test_upside_risk_subset_matches_performance_analytics_output() {
    let table = rd::upside_risk::EXPECTED_VALUES_BY_MAR_RISK_SUBSET;
    check_by_mar(table, Measures::upside_risk_subset, 15, "upside risk (subset) yearly");
    check_by_mar_daily(table, Measures::upside_risk_subset, 15, "upside risk (subset)");
}

#[test]
fn test_semi_deviation_matches_performance_analytics_output() {
    let actual = run(cfg(), Measures::semi_deviation);
    assert_series(&actual, rd::semi_deviation::EXPECTED_VALUES, Places(15), 0, "semi-deviation");
}

#[test]
fn test_downside_deviation_matches_performance_analytics_output() {
    check_by_mar(
        rd::downside_deviation::EXPECTED_VALUES_BY_MAR_FULL,
        Measures::downside_deviation,
        15,
        "downside deviation",
    );
}

#[test]
fn test_downside_deviation_subset_matches_performance_analytics_output() {
    check_by_mar(
        rd::downside_deviation::EXPECTED_VALUES_BY_MAR_SUBSET,
        Measures::downside_deviation_subset,
        15,
        "downside deviation subset",
    );
}

#[test]
fn test_downside_frequency_matches_performance_analytics_output() {
    check_by_mar(rd::downside_frequency::EXPECTED_VALUES_BY_MAR, Measures::downside_frequency, 15, "downside frequency");
}

#[test]
fn test_downside_potential_matches_performance_analytics_output() {
    check_by_mar(rd::downside_potential::EXPECTED_VALUES_BY_MAR, Measures::downside_potential, 15, "downside potential");
}

// Sharpe family

#[test]
fn test_sharpe_ratio_matches_performance_analytics_output() {
    for &(rf, expected) in rd::sharpe_ratio::EXPECTED_VALUES_BY_RF_STDEV {
        let actual = run(cfg().rf(rf), Measures::sharpe_ratio);
        let places = if rf < 0.25 { 13 } else { 12 };
        assert_series(&actual, expected, Places(places), 0, &format!("Sharpe ratio (stdev) Rf {rf}"));
    }
}

fn check_by_p_rf(table: &[(f64, &[(f64, &[f64])])], method: Method, tol: Tol, label: &str) {
    for &(p, rf_pack) in table {
        for &(rf, expected) in rf_pack {
            let actual = run(cfg().rf(rf), |m| method(m, p));
            assert_series(&actual, expected, tol, 0, &format!("{label} conf {p} Rf {rf}"));
        }
    }
}

#[test]
fn test_sharpe_ratio_var_historical_matches_performance_analytics_output() {
    check_by_p_rf(
        rd::sharpe_ratio::EXPECTED_VALUES_BY_P_RF_VAR_HISTORICAL,
        Measures::sharpe_ratio_var_historical,
        Places(12),
        "Sharpe ratio (VaR historical)",
    );
}

#[test]
fn test_sharpe_ratio_var_gaussian_matches_performance_analytics_output() {
    check_by_p_rf(
        rd::sharpe_ratio::EXPECTED_VALUES_BY_P_RF_VAR_GAUSSIAN,
        Measures::sharpe_ratio_var_gaussian,
        Places(5),
        "Sharpe ratio (VaR Gaussian)",
    );
}

#[test]
fn test_sharpe_ratio_var_cornish_fisher_matches_performance_analytics_output() {
    check_by_p_rf(
        rd::sharpe_ratio::EXPECTED_VALUES_BY_P_RF_VAR_CORNISH_FISHER,
        Measures::sharpe_ratio_var_cornish_fisher,
        Places(6),
        "Sharpe ratio (VaR Cornish-Fisher)",
    );
}

#[test]
fn test_sharpe_ratio_es_historical_matches_performance_analytics_output() {
    for &(p, rf_pack) in rd::sharpe_ratio::EXPECTED_VALUES_BY_P_RF_ES_HISTORICAL {
        for &(rf, expected) in rf_pack {
            let actual = run(cfg().rf(rf), |m| m.sharpe_ratio_es_historical(p));
            assert_eq!(actual.len(), expected.len());
            for (i, (&a, &e)) in actual.iter().zip(expected).enumerate() {
                let prefix = format!("Sharpe ratio (ES historical) conf {p} Rf {rf} step {i}");
                if p == 0.9 && rf == 0.001 && i == 10 {
                    // The quantile is exactly the second-worst return.
                    // This implementation includes both tied-to-tail
                    // observations; the R reference includes only one.
                    let excess: Vec<f64> = BACON_PORTFOLIO_RETURNS[..11].iter().map(|r| r - rf).collect();
                    let excess_mean = fsum(&excess) / 11.0;
                    assert_float(a, excess_mean / 0.013, Places(13), &prefix);
                } else {
                    assert_float(a, e, Delta(1e-12), &prefix);
                }
            }
        }
    }
}

#[test]
fn test_sharpe_ratio_es_gaussian_matches_performance_analytics_output() {
    check_by_p_rf(
        rd::sharpe_ratio::EXPECTED_VALUES_BY_P_RF_ES_GAUSSIAN,
        Measures::sharpe_ratio_es_gaussian,
        Places(7),
        "Sharpe ratio (ES Gaussian)",
    );
}

#[test]
fn test_sharpe_ratio_es_cornish_fisher_matches_performance_analytics_output() {
    check_by_p_rf(
        rd::sharpe_ratio::EXPECTED_VALUES_BY_P_RF_ES_CORNISH_FISHER,
        Measures::sharpe_ratio_es_cornish_fisher,
        Places(5),
        "Sharpe ratio (ES Cornish-Fisher)",
    );
}

#[test]
fn test_downside_sharpe_ratio_matches_performance_analytics_output() {
    for &(rf, expected) in rd::downside_sharpe_ratio::EXPECTED_VALUES_BY_RF {
        let actual = run(cfg().rf(rf), Measures::downside_sharpe_ratio);
        assert_series(&actual, expected, Places(13), 0, &format!("downside Sharpe ratio Rf {rf}"));
    }
}

#[test]
fn test_adjusted_sharpe_ratio_matches_performance_analytics_output() {
    for &(rf, expected) in rd::adjusted_sharpe_ratio::EXPECTED_VALUES_BY_RF {
        let actual = run(cfg().rf(rf), Measures::adjusted_sharpe_ratio);
        let places = if rf < 0.2 { 12 } else { 11 };
        assert_series(&actual, expected, Places(places), 0, &format!("adjusted Sharpe ratio (stdev) Rf {rf}"));
    }
}

#[test]
fn test_adjusted_sharpe_ratio_skew_only_scale_and_translation_invariance() {
    // Metamorphic properties only (no reference data).
    let rf = 0.0042;
    let expected = run(cfg().rf(rf), Measures::adjusted_sharpe_ratio_skew_only);
    for scale in [4.2, -4.2] {
        for shift in [0.042, -0.042] {
            let transformed: Vec<f64> = BACON_PORTFOLIO_RETURNS.iter().map(|r| scale * r + shift).collect();
            let mut ratios = Measures::new(1.0, scale * rf + shift, 0.0, 0).unwrap();
            ratios.reset();
            for i in 0..BACON_PORTFOLIO_LEN {
                ratios.add_return(transformed[i], transformed[i]);
                let a = ratios.adjusted_sharpe_ratio_skew_only();
                assert_float(
                    if scale > 0.0 { a } else { -a },
                    expected[i],
                    Places(14),
                    &format!("ASR skew-only (scale {scale} shift {shift}) step {i}"),
                );
            }
        }
    }
}

fn check_psr(table: &[(f64, &[(f64, &[f64])])], method: Method, label: &str) {
    for &(ref_sr, rf_pack) in table {
        for &(rf, expected) in rf_pack {
            let actual = run(cfg().rf(rf), |m| method(m, ref_sr));
            assert_series(&actual, expected, Places(14), 0, &format!("{label} reference_sr {ref_sr} Rf {rf}"));
        }
    }
}

#[test]
fn test_probabilistic_sharpe_ratio_matches_performance_analytics_output() {
    check_psr(
        rd::probabilistic_sharpe_ratio::EXPECTED_VALUES_BY_REFSR_RF,
        Measures::probabilistic_sharpe_ratio,
        "probabilistic Sharpe ratio",
    );
}

#[test]
fn test_probabilistic_sharpe_ratio_full_matches_performance_analytics_output() {
    check_psr(
        rd::probabilistic_sharpe_ratio::EXPECTED_VALUES_BY_REFSR_RF_FULL,
        Measures::probabilistic_sharpe_ratio_full,
        "probabilistic Sharpe ratio (full)",
    );
}

#[test]
fn test_probabilistic_sharpe_ratio_symmetric_matches_performance_analytics_output() {
    check_psr(
        rd::probabilistic_sharpe_ratio::EXPECTED_VALUES_BY_REFSR_RF_SYMMETRIC,
        Measures::probabilistic_sharpe_ratio_symmetric,
        "probabilistic Sharpe ratio (symmetric)",
    );
}

#[test]
fn test_probabilistic_sharpe_ratio_gaussian_matches_performance_analytics_output() {
    check_psr(
        rd::probabilistic_sharpe_ratio::EXPECTED_VALUES_BY_REFSR_RF_GAUSSIAN,
        Measures::probabilistic_sharpe_ratio_gaussian,
        "probabilistic Sharpe ratio (Gaussian)",
    );
}
