//! Shooting Star pattern (2-candle bearish).

use crate::candlestick_patterns::CandlestickPatterns;
use crate::candlestick_patterns::core::{is_real_body_gap_up, lower_shadow, real_body_len, upper_shadow};
use crate::fuzzy;

/// Shooting Star: a two-candle bearish reversal pattern.
///
/// Must have:
/// - gap up from the previous candle (real body gap up),
/// - small real body,
/// - long upper shadow,
/// - very short lower shadow.
///
/// Returns:
/// Continuous float in [-100, 0].  More negative = stronger signal.
pub fn shooting_star(cp: &CandlestickPatterns) -> f64 {
    if !cp.enough(2, &[&cp.short_body, &cp.long_shadow, &cp.very_short_shadow]) {
        return 0.0;
    }

    let b1 = cp.bar(2);
    let b2 = cp.bar(1);

    if !is_real_body_gap_up(b1.o, b1.c, b2.o, b2.c) {
        return 0.0;
    }

    let mu_short = cp.mu_less(real_body_len(b2.o, b2.c), &cp.short_body, 1);
    let mu_long_us = cp.mu_greater(upper_shadow(b2.o, b2.h, b2.c), &cp.long_shadow, 1);
    let mu_short_ls = cp.mu_less(lower_shadow(b2.o, b2.l, b2.c), &cp.very_short_shadow, 1);

    let confidence = fuzzy::t_product_all(&[mu_short, mu_long_us, mu_short_ls]);
    -confidence * 100.0
}
