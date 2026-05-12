//! Evening Doji Star pattern (3-candle bearish reversal).

use crate::candlestick_patterns::CandlestickPatterns;
use crate::candlestick_patterns::core::{is_black, is_real_body_gap_up, is_white, real_body_len};
use crate::fuzzy;

const EVENING_DOJI_STAR_PENETRATION_FACTOR: f64 = 0.3;

/// Evening Doji Star: a three-candle bearish reversal pattern.
///
/// Must have:
/// - first candle: long white real body,
/// - second candle: doji that gaps up (real body gap up from the first),
/// - third candle: black real body that moves well within the first candle's
/// real body.
///
/// The meaning of "long" is specified with self._long_body.
/// The meaning of "doji" is specified with self._doji_body.
/// The meaning of "short" is specified with self._short_body.
///
/// Category A: always bearish (continuous).
///
/// Returns:
/// Continuous float in [-100, 0].  Always bearish.
pub fn evening_doji_star(cp: &CandlestickPatterns) -> f64 {
    if !cp.enough(3, &[&cp.long_body, &cp.doji_body, &cp.short_body]) {
        return 0.0;
    }

    let b1 = cp.bar(3);
    let b2 = cp.bar(2);
    let b3 = cp.bar(1);

    // Crisp gates: color checks and gap.
    if !(is_white(b1.o, b1.c)
        && is_real_body_gap_up(b1.o, b1.c, b2.o, b2.c)
        && is_black(b3.o, b3.c)) {
        return 0.0;
    }

    // Fuzzy conditions.
    // c3 < c1 - rb1 * penetration
    let mu_long1 = cp.mu_greater(real_body_len(b1.o, b1.c), &cp.long_body, 3);
    let mu_doji2 = cp.mu_less(real_body_len(b2.o, b2.c), &cp.doji_body, 2);

    let rb1 = real_body_len(b1.o, b1.c);
    let threshold = b1.c - rb1 * EVENING_DOJI_STAR_PENETRATION_FACTOR;
    let width = cp.fuzz_ratio * rb1 * EVENING_DOJI_STAR_PENETRATION_FACTOR;
    let mu_penetration = cp.mu_lt_raw(b3.c, threshold, width);

    let confidence = fuzzy::t_product_all(&[mu_long1, mu_doji2, mu_penetration]);
    -confidence * 100.0
}
