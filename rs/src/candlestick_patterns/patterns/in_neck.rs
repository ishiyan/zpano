//! In Neck pattern (2-candle bearish continuation).

use crate::candlestick_patterns::CandlestickPatterns;
use crate::candlestick_patterns::core::{is_black, is_white, real_body_len};
use crate::fuzzy;

/// In Neck: a two-candle bearish continuation pattern.
///
/// Must have:
/// - first candle: long black,
/// - second candle: white, opens below the prior low, closes slightly
/// into the prior real body (close near the prior close).
///
/// The meaning of "long" is specified with self._long_body.
/// The meaning of "near" is specified with self._near.
///
/// Category A: always bearish (continuous).
///
/// Returns:
/// Continuous float in [-100, 0].  Always bearish.
pub fn in_neck(cp: &CandlestickPatterns) -> f64 {
    if !cp.enough(2, &[&cp.long_body, &cp.near]) {
        return 0.0;
    }

    let b1 = cp.bar(2);
    let b2 = cp.bar(1);

    // Crisp gates: color checks and open below prior low.
    if !(is_black(b1.o, b1.c) && is_white(b2.o, b2.c) && b2.o < b1.l) {
        return 0.0;
    }

    // Fuzzy conditions.
    let mu_long1 = cp.mu_greater(real_body_len(b1.o, b1.c), &cp.long_body, 2);
    // Close near prior close: crisp was abs(c2-c1) < near_avg.
    // Model as mu_less(abs_diff, near_avg) — crossover at near boundary.
    let mu_near_close = cp.mu_less((b2.c - b1.c).abs(), &cp.near, 1);

    let confidence = fuzzy::t_product_all(&[mu_long1, mu_near_close]);
    -confidence * 100.0
}
