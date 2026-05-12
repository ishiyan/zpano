//! Rickshaw Man pattern (1-candle).

use crate::candlestick_patterns::CandlestickPatterns;
use crate::candlestick_patterns::core::{lower_shadow, real_body_len, upper_shadow};
use crate::fuzzy;

/// Rickshaw Man: a one-candle doji pattern.
///
/// Must have:
/// - doji body (very small real body),
/// - two long shadows,
/// - body near the midpoint of the high-low range.
///
/// Returns:
/// Continuous float in [0, 100].  Higher = stronger signal.
pub fn rickshaw_man(cp: &CandlestickPatterns) -> f64 {
    if !cp.enough(1, &[&cp.doji_body, &cp.long_shadow, &cp.near]) {
        return 0.0;
    }

    let b = cp.bar(1);

    let hl_range = b.h - b.l;
    let near_avg = cp.avg_cs(&cp.near, 1);
    let near_width = if near_avg > 0.0 { cp.fuzz_ratio * near_avg } else { 0.0 };

    let mu_doji = cp.mu_less(real_body_len(b.o, b.c), &cp.doji_body, 1);
    let mu_long_us = cp.mu_greater(upper_shadow(b.o, b.h, b.c), &cp.long_shadow, 1);
    let mu_long_ls = cp.mu_greater(lower_shadow(b.o, b.l, b.c), &cp.long_shadow, 1);
    let midpoint = b.l + hl_range / 2.0;
    let mu_near_mid_lo = cp.mu_lt_raw(f64::min(b.o, b.c), midpoint + near_avg, near_width);
    let mu_near_mid_hi = cp.mu_ge_raw(f64::max(b.o, b.c), midpoint - near_avg, near_width);

    let confidence = fuzzy::t_product_all(&[mu_doji, mu_long_us, mu_long_ls, mu_near_mid_lo, mu_near_mid_hi]);
    confidence * 100.0
}
