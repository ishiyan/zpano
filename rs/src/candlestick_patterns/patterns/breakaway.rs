//! Breakaway pattern (5-candle reversal).

use crate::candlestick_patterns::CandlestickPatterns;
use crate::candlestick_patterns::core::{is_black, is_real_body_gap_down, is_real_body_gap_up, is_white, real_body_len};
use crate::fuzzy;

/// Breakaway: a five-candle reversal pattern.
///
/// Bullish: first candle is long black, second candle is black gapping down,
/// third and fourth candles have consecutively lower highs and lows, fifth
/// candle is white closing into the gap (between first and second candle's
/// real bodies).
///
/// Bearish: mirror image with colors reversed and gaps reversed.
///
/// Category C: both branches evaluated, return stronger signal.
///
/// Returns:
/// Continuous float in [-100, +100].
pub fn breakaway(cp: &CandlestickPatterns) -> f64 {
    if !cp.enough(5, &[&cp.long_body]) {
        return 0.0;
    }

    let b1 = cp.bar(5);
    let b2 = cp.bar(4);
    let b3 = cp.bar(3);
    let b4 = cp.bar(2);
    let b5 = cp.bar(1);

    // Fuzzy: 1st candle is long.
    // Fuzzy: c5 > o2 and c5 < c1 (closing into the gap).
    let mu_long1 = cp.mu_greater(real_body_len(b1.o, b1.c), &cp.long_body, 5);

    // Bullish breakaway.
    let mut bull_signal = 0.0;
    if is_black(b1.o, b1.c) && is_black(b2.o, b2.c)
        && is_black(b4.o, b4.c) && is_white(b5.o, b5.c)
        && b3.h < b2.h && b3.l < b2.l
        && b4.h < b3.h && b4.l < b3.l
        && is_real_body_gap_down(b1.o, b1.c, b2.o, b2.c)
    {
        let rb1 = real_body_len(b1.o, b1.c);
        let width = if rb1 > 0.0 { cp.fuzz_ratio * rb1 } else { 0.0 };
        let mu_c5_above_o2 = cp.mu_gt_raw(b5.c, b2.o, width);
        let mu_c5_below_c1 = cp.mu_lt_raw(b5.c, b1.c, width);
        let conf = fuzzy::t_product_all(&[mu_long1, mu_c5_above_o2, mu_c5_below_c1]);
        bull_signal = conf * 100.0;
    }

    // Bearish breakaway.
    let mut bear_signal = 0.0;
    if is_white(b1.o, b1.c) && is_white(b2.o, b2.c)
        && is_white(b4.o, b4.c) && is_black(b5.o, b5.c)
        && b3.h > b2.h && b3.l > b2.l
        && b4.h > b3.h && b4.l > b3.l
        && is_real_body_gap_up(b1.o, b1.c, b2.o, b2.c)
    {
        let rb1 = real_body_len(b1.o, b1.c);
        let width = if rb1 > 0.0 { cp.fuzz_ratio * rb1 } else { 0.0 };
        let mu_c5_below_o2 = cp.mu_lt_raw(b5.c, b2.o, width);
        let mu_c5_above_c1 = cp.mu_gt_raw(b5.c, b1.c, width);
        let conf = fuzzy::t_product_all(&[mu_long1, mu_c5_below_o2, mu_c5_above_c1]);
        bear_signal = -conf * 100.0;
    }

    if bull_signal.abs() >= bear_signal.abs() { bull_signal } else { bear_signal }
}
