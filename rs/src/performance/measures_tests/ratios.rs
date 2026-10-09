//! Sortino, Omega, Kappa, win/loss and related ratio tests.

use super::*;
use crate::performance::reference_data as rd;

fn check_by_mar(table: &[(f64, &[f64])], f: fn(&Measures) -> f64, places: i32, label: &str) {
    for &(mar, expected) in table {
        let actual = run(cfg().mar(mar), f);
        assert_series(&actual, expected, Places(places), 0, &format!("{label} MAR {mar}"));
    }
}

fn check_plain(expected: &[f64], f: fn(&Measures) -> f64, places: i32, label: &str) {
    let actual = run(cfg(), f);
    assert_series(&actual, expected, Places(places), 0, label);
}

// TestSortinoRatio

#[test]
fn test_sortino_ratio_matches_performance_analytics_output() {
    check_by_mar(rd::sortino_ratio::EXPECTED_VALUES_BY_MAR, Measures::sortino_ratio, 14, "sortino ratio");
}

#[test]
fn test_sortino_ratio_jack_schager_sqrt2_version() {
    for &(mar, expected) in rd::sortino_ratio::EXPECTED_VALUES_BY_MAR {
        let expected: Vec<f64> = expected.iter().map(|x| x / 2f64.sqrt()).collect();
        let actual = run(cfg().mar(mar), Measures::sortino_ratio_sqrt2);
        assert_series(&actual, &expected, Places(14), 0, &format!("sortino ratio (sqrt2) MAR {mar}"));
    }
}

#[test]
fn test_sortino_satchell_ratio_should_be_computable() {
    let actual = run(cfg(), Measures::sortino_satchell_ratio);
    assert_float(actual[actual.len() - 1], 0.3923720287950653, Places(15), "Sortino-Satchell ratio");
}

// Omega

#[test]
fn test_omega_ratio_matches_performance_analytics_output() {
    check_by_mar(rd::omega_ratio::EXPECTED_VALUES_BY_MAR, Measures::omega_ratio, 13, "omega ratio");
}

#[test]
fn test_omega_sharpe_ratio_matches_performance_analytics_output() {
    check_by_mar(rd::omega_sharpe_ratio::EXPECTED_VALUES_BY_MAR, Measures::omega_sharpe_ratio, 13, "omega Sharpe ratio");
}

#[test]
fn test_omega_excess_return_matches_performance_analytics_output() {
    for &(mar, expected) in rd::omega_excess_return::EXPECTED_VALUES_BY_MAR_WITH_BENCHMARK {
        let actual = run(cfg().daily().mar(compound(mar, 252.0)), Measures::omega_excess_return);
        assert_series(&actual, expected, Places(11), 0, &format!("omega excess return (with benchmark) MAR {mar}"));
    }
    for &(mar, expected) in rd::omega_excess_return::EXPECTED_VALUES_BY_MAR_WITH_SELF {
        let c = cfg().daily().mar(compound(mar, 252.0)).bench(&BACON_PORTFOLIO_RETURNS);
        let actual = run(c, Measures::omega_excess_return);
        assert_series(&actual, expected, Places(11), 0, &format!("omega excess return (with self) MAR {mar}"));
    }
}

// TestKappaRatio (generated data is yearly)

#[test]
fn test_kappa_ratio_matches_performance_analytics_output() {
    check_by_mar(rd::kappa_ratio::EXPECTED_VALUES_BY_MAR_ORDER_1, Measures::kappa_1_ratio, 13, "kappa 1 ratio");
    check_by_mar(rd::kappa_ratio::EXPECTED_VALUES_BY_MAR_ORDER_2, Measures::kappa_2_ratio, 14, "kappa 2 ratio");
    check_by_mar(rd::kappa_ratio::EXPECTED_VALUES_BY_MAR_ORDER_3, Measures::kappa_3_ratio, 14, "kappa 3 ratio");
    check_by_mar(rd::kappa_ratio::EXPECTED_VALUES_BY_MAR_ORDER_4, Measures::kappa_4_ratio, 14, "kappa 4 ratio");
}

// TestProspectRatio

#[test]
fn test_prospect_ratio_matches_performance_analytics_output() {
    check_by_mar(
        rd::prospect_ratio::EXPECTED_VALUES_BY_MAR_PERFAN,
        Measures::prospect_ratio_performance_analytics,
        13,
        "Prospect ratio PerformanceAnalytics version (yearly)",
    );
}

#[test]
fn test_prospect_ratio_matches_reference_implementation_output() {
    for &(mar, expected) in rd::prospect_ratio::EXPECTED_VALUES_BY_MAR_REFERENCE {
        let actual = run(cfg().mar(mar), |m| m.prospect_ratio(2.25));
        assert_series(&actual, expected, Places(15), 0, &format!("Prospect ratio (yearly, MAR {mar})"));
    }
}

// Bernardo-Ledoit, D-ratio, gain-loss

#[test]
fn test_bernardo_ledoit_ratio_matches_performance_analytics_output() {
    check_plain(rd::bernardo_ledoit_ratio::EXPECTED_VALUES, Measures::bernardo_ledoit_ratio, 13, "Bernardo-Ledoit ratio");
}

#[test]
fn test_d_ratio_matches_performance_analytics_output() {
    check_plain(rd::d_ratio::EXPECTED_VALUES_PERFAN, Measures::d_ratio, 15, "d-ratio");
}

#[test]
fn test_gain_loss_ratio_matches_performance_analytics_output() {
    check_plain(rd::bernardo_ledoit_ratio::EXPECTED_VALUES, Measures::gain_loss_ratio, 13, "gain-loss ratio");
}

// Win / loss (calculated by hand)

#[test]
fn test_mean_non_zero_return_calculated_by_hand() {
    check_plain(rd::mean_non_zero_return::EXPECTED_VALUES, Measures::mean_non_zero_return, 15, "mean non-zero return");
}

#[test]
fn test_mean_win_return_calculated_by_hand() {
    check_plain(rd::mean_win_return::EXPECTED_VALUES, Measures::mean_win_return, 15, "mean win return");
}

#[test]
fn test_mean_loss_return_calculated_by_hand() {
    check_plain(rd::mean_loss_return::EXPECTED_VALUES, Measures::mean_loss_return, 15, "mean loss return");
}

#[test]
fn test_win_rate_calculated_by_hand() {
    check_plain(rd::win_rate::EXPECTED_VALUES, Measures::win_rate, 15, "win rate");
}

#[test]
fn test_loss_rate_calculated_by_hand() {
    check_plain(rd::loss_rate::EXPECTED_VALUES, Measures::loss_rate, 15, "loss rate");
}

// TestVolatilitySkewness

#[test]
fn test_volatility_skewness_matches_performance_analytics_output() {
    for &(mar, expected) in rd::volatility_skewness::EXPECTED_VALUES_BY_MAR_VOLATILITY {
        let actual = run(cfg().mar(mar), Measures::volatility_skewness);
        assert_series(&actual, expected, Places(13), 0, &format!("volatility skewness MAR {mar}"));
    }
    for &(mar, expected) in rd::volatility_skewness::EXPECTED_VALUES_BY_MAR_VARIABILITY {
        let actual = run(cfg().mar(mar), Measures::variability_skewness);
        assert_series(&actual, expected, Places(13), 0, &format!("variability skewness MAR {mar}"));
    }
}

// TestFarinelliTibilettiRatio

#[test]
fn test_farinelli_tibiletti_ratio_should_be_computable() {
    let mar = 0.005;

    let verify = |upper: u32, lower: u32, related: &str, f: fn(&Measures) -> f64, transform: fn(f64) -> f64, places: i32| {
        let expected = run(cfg().mar(mar), f);
        let actual: Vec<f64> = run(cfg().mar(mar), |m| m.farinelli_tibiletti_ratio(upper, lower).unwrap())
            .into_iter()
            .map(transform)
            .collect();
        assert_series(
            &actual,
            &expected,
            Places(places),
            0,
            &format!("Farinelli-Tibiletti ratio (u {upper}, l {lower}) vs {related}"),
        );
    };

    verify(1, 1, "omega_ratio", Measures::omega_ratio, |r| r, 14);
    verify(1, 1, "kappa_1_ratio", Measures::kappa_1_ratio, |r| r - 1.0, 14);
    verify(1, 2, "upside_potential_ratio", Measures::upside_potential_ratio, |r| r, 15);
    verify(2, 2, "volatility_skewness", Measures::volatility_skewness, |r| r, 14);
    verify(2, 2, "variability_skewness", Measures::variability_skewness, |r| r * r, 13);

    let n = BACON_PORTFOLIO_LEN as f64;
    for upper in 1..=4u32 {
        for lower in 1..=4u32 {
            let upm = BACON_PORTFOLIO_RETURNS
                .iter()
                .map(|r| (r - mar).max(0.0).powf(upper as f64))
                .sum::<f64>()
                / n;
            let lpm = BACON_PORTFOLIO_RETURNS
                .iter()
                .map(|r| (mar - r).max(0.0).powf(lower as f64))
                .sum::<f64>()
                / n;
            let expected = upm.powf(1.0 / upper as f64) / lpm.powf(1.0 / lower as f64);
            let actual = run(cfg().mar(mar), |m| m.farinelli_tibiletti_ratio(upper, lower).unwrap());
            assert_float(
                actual[actual.len() - 1],
                expected,
                Places(15),
                &format!("Farinelli-Tibiletti ratio (u {upper}, l {lower}) vs manual calculation"),
            );
        }
    }
}

#[test]
fn test_farinelli_tibiletti_ratio_rejects_invalid_orders() {
    let m = make_measures(0, 0.0, 0.0, YEARLY);
    for order in [0, 5] {
        assert_eq!(m.farinelli_tibiletti_ratio(order, 2), Err("upper_order must be 1, 2, 3, or 4".to_string()));
        assert_eq!(m.farinelli_tibiletti_ratio(2, order), Err("lower_order must be 1, 2, 3, or 4".to_string()));
    }
    assert_eq!(m.farinelli_tibiletti_ratio(0, 0), Err("upper_order must be 1, 2, 3, or 4".to_string()));
}

// TestRachevRatio

#[test]
fn test_rachev_ratio_matches_performance_analytics_output() {
    for (alpha, table) in [
        (0.05, rd::rachev_ratio::EXPECTED_VALUES_BY_BETA_RF_ALFA_0_05),
        (0.1, rd::rachev_ratio::EXPECTED_VALUES_BY_BETA_RF_ALFA_0_1),
    ] {
        for &(beta, bundle) in table {
            for &(rf, expected) in bundle {
                let actual = run(cfg().rf(rf), |m| m.rachev_ratio(alpha, beta).unwrap());
                assert_series(&actual, expected, Places(14), 0, &format!("Rachev ratio (alpha {alpha} beta {beta} Rf {rf})"));
            }
        }
    }
}

#[test]
fn test_rachev_ratio_validation_order() {
    let mut m = make_measures(0, 0.0, 0.0, YEARLY);
    m.add_return(0.01, 0.01);
    // Fewer than two returns: NaN before validation.
    assert!(m.rachev_ratio(2.0, 2.0).unwrap().is_nan());
    m.add_return(-0.01, 0.01);
    assert_eq!(m.rachev_ratio(0.0, 0.1), Err("alpha must be between 0 and 1".to_string()));
    assert_eq!(m.rachev_ratio(0.1, 1.0), Err("beta must be between 0 and 1".to_string()));
    assert!(m.rachev_ratio(0.1, 0.1).is_ok());
}
