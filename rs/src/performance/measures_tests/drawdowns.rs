//! Drawdown, drawdown-ratio and CDaR tests.

use super::*;
use crate::performance::reference_data as rd;

/// Yearly and daily runs (the measure does not depend on periodicity).
fn check_yearly_daily(expected: &[f64], f: fn(&Measures) -> f64, tol: Tol, label: &str) {
    let actual = run(cfg(), f);
    assert_series(&actual, expected, tol, 0, &format!("{label} (yearly)"));
    let actual = run(cfg().daily(), f);
    assert_series(&actual, expected, tol, 0, &format!("{label} (daily)"));
}

// Drawdown series

#[test]
fn test_drawdowns_cumulative_matches_performance_analytics_output() {
    let actual = run(cfg(), Measures::drawdowns_cumulative);
    for &(i, expected) in rd::drawdowns_cumulative::EXPECTED_VALUES_BY_INDEX {
        assert_series(&actual[i as usize], expected, Places(15), 0, &format!("drawdowns cumulative (i {i})"));
    }
}

#[test]
fn test_drawdowns_cumulative_matches_bacon_2023_output() {
    let actual = run(cfg().data(&BACON_2023_PORTFOLIO_RETURNS), Measures::drawdowns_cumulative);
    assert_series(
        &actual[actual.len() - 1],
        &BACON_2023_DRAWDOWN_FROM_PEAK,
        Places(4),
        0,
        "drawdowns cumulative (bacon 2023)",
    );
}

#[test]
fn test_min_drawdowns_cumulative_matches_performance_analytics_output() {
    let actual = run(cfg(), Measures::min_drawdowns_cumulative);
    assert_series(&actual, rd::min_drawdowns_cumulative::EXPECTED_VALUES, Places(15), 0, "min drawdowns cumulative");
}

#[test]
fn test_worst_drawdowns_cumulative_matches_performance_analytics_output() {
    let actual = run(cfg(), Measures::worst_drawdowns_cumulative);
    assert_series(
        &actual,
        rd::min_drawdowns_cumulative::EXPECTED_VALUES_INVERTED,
        Places(15),
        0,
        "worst drawdowns cumulative",
    );
}

#[test]
fn test_drawdowns_high_watermark_matches_performance_analytics_output() {
    let actual = run(cfg(), Measures::drawdowns_high_watermark);
    for &(i, expected) in rd::drawdowns_high_watermark::EXPECTED_VALUES_BY_INDEX {
        assert_series(&actual[i as usize], expected, Places(15), 0, &format!("drawdowns high_watermark (i {i})"));
    }
}

#[test]
fn test_drawdowns_continuous_runs_matches_bacon_2023_output() {
    // The book's four-decimal drawdown values are approximate; the core
    // run tracker has separate exact rolling-window tests.
    let actual = run(cfg().data(&BACON_2023_PORTFOLIO_RETURNS), |m| m.drawdowns_continuous_runs(None));
    let actual = &actual[actual.len() - 1];
    assert_series(
        actual,
        &BACON_2023_DRAWDOWN_CONTINUOUS_WITHOUT_ZEROES,
        Delta(0.002),
        0,
        "drawdowns continuous runs (bacon 2023)",
    );
}

#[test]
fn test_drawdowns_continuous_runs_max_runs() {
    let mut m = make_measures(0, 0.0, 0.0, YEARLY);
    assert!(m.drawdowns_continuous_runs(Some(3)).is_empty());
    add_bacon(&mut m, &BACON_2023_PORTFOLIO_RETURNS, &BACON_2023_PORTFOLIO_RETURNS);
    let all = m.drawdowns_continuous_runs(None);
    assert_eq!(m.drawdowns_continuous_runs(Some(0)), all);
    let mut worst = all.clone();
    worst.sort_by(|a, b| a.partial_cmp(b).unwrap());
    worst.truncate(3);
    assert_eq!(m.drawdowns_continuous_runs(Some(3)), worst);
    assert_eq!(m.drawdowns_continuous_runs(Some(100)).len(), all.len());
}

// Drawdown ratios

#[test]
fn test_calmar_ratio_matches_performance_analytics_output() {
    check_yearly_daily(rd::calmar_ratio::EXPECTED_VALUES, Measures::calmar_ratio, Places(13), "calmar ratio");
}

#[test]
fn test_sterling_ratio_matches_performance_analytics_output() {
    for &(excess, expected) in rd::sterling_ratio::EXPECTED_VALUES_BY_EXCESS {
        let actual = run(cfg(), |m| m.sterling_ratio(excess));
        assert_series(&actual, expected, Places(13), 0, &format!("sterling ratio (yearly, excess {excess})"));
    }
    for &(excess, expected) in rd::sterling_ratio::EXPECTED_VALUES_BY_EXCESS {
        let actual = run(cfg().daily(), |m| m.sterling_ratio(excess));
        assert_series(&actual, expected, Places(13), 0, &format!("sterling ratio (daily, excess {excess})"));
    }
}

#[test]
fn test_burke_ratio_matches_performance_analytics_output() {
    for &(rf, expected) in rd::burke_ratio::EXPECTED_VALUES_BY_RF {
        let actual = run(cfg().rf(rf), Measures::burke_ratio);
        assert_series(&actual, expected, Places(11), 0, &format!("burke ratio (yearly, Rf {rf})"));
    }
}

#[test]
fn test_burke_ratio_modified_matches_performance_analytics_output() {
    for &(rf, expected) in rd::burke_ratio::EXPECTED_VALUES_BY_RF_MODIFIED {
        let actual = run(cfg().rf(rf), Measures::burke_ratio_modified);
        assert_series(&actual, expected, Places(11), 0, &format!("burke ratio modified (Rf {rf})"));
    }
}

#[test]
fn test_pain_index_matches_performance_analytics_output() {
    // R's drawdown series differs slightly from the high-water-mark series
    // used here; the documented formula test checks the exact formula.
    check_yearly_daily(rd::pain_index::EXPECTED_VALUES, Measures::pain_index, Delta(0.00098), "pain index");
}

#[test]
fn test_pain_ratio_matches_performance_analytics_output() {
    for &(rf, expected) in rd::pain_ratio::EXPECTED_VALUES_BY_RF {
        let tol = Delta(if rf < 0.04 { 0.016 } else { 0.111 });
        let actual = run(cfg().rf(rf), Measures::pain_ratio);
        assert_series(&actual, expected, tol, 0, &format!("pain ratio (yearly, Rf {rf})"));
        let actual = run(cfg().daily().rf(compound(rf, 252.0)), Measures::pain_ratio);
        assert_series(&actual, expected, tol, 0, &format!("pain ratio (daily, Rf {rf})"));
    }
}

#[test]
fn test_ulcer_index_matches_performance_analytics_output() {
    check_yearly_daily(rd::ulcer_index::EXPECTED_VALUES, Measures::ulcer_index, Delta(0.00192), "ulcer index");
}

#[test]
fn test_martin_ratio_matches_performance_analytics_output() {
    for &(rf, expected) in rd::martin_ratio::EXPECTED_VALUES_BY_RF {
        let actual = run(cfg().rf(rf), Measures::martin_ratio);
        assert_series(&actual, expected, Delta(0.0630), 0, &format!("martin ratio (yearly, Rf {rf})"));
        let actual = run(cfg().daily().rf(compound(rf, 252.0)), Measures::martin_ratio);
        assert_series(&actual, expected, Delta(0.0630), 0, &format!("martin ratio (daily, Rf {rf})"));
    }
}

// Drawdown episodes

#[test]
fn test_drawdown_average_matches_performance_analytics_output() {
    check_yearly_daily(rd::drawdown_average::EXPECTED_VALUES, Measures::drawdown_average, Places(15), "drawdown average");
}

#[test]
fn test_drawdown_average_length_matches_performance_analytics_output() {
    check_yearly_daily(
        rd::drawdown_average_length::EXPECTED_VALUES,
        Measures::drawdown_average_length,
        Places(15),
        "drawdown average length",
    );
}

#[test]
fn test_drawdown_average_peak_to_trough_matches_performance_analytics_output() {
    check_yearly_daily(
        rd::drawdown_average_peak_to_trough::EXPECTED_VALUES,
        Measures::drawdown_average_peak_to_trough,
        Places(15),
        "drawdown average peak-to-trough",
    );
}

#[test]
fn test_drawdown_average_recovery_matches_performance_analytics_output() {
    check_yearly_daily(
        rd::drawdown_average_recovery::EXPECTED_VALUES,
        Measures::drawdown_average_recovery,
        Places(15),
        "drawdown average recovery",
    );
}

#[test]
fn test_drawdown_deviation_matches_performance_analytics_output() {
    check_yearly_daily(
        rd::drawdown_deviation::EXPECTED_VALUES,
        Measures::drawdown_deviation,
        Places(15),
        "drawdown deviation",
    );
}

// CDaR

#[test]
fn test_cdar_average_matches_performance_analytics_output() {
    // R and this implementation select the continuous drawdown tail
    // differently; the documented formula test checks the linear quantile.
    for &(p, expected) in rd::cdar::EXPECTED_VALUES_BY_P_AVERAGE_GEOMETRIC_INVERTED {
        let actual = run(cfg(), |m| m.cdar_average(p).unwrap());
        assert_series(&actual, expected, Delta(0.02938), 0, &format!("CDaR average geometric (yearly) p {p}"));
    }
    for &(p, expected) in rd::cdar::EXPECTED_VALUES_BY_P_AVERAGE_GEOMETRIC_INVERTED {
        let actual = run(cfg().daily(), |m| m.cdar_average(p).unwrap());
        assert_series(&actual, expected, Delta(0.02938), 0, &format!("CDaR average geometric (daily) p {p}"));
    }
}

#[test]
fn test_cdar_discrete_matches_performance_analytics_output() {
    for &(p, expected) in rd::cdar::EXPECTED_VALUES_BY_P_DISCRETE_GEOMETRIC_INVERTED {
        let actual = run(cfg(), |m| m.cdar_discrete(p).unwrap());
        assert_series(&actual, expected, Places(15), 0, &format!("CDaR discrete geometric (yearly) p {p}"));
    }
    for &(p, expected) in rd::cdar::EXPECTED_VALUES_BY_P_DISCRETE_GEOMETRIC_INVERTED {
        let actual = run(cfg().daily(), |m| m.cdar_discrete(p).unwrap());
        assert_series(&actual, expected, Places(15), 0, &format!("CDaR discrete geometric (daily) p {p}"));
    }
}

#[test]
fn test_cdar_rejects_invalid_confidence() {
    let mut m = make_measures(0, 0.0, 0.0, YEARLY);
    add_bacon(&mut m, &BACON_PORTFOLIO_RETURNS, &BACON_BENCHMARK_RETURNS);
    let msg = Err("confidence must be between 0 and 1".to_string());
    for confidence in [0.0, 1.0, f64::NAN] {
        assert_eq!(m.cdar_average(confidence), msg);
        assert_eq!(m.cdar_discrete(confidence), msg);
        assert_eq!(m.cdar_beta(confidence), msg);
        assert_eq!(m.cdar_alpha(confidence), msg);
    }
}

#[test]
fn test_cdar_beta_discrete_tail_selection() {
    let mut m = make_measures(0, 0.0, 0.0, YEARLY);
    for (portfolio, benchmark) in [(0.1, 0.1), (-0.05, -0.1), (0.2, 0.2), (-0.1, -0.2), (0.3, 0.3), (-0.15, -0.3)] {
        m.add_return(portfolio, benchmark);
    }
    // At 50% confidence, two of three episodes are selected. The
    // denominator is the second-worst depth, -0.2.
    assert_float(m.cdar_beta(0.5).unwrap(), (-0.15 - 0.1) / (2.0 * -0.2), Places(14), "cdar beta 0.5");
    assert_float(m.cdar_beta(0.8).unwrap(), -0.15 / -0.3, Places(14), "cdar beta 0.8");
}

#[test]
fn test_cdar_beta_matches_performance_analytics_output() {
    for &(p, expected) in rd::cdar_beta::EXPECTED_VALUES_BY_P_GEOMETRIC {
        let actual = run(cfg(), |m| m.cdar_beta(p).unwrap());
        assert_series(&actual, expected, Places(13), 0, &format!("CDaR beta geometric (yearly) p {p}"));
    }
    for &(p, expected) in rd::cdar_beta::EXPECTED_VALUES_BY_P_GEOMETRIC {
        let actual = run(cfg().daily(), |m| m.cdar_beta(p).unwrap());
        assert_series(&actual, expected, Places(13), 0, &format!("CDaR beta geometric (daily) p {p}"));
    }
}

#[test]
fn test_cdar_beta_mathematical_properties() {
    // With one selected episode, identical portfolio and benchmark returns
    // give the same numerator and denominator.
    let mut measures = make_measures(0, 0.0, 0.0, YEARLY);
    for ret in [0.05, -0.1] {
        measures.add_return(ret, ret);
    }
    assert_float(measures.cdar_beta(0.95).unwrap(), 1.0, Places(15), "CDaR beta (one episode) identity");

    let mut measures = make_measures(0, 0.0, 0.0, YEARLY);
    add_bacon(&mut measures, &BACON_PORTFOLIO_RETURNS, &BACON_PORTFOLIO_RETURNS);
    assert_float(measures.cdar_beta(0.95).unwrap(), 1.0, Places(14), "CDaR beta (Bacon) identity");

    // No drawdowns
    let mut transformed = make_measures(0, 0.0, 0.0, YEARLY);
    add_bacon(&mut transformed, &[0.01; 24], &[0.02; 24]);
    assert_float(transformed.cdar_beta(0.95).unwrap(), f64::NAN, Places(15), "CDaR beta (geometric) no drawdowns");

    // Zero returns (portfolio = 0)
    let mut transformed = make_measures(0, 0.0, 0.0, YEARLY);
    add_bacon(&mut transformed, &[0.0; 24], &BACON_BENCHMARK_RETURNS);
    assert_float(transformed.cdar_beta(0.95).unwrap(), 0.0, Places(15), "CDaR beta (geometric) zero returns");

    for confidence in [0.0, 1.0] {
        assert!(measures.cdar_beta(confidence).is_err());
    }
}

#[test]
fn test_cdar_alpha_matches_performance_analytics_output() {
    // PerformanceAnalytics hardcodes a period of 12 when annualizing the
    // means; this implementation uses periods_per_annum, so monthly data
    // is used to match the reference.
    for &(p, expected) in rd::cdar_alpha::EXPECTED_VALUES_BY_P_GEOMETRIC {
        let actual = run(cfg().monthly(), |m| m.cdar_alpha(p).unwrap());
        assert_series(&actual, expected, Places(14), 0, &format!("CDaR alpha geometric (monthly) p {p}"));
    }
}

#[test]
fn test_cdar_alpha_mathematical_properties() {
    // Identity (portfolio == benchmark).
    let mut measures = make_measures(0, 0.0, 0.0, MONTHLY);
    add_bacon(&mut measures, &BACON_PORTFOLIO_RETURNS, &BACON_PORTFOLIO_RETURNS);
    assert_float(measures.cdar_alpha(0.95).unwrap(), 0.0, Places(14), "CDaR alpha (geometric) identity");

    // No drawdowns
    let mut transformed = make_measures(0, 0.0, 0.0, YEARLY);
    add_bacon(&mut transformed, &[0.01; 24], &[0.02; 24]);
    assert_float(transformed.cdar_alpha(0.95).unwrap(), f64::NAN, Places(15), "CDaR alpha (geometric) no drawdowns");

    // Zero returns (portfolio = 0)
    let mut transformed = make_measures(0, 0.0, 0.0, YEARLY);
    add_bacon(&mut transformed, &[0.0; 24], &BACON_BENCHMARK_RETURNS);
    assert_float(transformed.cdar_alpha(0.95).unwrap(), 0.0, Places(15), "CDaR alpha (geometric) zero returns");
}

#[test]
fn test_reward_to_conditional_drawdown_definition() {
    // Geometric mean return divided by the mean magnitude of the worst
    // max(1, int(n * (1 - confidence))) drawdowns.
    for confidence in [0.8, 0.95] {
        let mut m = make_measures(0, 0.0, 0.0, YEARLY);
        for i in 0..BACON_PORTFOLIO_LEN {
            m.add_return(BACON_PORTFOLIO_RETURNS[i], BACON_BENCHMARK_RETURNS[i]);
            let mut dd = m.drawdowns_high_watermark();
            dd.sort_by(|a, b| a.partial_cmp(b).unwrap());
            let n_tail = ((dd.len() as f64 * (1.0 - confidence)) as usize).max(1);
            let tail = &dd[..n_tail];
            let cdar = -tail.iter().sum::<f64>() / tail.len() as f64;
            let expected = if cdar != 0.0 { m.geometric_mean_return() / cdar } else { f64::NAN };
            assert_float(
                m.reward_to_conditional_drawdown(confidence),
                expected,
                Places(15),
                &format!("confidence {confidence} step {i}"),
            );
        }
    }
}
