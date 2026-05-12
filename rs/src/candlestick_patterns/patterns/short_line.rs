//! Short Line pattern (1-candle).

use crate::candlestick_patterns::CandlestickPatterns;
use crate::candlestick_patterns::core::{is_black, is_white, lower_shadow, real_body_len, upper_shadow};
use crate::fuzzy;

/// Short Line: a one-candle pattern.
///
/// A candle with a short body, short upper shadow, and short lower shadow.
///
/// The meaning of "short" for body is specified with self._short_body.
/// The meaning of "short" for shadows is specified with self._short_shadow.
///
/// Category C: color determines sign.
///
/// Returns:
/// Continuous float in [-100, +100].
pub fn short_line(cp: &CandlestickPatterns) -> f64 {
    if !cp.enough(1, &[&cp.short_body, &cp.short_shadow]) {
        return 0.0;
    }

    let b = cp.bar(1);

    let mu_short_body = cp.mu_less(real_body_len(b.o, b.c), &cp.short_body, 1);
    let mu_short_us = cp.mu_less(upper_shadow(b.o, b.h, b.c), &cp.short_shadow, 1);
    let mu_short_ls = cp.mu_less(lower_shadow(b.o, b.l, b.c), &cp.short_shadow, 1);

    let confidence = fuzzy::t_product_all(&[mu_short_body, mu_short_us, mu_short_ls]);

    if is_white(b.o, b.c) {
        return confidence * 100.0;
    }
    if is_black(b.o, b.c) {
        return -confidence * 100.0;
    }
    0.0
}
