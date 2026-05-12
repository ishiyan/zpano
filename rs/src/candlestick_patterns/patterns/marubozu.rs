//! Marubozu pattern (1-candle).

use crate::candlestick_patterns::CandlestickPatterns;
use crate::candlestick_patterns::core::{is_white, lower_shadow, real_body_len, upper_shadow};
use crate::fuzzy;

/// Marubozu: a one-candle pattern.
///
/// Must have:
/// - long real body,
/// - very short upper shadow,
/// - very short lower shadow.
///
/// The meaning of "long" is specified with self._long_body.
/// The meaning of "very short" for shadows is specified with
/// self._very_short_shadow.
///
// Crisp direction from color.
/// Category B: direction from candle color.
///
/// Returns:
/// Continuous float in [-100, +100].
pub fn marubozu(cp: &CandlestickPatterns) -> f64 {
    if !cp.enough(1, &[&cp.long_body, &cp.very_short_shadow]) {
        return 0.0;
    }

    let b = cp.bar(1);
    // Fuzzy: long body, very short shadows.
    let mu_long = cp.mu_greater(real_body_len(b.o, b.c), &cp.long_body, 1);
    let mu_us = cp.mu_less(upper_shadow(b.o, b.h, b.c), &cp.very_short_shadow, 1);
    let mu_ls = cp.mu_less(lower_shadow(b.o, b.l, b.c), &cp.very_short_shadow, 1);

    let confidence = fuzzy::t_product_all(&[mu_long, mu_us, mu_ls]);
    // Crisp direction from color.
    let direction: i32 = if !is_white(b.o, b.c) { -1 } else { 1 };
    (direction as f64) * confidence * 100.0
}
