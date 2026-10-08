//! Tests for the VaR / ES helpers (Python `test_risk_helpers.py`).

use std::collections::VecDeque;

use crate::streaming_kbn::RawMomentsKleinKbn;

use super::es::{es_cornish_fisher, es_gaussian, es_historical};
use super::test_support::almost_equal;
use super::var::{var_cornish_fisher, var_gaussian, var_historical};

#[test]
fn test_historical_quantile_tail_and_risk_free_rate() {
    let returns: VecDeque<f64> = VecDeque::from(vec![-0.2, -0.1, 0.0, 0.1]);
    assert!(almost_equal(var_historical(&returns, 0.0, 0.75), 0.125, 15));
    assert!(almost_equal(es_historical(&returns, 0.0, 0.75), 0.2, 15));
    assert!(almost_equal(var_historical(&returns, 0.01, 0.75), 0.135, 15));
    assert!(almost_equal(es_historical(&returns, 0.01, 0.75), 0.21, 15));
    assert_eq!(Vec::from(returns), vec![-0.2, -0.1, 0.0, 0.1]);
}

#[test]
fn test_empty_historical_inputs() {
    // Python also tests `None`, which has no Rust counterpart.
    let empty_deque: VecDeque<f64> = VecDeque::new();
    assert!(var_historical(&empty_deque, 0.0, 0.95).is_nan());
    assert!(es_historical(&empty_deque, 0.0, 0.95).is_nan());
    let empty_slice: &[f64] = &[];
    assert!(var_historical(empty_slice, 0.0, 0.95).is_nan());
    assert!(es_historical(empty_slice, 0.0, 0.95).is_nan());
}

#[test]
fn test_cornish_fisher_falls_back_for_one_sample() {
    let mut moments = RawMomentsKleinKbn::default();
    moments.update(0.02);
    assert!(almost_equal(var_gaussian(&moments, 0.95), -0.02, 15));
    assert!(almost_equal(es_gaussian(&moments, 0.95), -0.02, 15));
    assert_eq!(var_cornish_fisher(&moments, 0.95), var_gaussian(&moments, 0.95));
    assert_eq!(es_cornish_fisher(&moments, 0.95), es_gaussian(&moments, 0.95));
}
