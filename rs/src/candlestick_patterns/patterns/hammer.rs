//! Hammer pattern (2-candle bullish).

use crate::candlestick_patterns::CandlestickPatterns;
use crate::candlestick_patterns::core::{lower_shadow, real_body_len, upper_shadow};
use crate::fuzzy;

/// Hammer: a two-candle bullish reversal pattern.
///
/// Must have:
/// - small real body,
/// - long lower shadow,
/// - no or very short upper shadow,
/// - body is below or near the lows of the previous candle.
///
/// Returns:
/// Continuous float in [0, 100].  Higher = stronger hammer signal.
pub fn hammer(cp: &CandlestickPatterns) -> f64 {
    if !cp.enough(2, &[&cp.short_body, &cp.long_shadow, &cp.very_short_shadow, &cp.near]) {
        return 0.0;
    }

    let b1 = cp.bar(2);
    let b2 = cp.bar(1);

    let near_avg = cp.avg_cs(&cp.near, 2);
    let near_width = if near_avg > 0.0 { cp.fuzz_ratio * near_avg } else { 0.0 };

    let mu_short = cp.mu_less(real_body_len(b2.o, b2.c), &cp.short_body, 1);
    let mu_long_ls = cp.mu_greater(lower_shadow(b2.o, b2.l, b2.c), &cp.long_shadow, 1);
    let mu_short_us = cp.mu_less(upper_shadow(b2.o, b2.h, b2.c), &cp.very_short_shadow, 1);
    let mu_near_low = cp.mu_lt_raw(f64::min(b2.o, b2.c), b1.l + near_avg, near_width);

    let confidence = fuzzy::t_product_all(&[mu_short, mu_long_ls, mu_short_us, mu_near_low]);
    confidence * 100.0
}
