//! Documented formula checks and rolling-window equivalence.

use super::*;
use crate::performance::core::test_support::{gauss, prod, Rng};

// TestDocumentedFormulas: formula checks alongside the R fixtures.

#[test]
fn test_documented_formulas_drawdown_risk_and_ratios() {
    for rf in [0.0, 0.05] {
        let mut m = make_measures(0, rf, 0.0, YEARLY);
        for i in 0..BACON_PORTFOLIO_LEN {
            m.add_return(BACON_PORTFOLIO_RETURNS[i], BACON_BENCHMARK_RETURNS[i]);
            let drawdowns = m.drawdowns_high_watermark();
            let n = (i + 1) as f64;
            let pain = -fsum(&drawdowns) / n;
            let squares: Vec<f64> = drawdowns.iter().map(|x| x * x).collect();
            let ulcer = (fsum(&squares) / n).sqrt();
            let geometric_return = prod(BACON_PORTFOLIO_RETURNS[..=i].iter().map(|x| 1.0 + x)).powf(1.0 / n) - 1.0;
            assert_float(m.pain_index(), pain, Places(14), &format!("pain index rf {rf} step {i}"));
            assert_float(m.ulcer_index(), ulcer, Places(14), &format!("ulcer index rf {rf} step {i}"));
            if pain > 0.0 {
                assert_float(m.pain_ratio(), (geometric_return - rf) / pain, Places(12), &format!("pain ratio rf {rf} step {i}"));
            }
            if ulcer > 0.0 {
                assert_float(
                    m.martin_ratio(),
                    (geometric_return - rf) / ulcer,
                    Places(12),
                    &format!("martin ratio rf {rf} step {i}"),
                );
            }
        }
    }
}

#[test]
fn test_documented_formulas_cdar_average_and_alpha() {
    let mut m = make_measures(0, 0.0, 0.0, MONTHLY);
    for i in 0..BACON_PORTFOLIO_LEN {
        m.add_return(BACON_PORTFOLIO_RETURNS[i], BACON_BENCHMARK_RETURNS[i]);
        let mut drawdowns = m.drawdowns_high_watermark();
        drawdowns.sort_by(|a, b| a.partial_cmp(b).unwrap());
        let len = drawdowns.len();
        for confidence in [0.9, 0.95] {
            let position = (1.0 - confidence) * (len - 1) as f64;
            let lo = position as usize;
            let q = drawdowns[lo] + (position - lo as f64) * (drawdowns[(lo + 1).min(len - 1)] - drawdowns[lo]);
            let tail: Vec<f64> = drawdowns.iter().copied().filter(|&d| d <= q).collect();
            let expected_cdar = if q < 0.0 { -fsum(&tail) / tail.len() as f64 } else { 0.0 };
            assert_float(
                m.cdar_average(confidence).unwrap(),
                expected_cdar,
                Places(14),
                &format!("cdar average confidence {confidence} step {i}"),
            );
            let beta = m.cdar_beta(confidence).unwrap();
            if beta.is_finite() {
                let portfolio_mean = fmean(&BACON_PORTFOLIO_RETURNS[..len]);
                let benchmark_mean = fmean(&BACON_BENCHMARK_RETURNS[..len]);
                let expected_alpha =
                    (1.0 + portfolio_mean).powf(12.0) - 1.0 - beta * ((1.0 + benchmark_mean).powf(12.0) - 1.0);
                assert_float(
                    m.cdar_alpha(confidence).unwrap(),
                    expected_alpha,
                    Places(13),
                    &format!("cdar alpha confidence {confidence} step {i}"),
                );
            }
        }
    }
}

#[test]
fn test_documented_formulas_m_squared_and_jensen_alpha_alternative() {
    for (periods, periodic_rf) in [(1.0, 0.05), (12.0, 0.05), (252.0, 0.01)] {
        let annual_rf = compound(periodic_rf, periods);
        let mut m = Measures::new(periods, annual_rf, 0.0, 0).unwrap();
        for i in 0..BACON_PORTFOLIO_LEN {
            m.add_return(BACON_PORTFOLIO_RETURNS[i], BACON_BENCHMARK_RETURNS[i]);
            let n = i + 1;
            if n < 2 {
                continue;
            }
            let nf = n as f64;
            let portfolio = &BACON_PORTFOLIO_RETURNS[..n];
            let benchmark = &BACON_BENCHMARK_RETURNS[..n];
            let p_ann = prod(portfolio.iter().map(|x| 1.0 + x)).powf(periods / nf) - 1.0;
            let b_ann = prod(benchmark.iter().map(|x| 1.0 + x)).powf(periods / nf) - 1.0;
            let p_std = pstdev(portfolio);
            let b_std = pstdev(benchmark);
            let scale = b_std / p_std;
            let expected_m2 = p_ann * scale + annual_rf * (1.0 - scale);
            assert_float(m.m_squared(), expected_m2, Rel(1e-11, 1e-10), &format!("M squared periods={periods} n={n}"));
            let expected_excess = (1.0 + expected_m2) / (1.0 + b_ann) - 1.0;
            assert_float(
                m.m_squared_excess(),
                expected_excess,
                Rel(1e-11, 1e-10),
                &format!("M squared excess periods={periods} n={n}"),
            );
            let systematic_risk = m.sfm_beta().abs() * stdev(benchmark) * periods.sqrt();
            if systematic_risk > 0.0 && systematic_risk.is_finite() {
                assert_float(
                    m.jensen_alpha_alternative(),
                    m.jensen_alpha() / systematic_risk,
                    Rel(1e-11, 1e-10),
                    &format!("Jensen alpha alternative periods={periods} n={n}"),
                );
            }
        }
    }
}

#[test]
fn test_documented_formulas_m_squared_equal_volatility_with_extreme_rate() {
    let mut m = Measures::new(252.0, 1.3f64.powf(252.0) - 1.0, 0.0, 0).unwrap();
    for (portfolio, benchmark) in [(0.125, 0.25), (0.375, 0.5)] {
        m.add_return(portfolio, benchmark);
    }
    // Binary-exact inputs give exactly equal portfolio and benchmark
    // volatility, so the annual risk-free terms must cancel.
    let expected = (1.125f64 * 1.375).powf(126.0) - 1.0;
    assert_float(m.m_squared(), expected, Rel(1e-14, 0.0), "M squared");
}

// TestRollingWindow: a rolling window gives the same results as a fresh
// instance fed only the last N returns.

#[test]
fn test_rolling_window_rolling_matches_fresh() {
    // At every step, including while the window is still filling, every
    // public measure (methods with default arguments) of a rolling-window
    // instance equals that of a fresh instance fed only the returns in the
    // window.
    let mut rng = Rng::new(42);
    let random_returns: Vec<f64> = (0..150).map(|_| gauss(&mut rng, 0.002, 0.03)).collect();
    let random_benchmark: Vec<f64> = (0..150).map(|_| gauss(&mut rng, 0.001, 0.025)).collect();
    struct Config<'a> {
        window: usize,
        periods_per_annum: f64,
        annual_rf: f64,
        annual_mar: f64,
        returns: &'a [f64],
        benchmark: &'a [f64],
    }
    let configs = [
        Config {
            window: 10,
            periods_per_annum: 1.0,
            annual_rf: 0.0,
            annual_mar: 0.0,
            returns: &BACON_PORTFOLIO_RETURNS,
            benchmark: &BACON_BENCHMARK_RETURNS,
        },
        Config {
            window: 30,
            periods_per_annum: 12.0,
            annual_rf: 0.05,
            annual_mar: 0.03,
            returns: &random_returns,
            benchmark: &random_benchmark,
        },
    ];
    let names = public_measures();
    for c in &configs {
        let window = c.window;
        let create = || Measures::new(c.periods_per_annum, c.annual_rf, c.annual_mar, 0).unwrap();
        let mut rolling = Measures::new(c.periods_per_annum, c.annual_rf, c.annual_mar, window).unwrap();
        for i in 0..c.returns.len() {
            rolling.add_return(c.returns[i], c.benchmark[i]);
            let mut fresh = create();
            for j in (i + 1).saturating_sub(window)..=i {
                fresh.add_return(c.returns[j], c.benchmark[j]);
            }
            for (name, eval) in &names {
                let prefix = format!("window {window} step {i} {name}");
                match (eval(&rolling), eval(&fresh)) {
                    (Value::L(actual), Value::L(expected)) => {
                        assert_series(&actual, &expected, Places(12), 0, &prefix);
                        assert_eq!(actual.len(), expected.len(), "{prefix}");
                    }
                    (Value::B(actual), Value::B(expected)) => assert_eq!(actual, expected, "{prefix}"),
                    (Value::F(actual), Value::F(expected)) => {
                        let tol = if expected.is_finite() { Delta(1e-12 * expected.abs().max(1.0)) } else { Places(15) };
                        assert_float(actual, expected, tol, &prefix);
                    }
                    (a, e) => panic!("{prefix}: value kinds differ: {a:?} vs {e:?}"),
                }
            }
        }
    }
}
