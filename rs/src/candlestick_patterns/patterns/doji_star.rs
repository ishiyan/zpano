//! Doji Star pattern (2-candle reversal).

use crate::candlestick_patterns::CandlestickPatterns;
use crate::candlestick_patterns::core::{is_real_body_gap_down, is_real_body_gap_up, real_body_len};
use crate::fuzzy;

/// Doji Star: a two-candle reversal pattern.
///
/// Must have:
/// - first candle: long real body,
/// - second candle: doji that gaps away from the first candle.
///
/// - bearish: first candle is long white, doji gaps up,
/// - bullish: first candle is long black, doji gaps down.
///
/// The meaning of "long" is specified with self._long_body.
/// The meaning of "doji" is specified with self._doji_body.
///
/// Category B: direction from 1st candle color (continuous).
///
/// Returns:
/// Continuous float in [-100, +100].
pub fn doji_star(cp: &CandlestickPatterns) -> f64 {
    if !cp.enough(2, &[&cp.long_body, &cp.doji_body]) {
        return 0.0;
    }

    let b1 = cp.bar(2);
    let b2 = cp.bar(1);

    // Direction: opposite of 1st candle color.
    let color1: i32 = if b1.c < b1.o { -1 } else { 1 };

    // Crisp gates: gap direction must match color.
    if color1 == 1 && !is_real_body_gap_up(b1.o, b1.c, b2.o, b2.c) {
        return 0.0;
    }
    if color1 == -1 && !is_real_body_gap_down(b1.o, b1.c, b2.o, b2.c) {
        return 0.0;
    }

    // Fuzzy conditions.
    let mu_long1 = cp.mu_greater(real_body_len(b1.o, b1.c), &cp.long_body, 2);
    let mu_doji2 = cp.mu_less(real_body_len(b2.o, b2.c), &cp.doji_body, 1);

    let confidence = fuzzy::t_product_all(&[mu_long1, mu_doji2]);
    let direction: f64 = if color1 == -1 { 1.0 } else { -1.0 };
    direction * confidence * 100.0
}
