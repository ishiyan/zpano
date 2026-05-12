//! Spinning Top pattern (1-candle).

use crate::candlestick_patterns::CandlestickPatterns;
use crate::candlestick_patterns::core::{is_black, is_white, lower_shadow, real_body_len, upper_shadow};
use crate::fuzzy;

/// Spinning Top: a one-candle pattern.
///
/// A candle with a small body and shadows longer than the body on both sides.
///
/// The meaning of "short" is specified with self._short_body.
///
/// Category C: color determines sign.
///
/// Returns:
/// Continuous float in [-100, +100].
pub fn spinning_top(cp: &CandlestickPatterns) -> f64 {
    if !cp.enough(1, &[&cp.short_body]) {
        return 0.0;
    }

    let b = cp.bar(1);

    let rb = real_body_len(b.o, b.c);

    let mu_short = cp.mu_less(rb, &cp.short_body, 1);

    // Shadows > body: positional comparisons.
    let us = upper_shadow(b.o, b.h, b.c);
    let ls = lower_shadow(b.o, b.l, b.c);
    let width_us = if rb > 0.0 { cp.fuzz_ratio * rb } else { 0.0 };
    let width_ls = if rb > 0.0 { cp.fuzz_ratio * rb } else { 0.0 };
    let mu_us_gt_rb = cp.mu_gt_raw(us, rb, width_us);
    let mu_ls_gt_rb = cp.mu_gt_raw(ls, rb, width_ls);

    let confidence = fuzzy::t_product_all(&[mu_short, mu_us_gt_rb, mu_ls_gt_rb]);

    if is_white(b.o, b.c) {
        return confidence * 100.0;
    }
    if is_black(b.o, b.c) {
        return -confidence * 100.0;
    }
    0.0
}
