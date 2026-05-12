//! Dark Cloud Cover pattern (2-candle bearish).

use crate::candlestick_patterns::CandlestickPatterns;
use crate::candlestick_patterns::core::{is_black, is_white, real_body_len};
use crate::fuzzy;

const DARK_CLOUD_COVER_PENETRATION_FACTOR: f64 = 0.5;

/// Dark Cloud Cover: a two-candle bearish reversal pattern.
///
/// Must have:
/// - first candle: long white candle,
/// - second candle: black candle that opens above the prior high and
/// closes well within the first candle's real body (below the midpoint).
///
/// Returns:
/// Continuous float in [-100, 0].  More negative = stronger signal.
pub fn dark_cloud_cover(cp: &CandlestickPatterns) -> f64 {
    if !cp.enough(2, &[&cp.long_body]) {
        return 0.0;
    }

    let b1 = cp.bar(2);
    let b2 = cp.bar(1);

    // Color checks stay crisp.
    if !is_white(b1.o, b1.c) || !is_black(b2.o, b2.c) {
        return 0.0;
    }

    let rb1 = real_body_len(b1.o, b1.c);
    let eq_avg = cp.avg_cs(&cp.equal, 1);
    let eq_width = if eq_avg > 0.0 { cp.fuzz_ratio * eq_avg } else { 0.0 };

    let mu_long = cp.mu_greater(rb1, &cp.long_body, 2);
    let mu_open_above = cp.mu_gt_raw(b2.o, b1.h, eq_width);
    let pen_threshold = b1.c - rb1 * DARK_CLOUD_COVER_PENETRATION_FACTOR;
    let pen_product = rb1 * DARK_CLOUD_COVER_PENETRATION_FACTOR;
    let pen_width = if pen_product > 0.0 { cp.fuzz_ratio * pen_product } else { 0.0 };
    let mu_pen = cp.mu_lt_raw(b2.c, pen_threshold, pen_width);
    let mu_above_open1 = cp.mu_gt_raw(b2.c, b1.o, eq_width);

    let confidence = fuzzy::t_product_all(&[mu_long, mu_open_above, mu_pen, mu_above_open1]);
    -confidence * 100.0
}
