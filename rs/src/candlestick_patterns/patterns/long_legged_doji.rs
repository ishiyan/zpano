//! Long Legged Doji pattern (1-candle).

use crate::candlestick_patterns::CandlestickPatterns;
use crate::candlestick_patterns::core::{lower_shadow, real_body_len, upper_shadow};
use crate::fuzzy;

/// Long Legged Doji: a one-candle pattern.
///
/// Must have:
/// - doji body (very small real body),
/// - one or both shadows are long.
///
/// Returns:
/// Continuous float in [0, 100].  Higher = stronger signal.
pub fn long_legged_doji(cp: &CandlestickPatterns) -> f64 {
    if !cp.enough(1, &[&cp.doji_body, &cp.long_shadow]) {
        return 0.0;
    }

    let b = cp.bar(1);
    let mu_doji = cp.mu_less(real_body_len(b.o, b.c), &cp.doji_body, 1);
    let mu_long_us = cp.mu_greater(upper_shadow(b.o, b.h, b.c), &cp.long_shadow, 1);
    let mu_long_ls = cp.mu_greater(lower_shadow(b.o, b.l, b.c), &cp.long_shadow, 1);
    let mu_any_long = fuzzy::s_max(mu_long_us, mu_long_ls);

    let confidence = fuzzy::t_product_all(&[mu_doji, mu_any_long]);
    confidence * 100.0
}
