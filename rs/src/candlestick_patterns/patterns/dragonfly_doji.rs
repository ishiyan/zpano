//! Dragonfly Doji pattern (1-candle).

use crate::candlestick_patterns::CandlestickPatterns;
use crate::candlestick_patterns::core::{lower_shadow, real_body_len, upper_shadow};
use crate::fuzzy;

/// Dragonfly Doji: a one-candle pattern.
///
/// Must have:
/// - doji body (very small real body relative to high-low range),
/// - no or very short upper shadow,
/// - lower shadow is not very short.
///
/// Returns:
/// Continuous float in [0, 100].  Higher = stronger signal.
pub fn dragonfly_doji(cp: &CandlestickPatterns) -> f64 {
    if !cp.enough(1, &[&cp.doji_body, &cp.very_short_shadow]) {
        return 0.0;
    }

    let b = cp.bar(1);
    let mu_doji = cp.mu_less(real_body_len(b.o, b.c), &cp.doji_body, 1);
    let mu_short_us = cp.mu_less(upper_shadow(b.o, b.h, b.c), &cp.very_short_shadow, 1);
    let mu_long_ls = cp.mu_greater(lower_shadow(b.o, b.l, b.c), &cp.very_short_shadow, 1);

    let confidence = fuzzy::t_product_all(&[mu_doji, mu_short_us, mu_long_ls]);
    confidence * 100.0
}
