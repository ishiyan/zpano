/// Signal composition utilities.
use crate::fuzzy::{t_product_all, s_probabilistic, f_not};

/// Combine signals with product t-norm (fuzzy AND).
pub fn signal_and(signals: &[f64]) -> f64 {
    t_product_all(signals)
}

/// Combine two signals with probabilistic s-norm (fuzzy OR).
pub fn signal_or(a: f64, b: f64) -> f64 {
    s_probabilistic(a, b)
}

/// Negate a signal (fuzzy complement). Returns `1 - signal`.
pub fn signal_not(signal: f64) -> f64 {
    f_not(signal)
}

/// Filter weak signals below `min_strength` to zero.
/// Signals at or above the threshold pass through unchanged.
pub fn signal_strength(signal: f64, min_strength: f64) -> f64 {
    if signal >= min_strength { signal } else { 0.0 }
}

#[cfg(test)]
mod tests {
    use super::*;

    fn almost_equal(a: f64, b: f64, eps: f64) -> bool { (a - b).abs() < eps }

    #[test]
    fn test_and_all_high() { assert!(almost_equal(signal_and(&[0.9, 0.8, 0.95]), 0.9 * 0.8 * 0.95, 1e-10)); }
    #[test]
    fn test_and_one_zero() { assert!(almost_equal(signal_and(&[0.9, 0.0, 0.8]), 0.0, 1e-10)); }
    #[test]
    fn test_and_all_one() { assert!(almost_equal(signal_and(&[1.0, 1.0, 1.0]), 1.0, 1e-10)); }
    #[test]
    fn test_and_two() { assert!(almost_equal(signal_and(&[0.6, 0.7]), 0.42, 1e-10)); }
    #[test]
    fn test_or_both_high() { assert!(almost_equal(signal_or(0.8, 0.9), 0.8 + 0.9 - 0.8 * 0.9, 1e-10)); }
    #[test]
    fn test_or_one_zero() { assert!(almost_equal(signal_or(0.0, 0.7), 0.7, 1e-10)); }
    #[test]
    fn test_or_both_one() { assert!(almost_equal(signal_or(1.0, 1.0), 1.0, 1e-10)); }
    #[test]
    fn test_not_zero() { assert!(almost_equal(signal_not(0.0), 1.0, 1e-10)); }
    #[test]
    fn test_not_one() { assert!(almost_equal(signal_not(1.0), 0.0, 1e-10)); }
    #[test]
    fn test_not_half() { assert!(almost_equal(signal_not(0.5), 0.5, 1e-10)); }
    #[test]
    fn test_strength_above() { assert_eq!(signal_strength(0.8, 0.5), 0.8); }
    #[test]
    fn test_strength_below() { assert_eq!(signal_strength(0.3, 0.5), 0.0); }
    #[test]
    fn test_strength_at() { assert_eq!(signal_strength(0.5, 0.5), 0.5); }
    #[test]
    fn test_strength_just_below() { assert_eq!(signal_strength(0.499, 0.5), 0.0); }
}
