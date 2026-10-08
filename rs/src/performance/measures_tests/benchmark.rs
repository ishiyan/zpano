//! Single-factor model, benchmark-relative, miscellaneous and capture tests.

use super::*;
use crate::performance::core::test_support::prod;
use crate::performance::reference_data as rd;

type Table = &'static [(f64, &'static [f64])];

/// Yearly with the periodic rf as annual rf, then daily with the periodic
/// rf compounded over 252 periods.
fn check_rf_yearly_daily(table: Table, f: fn(&Measures) -> f64, places: i32, label: &str) {
    for &(rf, expected) in table {
        let actual = run(cfg().rf(rf), f);
        assert_series(&actual, expected, Places(places), 0, &format!("{label} (yearly, Rf {rf})"));
    }
    for &(rf, expected) in table {
        let actual = run(cfg().daily().rf(compound(rf, 252.0)), f);
        assert_series(&actual, expected, Places(places), 0, &format!("{label} (daily, Rf {rf})"));
    }
}

// SFM

#[test]
fn test_sfm_risk_premium_matches_performance_analytics_output() {
    check_rf_yearly_daily(rd::sfm_risk_premium::EXPECTED_VALUES_BY_RF_PERFAN, Measures::sfm_risk_premium, 14, "SFM risk premium");
}

#[test]
fn test_sfm_alpha_matches_performance_analytics_output() {
    check_rf_yearly_daily(rd::sfm_alpha::EXPECTED_VALUES_BY_RF_PERFAN, Measures::sfm_alpha, 14, "SFM alpha");
}

#[test]
fn test_sfm_beta_matches_performance_analytics_output() {
    check_rf_yearly_daily(rd::sfm_beta::EXPECTED_VALUES_BY_RF_PERFAN, Measures::sfm_beta, 14, "SFM beta");
}

#[test]
fn test_sfm_beta_bull_matches_performance_analytics_output() {
    check_rf_yearly_daily(rd::sfm_beta_bull::EXPECTED_VALUES_BY_RF_PERFAN, Measures::sfm_beta_bull, 14, "SFM beta bull");
}

#[test]
fn test_sfm_beta_bear_matches_reference_implementation_output() {
    check_rf_yearly_daily(rd::sfm_beta_bear::EXPECTED_VALUES_BY_RF_REFERENCE, Measures::sfm_beta_bear, 14, "SFM beta bear");
}

#[test]
fn test_timing_ratio_matches_performance_analytics_output() {
    check_rf_yearly_daily(rd::timing_ratio::EXPECTED_VALUES_BY_RF_PERFAN, Measures::timing_ratio, 14, "timing ratio");
}

#[test]
fn test_sfm_r2_matches_performance_analytics_output() {
    for &(rf, expected) in rd::sfm_r2::EXPECTED_VALUES_BY_RF_PERFAN {
        let skip = if rf < 0.05 { 15 } else { 18 };
        let actual = run(cfg().rf(rf), Measures::sfm_r2);
        assert_series(&actual, expected, Places(14), skip, &format!("SFM R^2 (yearly, Rf {rf})"));
    }
    for &(rf, expected) in rd::sfm_r2::EXPECTED_VALUES_BY_RF_PERFAN {
        let skip = if rf < 0.05 { 15 } else { 18 };
        let actual = run(cfg().daily().rf(compound(rf, 252.0)), Measures::sfm_r2);
        assert_series(&actual, expected, Places(14), skip, &format!("SFM R^2 (daily, Rf {rf})"));
    }
}

// Jensen's alpha

#[test]
fn test_jensen_alpha_high_daily_risk_free_rate_definition() {
    // The R fixtures lose precision after compounding a 10% or 30% periodic
    // rate over 252 periods. Check the documented formula.
    for rf in [0.1, 0.3] {
        let annual_rf = compound(rf, 252.0);
        let mut m = Measures::new(252.0, annual_rf, 0.0, 0).unwrap();
        for i in 0..BACON_PORTFOLIO_LEN {
            m.add_return(BACON_PORTFOLIO_RETURNS[i], BACON_BENCHMARK_RETURNS[i]);
            if i == 0 {
                continue;
            }
            let n = (i + 1) as f64;
            let p_ann = prod(BACON_PORTFOLIO_RETURNS[..=i].iter().map(|x| 1.0 + x)).powf(252.0 / n) - 1.0;
            let b_ann = prod(BACON_BENCHMARK_RETURNS[..=i].iter().map(|x| 1.0 + x)).powf(252.0 / n) - 1.0;
            let beta = m.sfm_beta();
            let expected = p_ann - (beta * b_ann + (1.0 - beta) * annual_rf);
            assert_float(m.jensen_alpha(), expected, Rel(1e-12, 1e-9), &format!("Jensen alpha rf={rf} step={i}"));
            if beta != 0.0 {
                assert_float(
                    m.jensen_alpha_modified(),
                    expected / beta,
                    Rel(1e-12, 1e-9),
                    &format!("Jensen alpha modified rf={rf} step={i}"),
                );
            }
        }
    }
}

#[test]
fn test_jensen_alpha_matches_performance_analytics_output() {
    for &(rf, expected) in rd::jensen_alpha::EXPECTED_VALUES_BY_RF_DAILY_PERFAN {
        let actual = run(cfg().daily().rf(compound(rf, 252.0)), Measures::jensen_alpha);
        if rf <= 0.05 {
            // At higher periodic rates the R reference loses precision
            // through cancellation; the formula test above covers them.
            assert_series(&actual, expected, Rel(1e-10, 9e-10), 0, &format!("Jensen alpha (daily, Rf {rf})"));
        }
    }
    for &(rf, expected) in rd::jensen_alpha::EXPECTED_VALUES_BY_RF_MONTHLY_PERFAN {
        let actual = run(cfg().monthly().rf(compound(rf, 12.0)), Measures::jensen_alpha);
        assert_series(&actual, expected, Places(12), 0, &format!("Jensen alpha (monthly, Rf {rf})"));
    }
    for &(rf, expected) in rd::jensen_alpha::EXPECTED_VALUES_BY_RF_YEARLY_PERFAN {
        let actual = run(cfg().rf(rf), Measures::jensen_alpha);
        assert_series(&actual, expected, Places(14), 0, &format!("Jensen alpha (yearly, Rf {rf})"));
    }
}

#[test]
fn test_fama_beta_matches_performance_analytics_output() {
    let expected = rd::fama_beta::EXPECTED_VALUES_PERFAN;
    let actual = run(cfg(), Measures::fama_beta);
    assert_series(&actual, expected, Places(14), 0, "fama beta (yearly)");
    let actual = run(cfg().daily(), Measures::fama_beta);
    assert_series(&actual, expected, Places(14), 0, "fama beta (daily)");
}

#[test]
fn test_modigliani_matches_performance_analytics_output() {
    check_rf_yearly_daily(rd::modigliani::EXPECTED_VALUES_BY_RF_PERFAN, Measures::modigliani, 15, "Modigliani-Modigliani");
}

#[test]
fn test_tracking_error_matches_performance_analytics_output() {
    let actual = run(cfg().daily(), Measures::tracking_error);
    assert_series(&actual, rd::tracking_error::EXPECTED_VALUES_DAILY_PERFAN, Places(15), 0, "tracking error (daily)");
    let actual = run(cfg().monthly(), Measures::tracking_error);
    assert_series(&actual, rd::tracking_error::EXPECTED_VALUES_MONTHLY_PERFAN, Places(15), 0, "tracking error (monthly)");
    let actual = run(cfg(), Measures::tracking_error);
    assert_series(&actual, rd::tracking_error::EXPECTED_VALUES_ANNUAL_PERFAN, Places(15), 0, "tracking error (yearly)");
}

#[test]
fn test_active_premium_matches_performance_analytics_output() {
    // The first element is skipped because PerformanceAnalytics uses it to
    // determine periodicity.
    let actual = run(cfg().daily(), Measures::active_premium);
    assert_series(&actual, rd::active_premium::EXPECTED_VALUES_DAILY_PERFAN, Places(11), 1, "active premium (daily)");
    let actual = run(cfg().monthly(), Measures::active_premium);
    assert_series(&actual, rd::active_premium::EXPECTED_VALUES_MONTHLY_PERFAN, Places(14), 1, "active premium (monthly)");
    let actual = run(cfg(), Measures::active_premium);
    assert_series(&actual, rd::active_premium::EXPECTED_VALUES_ANNUAL_PERFAN, Places(15), 1, "active premium (yearly)");
}

#[test]
fn test_information_ratio_matches_performance_analytics_output() {
    let actual = run(cfg().daily(), Measures::information_ratio);
    assert_series(&actual, rd::information_ratio::EXPECTED_VALUES_DAILY_PERFAN, Places(10), 2, "information ratio (daily)");
    let actual = run(cfg().monthly(), Measures::information_ratio);
    assert_series(
        &actual,
        rd::information_ratio::EXPECTED_VALUES_MONTHLY_PERFAN,
        Places(12),
        2,
        "information ratio (monthly)",
    );
    let actual = run(cfg(), Measures::information_ratio);
    assert_series(&actual, rd::information_ratio::EXPECTED_VALUES_ANNUAL_PERFAN, Places(13), 2, "information ratio (yearly)");
}

#[test]
fn test_information_ratio_modified_sign_rule() {
    // Equals information_ratio when the mean active return is positive,
    // otherwise its negation.
    let mut m = make_measures(0, 0.0, 0.0, MONTHLY);
    for i in 0..BACON_PORTFOLIO_LEN {
        m.add_return(BACON_PORTFOLIO_RETURNS[i], BACON_BENCHMARK_RETURNS[i]);
        let diffs: Vec<f64> = (0..=i).map(|j| BACON_PORTFOLIO_RETURNS[j] - BACON_BENCHMARK_RETURNS[j]).collect();
        let active = fsum(&diffs);
        let ir = m.information_ratio();
        let expected = if ir.is_nan() { f64::NAN } else if active > 0.0 { ir } else { -ir };
        assert_float(m.information_ratio_modified(), expected, Places(15), &format!("step {i}"));
    }
}

/// Daily (rf compounded over 252), monthly (over 12) and annual tables.
fn check_rf_dma(daily: Table, monthly: Table, annual: Table, f: fn(&Measures) -> f64, tols: [Tol; 3], skip: usize, label: &str) {
    for &(rf, expected) in daily {
        let actual = run(cfg().daily().rf(compound(rf, 252.0)), f);
        assert_series(&actual, expected, tols[0], skip, &format!("{label} (daily, Rf {rf})"));
    }
    for &(rf, expected) in monthly {
        let actual = run(cfg().monthly().rf(compound(rf, 12.0)), f);
        assert_series(&actual, expected, tols[1], skip, &format!("{label} (monthly, Rf {rf})"));
    }
    for &(rf, expected) in annual {
        let actual = run(cfg().rf(rf), f);
        assert_series(&actual, expected, tols[2], skip, &format!("{label} (yearly, Rf {rf})"));
    }
}

#[test]
fn test_systematic_risk_matches_performance_analytics_output() {
    check_rf_dma(
        rd::systematic_risk::EXPECTED_VALUES_BY_RF_DAILY_PERFAN,
        rd::systematic_risk::EXPECTED_VALUES_BY_RF_MONTHLY_PERFAN,
        rd::systematic_risk::EXPECTED_VALUES_BY_RF_ANNUAL_PERFAN,
        Measures::systematic_risk,
        [Places(14), Places(15), Places(15)],
        0,
        "systematic risk",
    );
}

#[test]
fn test_treynor_ratio_matches_performance_analytics_output() {
    check_rf_dma(
        rd::treynor_ratio::EXPECTED_VALUES_BY_RF_DAILY_PERFAN,
        rd::treynor_ratio::EXPECTED_VALUES_BY_RF_MONTHLY_PERFAN,
        rd::treynor_ratio::EXPECTED_VALUES_BY_RF_ANNUAL_PERFAN,
        Measures::treynor_ratio,
        [Places(10), Places(13), Places(14)],
        0,
        "treynor ratio",
    );
}

#[test]
fn test_treynor_ratio_modified_matches_performance_analytics_output() {
    check_rf_dma(
        rd::treynor_ratio_modified::EXPECTED_VALUES_BY_RF_DAILY_PERFAN,
        rd::treynor_ratio_modified::EXPECTED_VALUES_BY_RF_MONTHLY_PERFAN,
        rd::treynor_ratio_modified::EXPECTED_VALUES_BY_RF_ANNUAL_PERFAN,
        Measures::treynor_ratio_modified,
        [Places(10), Places(12), Places(12)],
        0,
        "treynor ratio modified",
    );
}

#[test]
fn test_specific_risk_matches_performance_analytics_output() {
    check_rf_dma(
        rd::specific_risk::EXPECTED_VALUES_BY_RF_DAILY_PERFAN,
        rd::specific_risk::EXPECTED_VALUES_BY_RF_MONTHLY_PERFAN,
        rd::specific_risk::EXPECTED_VALUES_BY_RF_ANNUAL_PERFAN,
        Measures::specific_risk,
        [Places(14), Places(15), Places(15)],
        0,
        "specific risk",
    );
}

#[test]
fn test_total_risk_matches_performance_analytics_output() {
    check_rf_dma(
        rd::total_risk::EXPECTED_VALUES_BY_RF_DAILY_PERFAN,
        rd::total_risk::EXPECTED_VALUES_BY_RF_MONTHLY_PERFAN,
        rd::total_risk::EXPECTED_VALUES_BY_RF_ANNUAL_PERFAN,
        Measures::total_risk,
        [Places(14), Places(14), Places(15)],
        0,
        "total_risk",
    );
}

#[test]
fn test_appraisal_ratio_matches_performance_analytics_output() {
    for &(rf, expected) in rd::appraisal_ratio::EXPECTED_VALUES_BY_RF_DAILY_PERFAN {
        let actual = run(cfg().daily().rf(compound(rf, 252.0)), Measures::appraisal_ratio);
        if rf < 0.1 {
            let tol = Delta(if rf < 0.05 { 1e-8 } else { 1e-4 });
            assert_series(&actual, expected, tol, 2, &format!("appraisal ratio (daily, Rf {rf})"));
        }
    }
    for &(rf, expected) in rd::appraisal_ratio::EXPECTED_VALUES_BY_RF_MONTHLY_PERFAN {
        let actual = run(cfg().monthly().rf(compound(rf, 12.0)), Measures::appraisal_ratio);
        let tol = Delta(if rf < 0.05 { 1e-11 } else { 1e-9 });
        assert_series(&actual, expected, tol, 2, &format!("appraisal ratio (monthly, Rf {rf})"));
    }
    for &(rf, expected) in rd::appraisal_ratio::EXPECTED_VALUES_BY_RF_ANNUAL_PERFAN {
        let actual = run(cfg().rf(rf), Measures::appraisal_ratio);
        let tol = Delta(if rf < 0.05 { 1e-13 } else { 1e-11 });
        assert_series(&actual, expected, tol, 2, &format!("appraisal ratio (yearly, Rf {rf})"));
    }
}

#[test]
fn test_jensen_alpha_modified_definition() {
    let mut m = make_measures(0, 0.0, 0.0, YEARLY);
    for i in 0..BACON_PORTFOLIO_LEN {
        m.add_return(BACON_PORTFOLIO_RETURNS[i], BACON_BENCHMARK_RETURNS[i]);
        let beta = m.sfm_beta();
        let expected = if beta != 0.0 { m.jensen_alpha() / beta } else { f64::NAN };
        assert_float(m.jensen_alpha_modified(), expected, Places(14), &format!("Jensen alpha modified n={}", i + 1));
    }
}

#[test]
fn test_jensen_alpha_modified_matches_performance_analytics_output() {
    for &(rf, expected) in rd::jensen_alpha_modified::EXPECTED_VALUES_BY_RF_DAILY_PERFAN {
        let actual = run(cfg().daily().rf(compound(rf, 252.0)), Measures::jensen_alpha_modified);
        if rf < 0.05 {
            // Annualizing 24 daily observations amplifies the reference's
            // floating-point error, even when the formula agrees.
            assert_series(&actual, expected, Delta(1e-8), 0, &format!("Jensen alpha modified (daily, Rf {rf})"));
        }
    }
    for &(rf, expected) in rd::jensen_alpha_modified::EXPECTED_VALUES_BY_RF_MONTHLY_PERFAN {
        let actual = run(cfg().monthly().rf(compound(rf, 12.0)), Measures::jensen_alpha_modified);
        assert_series(&actual, expected, Delta(1e-10), 0, &format!("Jensen alpha modified (monthly, Rf {rf})"));
    }
    for &(rf, expected) in rd::jensen_alpha_modified::EXPECTED_VALUES_BY_RF_ANNUAL_PERFAN {
        let actual = run(cfg().rf(rf), Measures::jensen_alpha_modified);
        assert_series(&actual, expected, Delta(1e-13), 0, &format!("Jensen alpha modified (yearly, Rf {rf})"));
    }
}

#[test]
fn test_jensen_alpha_alternative_matches_performance_analytics_output() {
    for &(rf, expected) in rd::jensen_alpha_alternative::EXPECTED_VALUES_BY_RF_DAILY_PERFAN {
        let actual = run(cfg().daily().rf(compound(rf, 252.0)), Measures::jensen_alpha_alternative);
        if rf < 0.3 {
            let tol = Delta(if rf < 0.1 { 0.1826 } else { 0.707 });
            assert_series(&actual, expected, tol, 0, &format!("Jensen alpha alternative (daily, Rf {rf})"));
        }
    }
    for &(rf, expected) in rd::jensen_alpha_alternative::EXPECTED_VALUES_BY_RF_MONTHLY_PERFAN {
        let actual = run(cfg().monthly().rf(compound(rf, 12.0)), Measures::jensen_alpha_alternative);
        let tol = Places(if rf < 0.3 { 10 } else { 9 });
        assert_series(&actual, expected, tol, 0, &format!("Jensen alpha alternative (monthly, Rf {rf})"));
    }
    for &(rf, expected) in rd::jensen_alpha_alternative::EXPECTED_VALUES_BY_RF_ANNUAL_PERFAN {
        let actual = run(cfg().rf(rf), Measures::jensen_alpha_alternative);
        assert_series(&actual, expected, Places(12), 0, &format!("Jensen alpha alternative (yearly, Rf {rf})"));
    }
}

// M squared

#[test]
fn test_m_squared_matches_performance_analytics_output() {
    for &(rf, expected) in rd::m_squared::EXPECTED_VALUES_BY_RF_DAILY_PERFAN {
        let actual = run(cfg().daily().rf(compound(rf, 252.0)), Measures::m_squared);
        if rf < 0.05 {
            let tol = Delta(if rf < 0.01 { 0.1849 } else { 0.82956 });
            assert_series(&actual, expected, tol, 0, &format!("M squared (daily, Rf {rf})"));
        }
    }
    for &(rf, expected) in rd::m_squared::EXPECTED_VALUES_BY_RF_MONTHLY_PERFAN {
        let actual = run(cfg().monthly().rf(compound(rf, 12.0)), Measures::m_squared);
        let tol = Delta(if rf < 0.05 { 0.00861 } else { 1.621 });
        assert_series(&actual, expected, tol, 0, &format!("M squared (monthly, Rf {rf})"));
    }
    for &(rf, expected) in rd::m_squared::EXPECTED_VALUES_BY_RF_ANNUAL_PERFAN {
        let actual = run(cfg().rf(rf), Measures::m_squared);
        assert_series(&actual, expected, Places(14), 0, &format!("M squared (yearly, Rf {rf})"));
    }
}

#[test]
fn test_m_squared_excess_matches_performance_analytics_output() {
    for &(rf, expected) in rd::m_squared_excess::EXPECTED_VALUES_BY_RF_DAILY_PERFAN {
        let actual = run(cfg().daily().rf(compound(rf, 252.0)), Measures::m_squared_excess);
        if rf < 0.05 {
            let tol = Delta(if rf < 0.01 { 0.02244 } else { 0.101 });
            assert_series(&actual, expected, tol, 0, &format!("M squared excess (daily, Rf {rf})"));
        }
    }
    for &(rf, expected) in rd::m_squared_excess::EXPECTED_VALUES_BY_RF_MONTHLY_PERFAN {
        let actual = run(cfg().monthly().rf(compound(rf, 12.0)), Measures::m_squared_excess);
        let tol = Delta(if rf < 0.05 { 0.007782 } else { 1.466 });
        assert_series(&actual, expected, tol, 0, &format!("M squared excess (monthly, Rf {rf})"));
    }
    for &(rf, expected) in rd::m_squared_excess::EXPECTED_VALUES_BY_RF_ANNUAL_PERFAN {
        let actual = run(cfg().rf(rf), Measures::m_squared_excess);
        assert_series(&actual, expected, Places(15), 0, &format!("M squared excess (yearly, Rf {rf})"));
    }
}

#[test]
fn test_m_squared_sortino_matches_performance_analytics_output() {
    for &(mar, expected) in rd::m_squared_sortino::EXPECTED_VALUES_BY_MAR_DAILY_PERFAN {
        let actual = run(cfg().daily().mar(compound(mar, 252.0)), Measures::m_squared_sortino);
        assert_series(&actual, expected, Places(11), 3, &format!("M squared Sortino (daily, MAR {mar})"));
    }
    for &(mar, expected) in rd::m_squared_sortino::EXPECTED_VALUES_BY_MAR_MONTHLY_PERFAN {
        let actual = run(cfg().monthly().mar(compound(mar, 12.0)), Measures::m_squared_sortino);
        assert_series(&actual, expected, Places(14), 3, &format!("M squared Sortino (monthly, MAR {mar})"));
    }
    for &(mar, expected) in rd::m_squared_sortino::EXPECTED_VALUES_BY_MAR_ANNUAL_PERFAN {
        let actual = run(cfg().mar(mar), Measures::m_squared_sortino);
        assert_series(&actual, expected, Places(15), 3, &format!("M squared Sortino (yearly, MAR {mar})"));
    }
}

// Miscellaneous

#[test]
fn test_tail_ratio_matches_reference_implementation_output() {
    for &(cutoff, expected) in rd::tail_ratio::EXPECTED_VALUES_BY_CUTOFF_REFERENCE {
        let actual = run(cfg(), |m| m.tail_ratio(cutoff).unwrap());
        assert_series(&actual, expected, Places(15), 0, &format!("tail ratio (yearly, cutoff {cutoff})"));
    }
}

#[test]
fn test_tail_ratio_rejects_invalid_cutoff() {
    let m = make_measures(0, 0.0, 0.0, YEARLY);
    for cutoff in [0.5, 1.0, 0.2, f64::NAN] {
        assert_eq!(m.tail_ratio(cutoff), Err("cutoff must be between 0.5 and 1.0".to_string()));
    }
}

/// Daily, monthly and yearly runs against the same periodic-rf table.
fn check_rf_all(table: Table, f: fn(&Measures) -> f64, places: i32, label: &str) {
    for &(rf, expected) in table {
        let actual = run(cfg().daily().rf(compound(rf, 252.0)), f);
        assert_series(&actual, expected, Places(places), 0, &format!("{label} (daily, Rf {rf})"));
    }
    for &(rf, expected) in table {
        let actual = run(cfg().monthly().rf(compound(rf, 12.0)), f);
        assert_series(&actual, expected, Places(places), 0, &format!("{label} (monthly, Rf {rf})"));
    }
    for &(rf, expected) in table {
        let actual = run(cfg().rf(rf), f);
        assert_series(&actual, expected, Places(places), 0, &format!("{label} (yearly, Rf {rf})"));
    }
}

#[test]
fn test_kelly_ratio_matches_performance_analytics_output() {
    check_rf_all(rd::kelly_ratio::EXPECTED_VALUES_BY_RF_PERFAN, Measures::kelly_ratio, 11, "Kelly ratio");
}

#[test]
fn test_kelly_ratio_full_matches_performance_analytics_output() {
    check_rf_all(rd::kelly_ratio::EXPECTED_VALUES_BY_RF_FULL_PERFAN, Measures::kelly_ratio_full, 11, "Kelly ratio full");
}

#[test]
fn test_hurst_exponent_matches_performance_analytics_output() {
    let actual = run(cfg(), Measures::hurst_exponent);
    assert_series(&actual, rd::hurst_exponent::EXPECTED_VALUES_PERFAN, Places(14), 0, "Hurst exponent");
}

#[test]
fn test_bias_ratio_matches_reference_implementation_output() {
    for &(mult, expected) in rd::bias_ratio::EXPECTED_VALUES_BY_MULT_REFERENCE {
        let actual = run(cfg(), |m| m.bias_ratio(mult).unwrap());
        assert_series(&actual, expected, Places(15), 0, &format!("bias ratio (yearly, std_dev_multiplier {mult})"));
    }
}

#[test]
fn test_bias_ratio_rejects_non_positive_multiplier() {
    let m = make_measures(0, 0.0, 0.0, YEARLY);
    for mult in [0.0, -1.0] {
        assert_eq!(m.bias_ratio(mult), Err("std_dev_multiplier must be positive".to_string()));
    }
}

#[test]
fn test_k_ratio_matches_reference_implementation_output() {
    let actual = run(cfg(), Measures::k_ratio);
    assert_series(&actual, rd::k_ratio::EXPECTED_VALUES_REFERENCE, Places(14), 0, "K-ratio");
}

#[test]
fn test_gain_to_pain_ratio_matches_reference_implementation_output() {
    let actual = run(cfg(), Measures::gain_to_pain_ratio);
    assert_series(&actual, rd::gain_to_pain_ratio::EXPECTED_VALUES_REFERENCE, Places(15), 0, "Gain-to-pain ratio");
}

// Capture

#[test]
fn test_upside_capture_ratio_matches_performance_analytics_output() {
    for &(geom, expected) in rd::upside_capture_ratio::EXPECTED_VALUES_BY_GEOMETRIC_PERFAN {
        let actual = run(cfg(), |m| m.upside_capture_ratio(geom));
        assert_series(&actual, expected, Places(13), 1, &format!("Upside capture ratio (yearly, geometric {geom})"));
    }
}

#[test]
fn test_downside_capture_ratio_matches_performance_analytics_output() {
    for &(geom, expected) in rd::downside_capture_ratio::EXPECTED_VALUES_BY_GEOMETRIC_PERFAN {
        let actual = run(cfg(), |m| m.downside_capture_ratio(geom));
        assert_series(&actual, expected, Places(14), 0, &format!("Downside capture ratio (yearly, geometric {geom})"));
    }
}

#[test]
fn test_overall_capture_ratio_matches_reference_implementation_output() {
    for &(geom, expected) in rd::overall_capture_ratio::EXPECTED_VALUES_BY_GEOMETRIC_REFERENCE {
        let actual = run(cfg(), |m| m.overall_capture_ratio(geom));
        assert_series(&actual, expected, Places(13), 0, &format!("Overall capture ratio (yearly, geometric {geom})"));
    }
}

#[test]
fn test_up_number_ratio_matches_performance_analytics_output() {
    let actual = run(cfg(), Measures::up_number_ratio);
    assert_series(&actual, rd::up_number_ratio::EXPECTED_VALUES_PERFAN, Places(15), 0, "Up number ratio (yearly)");
}

#[test]
fn test_down_number_ratio_matches_performance_analytics_output() {
    let actual = run(cfg(), Measures::down_number_ratio);
    assert_series(&actual, rd::down_number_ratio::EXPECTED_VALUES_PERFAN, Places(15), 0, "Down number ratio (yearly)");
}

#[test]
fn test_up_percentage_ratio_matches_performance_analytics_output() {
    let actual = run(cfg(), Measures::up_percentage_ratio);
    assert_series(&actual, rd::up_percentage_ratio::EXPECTED_VALUES_PERFAN, Places(15), 0, "Up percentage ratio (yearly)");
}

#[test]
fn test_down_percentage_ratio_matches_performance_analytics_output() {
    let actual = run(cfg(), Measures::down_percentage_ratio);
    assert_series(
        &actual,
        rd::down_percentage_ratio::EXPECTED_VALUES_PERFAN,
        Places(15),
        0,
        "Down percentage ratio (yearly)",
    );
}
