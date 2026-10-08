//! Assertion helpers, edge cases, growth, moments and normality tests.

use super::*;
use crate::performance::core::test_support::{gauss, prod, Rng};
use crate::performance::measures::is_normal_from_jb;
use crate::performance::reference_data as rd;
use crate::streaming_kbn::RawMomentsKleinKbn;

// TestSeriesAssertions

#[test]
fn test_series_assertions_rejects_truncated_series() {
    assert!(check_series(&[1.0], &[1.0, 2.0], Places(15), 0).is_err());
    assert!(check_series(&[1.0, 2.0], &[1.0], Places(15), 0).is_err());
}

#[test]
fn test_series_assertions_accepts_generator() {
    let actual: Vec<f64> = [1.0, 2.0].iter().copied().collect();
    assert_series(&actual, &[1.0, 2.0], Places(15), 0, "generator");
}

#[test]
fn test_series_assertions_relative_tolerance_for_large_reference_values() {
    assert_series(&[1e12 + 0.1], &[1e12], Rel(1e-12, 0.0), 0, "relative");
    assert!(check_series(&[1e12 + 2.0], &[1e12], Rel(1e-12, 0.0), 0).is_err());
}

// TestEdgeCases

#[test]
fn test_edge_cases_empty() {
    // No property or method fails before the first return.
    let m = make_measures(0, 0.0, 0.0, YEARLY);
    for (_name, eval) in public_measures() {
        eval(&m);
    }
}

#[test]
fn test_edge_cases_single_return() {
    let mut m = make_measures(0, 0.0, 0.0, YEARLY);
    m.add_return(0.01, 0.02);
    for (_name, eval) in public_measures() {
        eval(&m);
    }
    assert!(m.sharpe_ratio().is_nan());
    assert_float(m.cumulative_geometric_return(), 0.01, Places(15), "cumulative geometric return");
}

#[test]
fn test_edge_cases_first_negative_return_is_a_drawdown() {
    // As in PerformanceAnalytics Drawdowns(), the high-water mark starts
    // at the initial equity 1.
    let mut m = make_measures(0, 0.0, 0.0, YEARLY);
    m.add_return(-0.05, -0.02);
    m.add_return(0.02, 0.01);
    assert_series(&m.drawdowns_high_watermark(), &[-0.05, -0.031], Places(15), 0, "hwm");
    assert_series(&m.drawdowns_cumulative(), &[-0.05, -0.031], Places(15), 0, "cumulative");
    assert_float(m.worst_drawdowns_cumulative(), 0.05, Places(15), "worst");
    assert_float(m.pain_index(), (0.05 + 0.031) / 2.0, Places(15), "pain index");
    assert_float(m.drawdown_average(), 0.05, Places(15), "drawdown average");
}

#[test]
fn test_edge_cases_long_daily_series() {
    // Every measure works with more observations than periods per annum.
    let mut rng = Rng::new(1);
    let mut m = Measures::new(252.0, 0.0, 0.0, 0).unwrap();
    for _ in 0..300 {
        let r = gauss(&mut rng, 0.0005, 0.01);
        let b = gauss(&mut rng, 0.0004, 0.01);
        m.add_return(r, b);
    }
    for (_name, eval) in public_measures() {
        eval(&m);
    }
    assert!(m.autocorrelation_penalty().is_finite());
}

// TestAutocorrelationPenalty (metamorphic properties only)

#[test]
fn test_autocorrelation_penalty_metamorphic_properties() {
    let n = BACON_2023_PORTFOLIO_LEN;
    // Constant returns
    let returns = vec![0.01; n];
    let actual = run(cfg().daily().data(&returns), Measures::autocorrelation_penalty);
    assert_float(actual[n - 1], 1.0, Places(15), "autocorrelation penalty (constant)");

    // Too few observations
    assert_float(actual[0], 1.0, Places(15), "autocorrelation penalty (len=0)");
    assert_float(actual[1], 1.0, Places(15), "autocorrelation penalty (len=1)");

    // Positive autocorrelation
    let returns: Vec<f64> = (0..n).map(|i| 0.01 * i as f64).collect();
    let actual = run(cfg().daily().data(&returns), Measures::autocorrelation_penalty);
    assert_float(actual[n - 1], 2.722393904531189, Places(15), "autocorrelation penalty (positive)");

    // Negative autocorrelation
    let returns: Vec<f64> = (0..n).map(|i| if i % 2 == 0 { 0.01 } else { -0.01 }).collect();
    let actual = run(cfg().daily().data(&returns), Measures::autocorrelation_penalty);
    assert_float(actual[n - 1], 0.16903085094570597, Places(15), "autocorrelation penalty (negative)");

    // Scale and translation invariance
    let expected = run(cfg().daily(), Measures::autocorrelation_penalty);
    for scale in [4.2, -4.2] {
        for shift in [0.042, -0.042] {
            let transformed: Vec<f64> = BACON_PORTFOLIO_RETURNS.iter().map(|r| scale * r + shift).collect();
            let actual = run(cfg().daily().data(&transformed), Measures::autocorrelation_penalty);
            assert_series(
                &actual,
                &expected,
                Places(15),
                0,
                &format!("autocorrelation penalty (transform) scale {scale} shift {shift}"),
            );
        }
    }
}

// TestCumulativeGeometricReturn

#[test]
fn test_cumulative_geometric_return_matches_performance_analytics_output() {
    let expected = rd::cumulative_geometric_return::EXPECTED_VALUES;
    for (ppa, label) in [(YEARLY, "yearly"), (MONTHLY, "monthly"), (DAILY, "daily")] {
        let c = Cfg { ppa, ..cfg() };
        let actual = run(c, Measures::cumulative_geometric_return);
        assert_series(&actual, expected, Places(14), 0, &format!("cumulative geometric return ({label})"));
    }
}

// TestGeometricMeanReturn

#[test]
fn test_geometric_mean_return_matches_performance_analytics_output() {
    let expected = rd::geometric_mean_return::EXPECTED_VALUES_GEOMETRIC;
    for (ppa, label) in [(YEARLY, "yearly"), (MONTHLY, "monthly"), (DAILY, "daily")] {
        let c = Cfg { ppa, ..cfg() };
        let actual = run(c, Measures::geometric_mean_return);
        assert_series(&actual, expected, Places(15), 0, &format!("geometric mean return ({label})"));
    }
}

// TestCompoundAnnualGrowthRate

#[test]
fn test_compound_annual_growth_rate_annualized_return_definition() {
    fn calculate(m: &Measures, periods_per_annum: f64) -> f64 {
        let growth = prod(m.returns.iter().map(|r| 1.0 + r));
        growth.powf(periods_per_annum / m.returns.len() as f64) - 1.0
    }
    for (ppa, places, label) in [(YEARLY, 15, "yearly"), (MONTHLY, 14, "monthly"), (DAILY, 11, "daily")] {
        let c = Cfg { ppa, ..cfg() };
        let expected = run(c, |m| calculate(m, ppa));
        let actual = run(c, Measures::compound_annual_growth_rate);
        assert_series(&actual, &expected, Places(places), 0, &format!("compound annual growth rate ({label})"));
    }
}

// TestSkewness

#[test]
fn test_skewness_matches_performance_analytics_output() {
    // Hardcoded bias=True in RawMomentsKleinKbn: only `moment` maps to
    // `skewness`.
    for &(method, expected) in rd::skewness::EXPECTED_VALUES_BY_METHOD {
        let f: fn(&Measures) -> f64 = match method {
            "moment" => Measures::skewness_moment,
            "fisher" => Measures::skewness_fisher,
            "sample" => Measures::skewness_sample,
            _ => panic!("unknown method {method}"),
        };
        let actual = run(cfg(), f);
        assert_series(&actual, expected, Places(14), 0, &format!("skewness_{method}"));
        if method == "moment" {
            let actual = run(cfg(), Measures::skewness);
            assert_series(&actual, expected, Places(14), 0, "skewness");
        }
    }
}

#[test]
fn test_skewness_raw_moments_klein_kbn() {
    // moment: bias=true; fisher: bias=false; sample: skewness_sample.
    for &(method, expected) in rd::skewness::EXPECTED_VALUES_BY_METHOD {
        let bias = method != "fisher";
        let mut kbn = RawMomentsKleinKbn::new(1, bias, true);
        for i in 0..BACON_PORTFOLIO_LEN {
            kbn.update(BACON_PORTFOLIO_RETURNS[i]);
            match method {
                "moment" => assert_float(
                    kbn.skewness_moment(),
                    kbn.skewness(),
                    Places(15),
                    &format!("step {i} skewness_{method} vs. skewness"),
                ),
                "fisher" => assert_float(
                    kbn.skewness_fisher(),
                    kbn.skewness(),
                    Places(15),
                    &format!("step {i} skewness_{method} vs. skewness"),
                ),
                _ => assert_float(
                    kbn.skewness_sample(),
                    expected[i],
                    Places(14),
                    &format!("step {i} skewness_{method}"),
                ),
            }
        }
    }
}

// TestKurtosis

#[test]
fn test_kurtosis_matches_performance_analytics_output() {
    for &(method, expected) in rd::kurtosis::EXPECTED_VALUES_BY_METHOD {
        let f: fn(&Measures) -> f64 = match method {
            "excess" => Measures::kurtosis_excess,
            "moment" => Measures::kurtosis_moment,
            "sample_corrected" => Measures::kurtosis_sample_corrected,
            "sample_excess" => Measures::kurtosis_sample_excess,
            _ => panic!("unknown method {method}"),
        };
        let actual = run(cfg(), f);
        assert_series(&actual, expected, Places(13), 0, &format!("kurtosis_{method}"));
        if method == "excess" {
            let actual = run(cfg(), Measures::kurtosis);
            assert_series(&actual, expected, Places(13), 0, "kurtosis");
        }
    }
}

#[test]
fn test_kurtosis_raw_moments_klein_kbn() {
    for &(method, expected) in rd::kurtosis::EXPECTED_VALUES_BY_METHOD {
        let (bias, fisher) = match method {
            "moment" => (true, false),
            "sample_corrected" => (false, false),
            "sample_excess" => (false, true),
            _ => (true, true),
        };
        let mut kbn = RawMomentsKleinKbn::new(1, bias, fisher);
        for i in 0..BACON_PORTFOLIO_LEN {
            kbn.update(BACON_PORTFOLIO_RETURNS[i]);
            let actual = match method {
                "excess" => kbn.kurtosis_excess(),
                "moment" => kbn.kurtosis_moment(),
                "sample_corrected" => kbn.kurtosis_sample_corrected(),
                _ => kbn.kurtosis_sample_excess(),
            };
            let dispatched = if method == "sample_corrected" { kbn.kurtosis_sample() } else { actual };
            assert_float(dispatched, kbn.kurtosis(), Places(15), &format!("step {i} kurtosis_{method} / kurtosis"));
            assert_float(actual, expected[i], Places(13), &format!("step {i} kurtosis_{method}"));
        }
    }
}

// TestSkewnessKurtosisRatio

#[test]
fn test_skewness_kurtosis_ratio_matches_performance_analytics_output() {
    let actual = run(cfg(), Measures::skewness_kurtosis_ratio);
    assert_series(&actual, rd::skewness_kurtosis_ratio::EXPECTED_VALUES, Places(14), 0, "skewness-kurtosis ratio");
}

// TestJarqueBeraNrmalityTestStatistic

#[test]
fn test_jarque_bera_matches_bacon3_output() {
    let actual = run(cfg().data(&BACON_2023_PORTFOLIO_RETURNS), Measures::jarque_bera_normality_test_statistic);
    // Chapter 5, exhibit 5.4
    assert_float(actual[BACON_2023_PORTFOLIO_LEN - 1], 0.34, Places(2), "Jarque-Bera normality (bacon3)");
}

#[test]
fn test_jarque_bera_matches_scipy_output() {
    let actual = run(cfg(), Measures::jarque_bera_normality_test_statistic);
    assert_series(
        &actual,
        rd::jarque_bera_normality_test_statistic::EXPECTED_VALUES,
        Places(14),
        0,
        "Jarque-Bera normality (scipy)",
    );
}

// TestIsNormalDistribution: Python patches the JB property with a
// PropertyMock; here the decision rule is tested directly.

#[test]
fn test_is_normal_distribution_mocked_jb() {
    // Since 5.0 < 5.991..., normality is accepted.
    assert!(is_normal_from_jb(5.0, 0.95).unwrap(), "normality accepted");
    assert!(!is_normal_from_jb(10.0, 0.95).unwrap(), "normality rejected");
    assert!(!is_normal_from_jb(f64::NAN, 0.95).unwrap(), "NaN statistic");
    assert_eq!(is_normal_from_jb(0.0, 1.0), Err("confidence must be between 0 and 1".to_string()));
    assert_eq!(is_normal_from_jb(0.0, 0.0), Err("confidence must be between 0 and 1".to_string()));
    assert!(is_normal_from_jb(8.0, 0.99).unwrap(), "custom confidence 8");
    assert!(!is_normal_from_jb(10.0, 0.99).unwrap(), "custom confidence 10");

    // The method delegates to the rule.
    let ratios = Measures::new(1.0, 0.0, 0.0, 0).unwrap();
    assert_eq!(ratios.is_normal_distribution(1.0), Ok(false), "NaN statistic checked first");
    let mut ratios = ratios;
    add_bacon(&mut ratios, &BACON_PORTFOLIO_RETURNS, &BACON_BENCHMARK_RETURNS);
    let jb = ratios.jarque_bera_normality_test_statistic();
    for confidence in [0.9, 0.95, 0.99] {
        assert_eq!(ratios.is_normal_distribution(confidence), is_normal_from_jb(jb, confidence));
    }
    assert!(ratios.is_normal_distribution(1.0).is_err());
    assert!(ratios.is_normal_distribution(0.0).is_err());
}

#[test]
fn test_constructor_rejects_non_positive_periods_per_annum() {
    for ppa in [0.0, -1.0] {
        assert_eq!(Measures::new(ppa, 0.0, 0.0, 0).unwrap_err(), "periods_per_annum must be positive");
    }
    let m = Measures::new(12.0, 0.05, 0.03, 0).unwrap();
    assert_eq!(m.periods_per_annum(), 12.0);
    assert_eq!(m.risk_free_rate(), 1.05f64.powf(1.0 / 12.0) - 1.0);
    assert_eq!(m.target_return(), 1.03f64.powf(1.0 / 12.0) - 1.0);
    let m = Measures::new(1.0, 0.05, 0.03, 0).unwrap();
    assert_eq!(m.risk_free_rate(), 0.05);
    assert_eq!(m.target_return(), 0.03);
}
