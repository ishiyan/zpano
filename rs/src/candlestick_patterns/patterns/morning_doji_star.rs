//! Morning Doji Star pattern (3-candle bullish reversal).

use crate::candlestick_patterns::CandlestickPatterns;
use crate::candlestick_patterns::core::{is_black, is_real_body_gap_down, is_white, real_body_len};
use crate::fuzzy;

const MORNING_DOJI_STAR_PENETRATION_FACTOR: f64 = 0.3;

/// Morning Doji Star: a three-candle bullish reversal pattern.
///
/// Must have:
/// - first candle: long black real body,
/// - second candle: doji that gaps down (real body gap down from the first),
/// - third candle: white real body that closes well within the first candle's
/// real body.
///
/// The meaning of "long" is specified with self._long_body.
/// The meaning of "doji" is specified with self._doji_body.
/// The meaning of "short" is specified with self._short_body.
///
/// Category A: always bullish (continuous).
///
/// Returns:
/// Continuous float in [0, +100].  Always bullish.
pub fn morning_doji_star(cp: &CandlestickPatterns) -> f64 {
    if !cp.enough(3, &[&cp.long_body, &cp.doji_body, &cp.short_body]) {
        return 0.0;
    }

    let b1 = cp.bar(3);
    let b2 = cp.bar(2);
    let b3 = cp.bar(1);

    // Crisp gates: color checks and gap.
    if !(is_black(b1.o, b1.c)
        && is_real_body_gap_down(b1.o, b1.c, b2.o, b2.c)
        && is_white(b3.o, b3.c)) {
        return 0.0;
    }

    // Fuzzy conditions.
    // c3 > c1 + rb1 * penetration
    let mu_long1 = cp.mu_greater(real_body_len(b1.o, b1.c), &cp.long_body, 3);
    let mu_doji2 = cp.mu_less(real_body_len(b2.o, b2.c), &cp.doji_body, 2);

    let rb1 = real_body_len(b1.o, b1.c);
    let threshold = b1.c + rb1 * MORNING_DOJI_STAR_PENETRATION_FACTOR;
    let width = cp.fuzz_ratio * rb1 * MORNING_DOJI_STAR_PENETRATION_FACTOR;
    let mu_penetration = cp.mu_gt_raw(b3.c, threshold, width);

    let confidence = fuzzy::t_product_all(&[mu_long1, mu_doji2, mu_penetration]);
    confidence * 100.0
}
