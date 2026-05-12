//! Piercing pattern (2-candle bullish).

use crate::candlestick_patterns::CandlestickPatterns;
use crate::candlestick_patterns::core::{is_black, is_white, real_body_len};
use crate::fuzzy;

/// Piercing: a two-candle bullish reversal pattern.
///
/// Must have:
/// - first candle: long black,
/// - second candle: long white that opens below the prior low and closes
/// above the midpoint of the first candle's real body but within the body.
///
/// Returns:
/// Continuous float in [0, 100].  Higher = stronger signal.
pub fn piercing(cp: &CandlestickPatterns) -> f64 {
    if !cp.enough(2, &[&cp.long_body]) {
        return 0.0;
    }

    let b1 = cp.bar(2);
    let b2 = cp.bar(1);

    // Color checks stay crisp.
    if !is_black(b1.o, b1.c) || !is_white(b2.o, b2.c) {
        return 0.0;
    }

    let rb1 = real_body_len(b1.o, b1.c);
    let eq_avg = cp.avg_cs(&cp.equal, 1);
    let eq_width = if eq_avg > 0.0 { cp.fuzz_ratio * eq_avg } else { 0.0 };

    let mu_long1 = cp.mu_greater(rb1, &cp.long_body, 2);
    let mu_long2 = cp.mu_greater(real_body_len(b2.o, b2.c), &cp.long_body, 1);
    let mu_open_below = cp.mu_lt_raw(b2.o, b1.l, eq_width);
    let pen_threshold = b1.c + rb1 * 0.5;
    let pen_width = if rb1 > 0.0 { cp.fuzz_ratio * rb1 * 0.5 } else { 0.0 };
    let mu_pen = cp.mu_gt_raw(b2.c, pen_threshold, pen_width);
    let mu_below_open1 = cp.mu_lt_raw(b2.c, b1.o, eq_width);

    let confidence = fuzzy::t_product_all(&[mu_long1, mu_long2, mu_open_below, mu_pen, mu_below_open1]);
    confidence * 100.0
}
