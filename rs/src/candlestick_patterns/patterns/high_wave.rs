//! High Wave pattern (1-candle).

use crate::candlestick_patterns::CandlestickPatterns;
use crate::candlestick_patterns::core::{is_white, lower_shadow, real_body_len, upper_shadow};
use crate::fuzzy;

/// High Wave: a one-candle pattern.
///
/// Must have:
/// - short real body,
/// - very long upper shadow,
/// - very long lower shadow.
///
/// The meaning of "short" is specified with self._short_body.
/// The meaning of "very long" (shadow) is specified with self._very_long_shadow.
///
/// Category C: color determines sign.
///
/// Returns:
/// Continuous float in [-100, +100].
pub fn high_wave(cp: &CandlestickPatterns) -> f64 {
    if !cp.enough(1, &[&cp.short_body, &cp.very_long_shadow]) {
        return 0.0;
    }

    let b = cp.bar(1);
    let mu_short = cp.mu_less(real_body_len(b.o, b.c), &cp.short_body, 1);
    let mu_long_us = cp.mu_greater(upper_shadow(b.o, b.h, b.c), &cp.very_long_shadow, 1);
    let mu_long_ls = cp.mu_greater(lower_shadow(b.o, b.l, b.c), &cp.very_long_shadow, 1);

    let confidence = fuzzy::t_product_all(&[mu_short, mu_long_us, mu_long_ls]);
    if is_white(b.o, b.c) {
        confidence * 100.0
    } else {
        -confidence * 100.0
    }
}
