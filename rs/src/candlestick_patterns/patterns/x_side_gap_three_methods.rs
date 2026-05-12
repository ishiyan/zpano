//! Up/Down-side Gap Three Methods pattern (3-candle continuation).

use crate::candlestick_patterns::CandlestickPatterns;
use crate::candlestick_patterns::core::{is_black, is_real_body_gap_down, is_real_body_gap_up, is_white, real_body_len};
use crate::fuzzy;

/// Up/Down-side Gap Three Methods: a three-candle continuation pattern.
///
/// Must have:
/// - first and second candles are the same color with a gap between them,
/// - third candle is opposite color, opens within the second candle's
/// real body and closes within the first candle's real body (fills the
/// gap).
///
/// Upside gap: two white candles with gap up, third is black = bullish.
/// Downside gap: two black candles with gap down, third is white = bearish.
///
/// Category C: both branches evaluated, return stronger signal.
///
/// Returns:
/// Continuous float in [-100, +100].
pub fn x_side_gap_three_methods(cp: &CandlestickPatterns) -> f64 {
    if !cp.enough(3, &[]) {
        return 0.0;
    }

    let b1 = cp.bar(3);
    let b2 = cp.bar(2);
    let b3 = cp.bar(1);

    // Upside gap: two whites gap up, third black fills.
    let mut bull_signal = 0.0;
    if is_white(b1.o, b1.c) && is_white(b2.o, b2.c) && is_black(b3.o, b3.c)
        && is_real_body_gap_up(b1.o, b1.c, b2.o, b2.c)
    {
        let rb2 = real_body_len(b2.o, b2.c);
        let width = if rb2 > 0.0 { cp.fuzz_ratio * rb2 } else { 0.0 };
        // o3 within 2nd body: o3 < c2 and o3 > o2
        let mu_o3_lt_c2 = cp.mu_lt_raw(b3.o, b2.c, width);
        let mu_o3_gt_o2 = cp.mu_gt_raw(b3.o, b2.o, width);
        // c3 within 1st body: c3 > o1 and c3 < c1
        let rb1 = real_body_len(b1.o, b1.c);
        let width1 = if rb1 > 0.0 { cp.fuzz_ratio * rb1 } else { 0.0 };
        let mu_c3_gt_o1 = cp.mu_gt_raw(b3.c, b1.o, width1);
        let mu_c3_lt_c1 = cp.mu_lt_raw(b3.c, b1.c, width1);
        let conf = fuzzy::t_product_all(&[mu_o3_lt_c2, mu_o3_gt_o2, mu_c3_gt_o1, mu_c3_lt_c1]);
        bull_signal = conf * 100.0;
    }

    // Downside gap: two blacks gap down, third white fills.
    let mut bear_signal = 0.0;
    if is_black(b1.o, b1.c) && is_black(b2.o, b2.c) && is_white(b3.o, b3.c)
        && is_real_body_gap_down(b1.o, b1.c, b2.o, b2.c)
    {
        let rb2 = real_body_len(b2.o, b2.c);
        let width = if rb2 > 0.0 { cp.fuzz_ratio * rb2 } else { 0.0 };
        // o3 within 2nd body: o3 > c2 and o3 < o2
        let mu_o3_gt_c2 = cp.mu_gt_raw(b3.o, b2.c, width);
        let mu_o3_lt_o2 = cp.mu_lt_raw(b3.o, b2.o, width);
        // c3 within 1st body: c3 < o1 and c3 > c1
        let rb1 = real_body_len(b1.o, b1.c);
        let width1 = if rb1 > 0.0 { cp.fuzz_ratio * rb1 } else { 0.0 };
        let mu_c3_lt_o1 = cp.mu_lt_raw(b3.c, b1.o, width1);
        let mu_c3_gt_c1 = cp.mu_gt_raw(b3.c, b1.c, width1);
        let conf = fuzzy::t_product_all(&[mu_o3_gt_c2, mu_o3_lt_o2, mu_c3_lt_o1, mu_c3_gt_c1]);
        bear_signal = -conf * 100.0;
    }

    if bull_signal.abs() >= bear_signal.abs() { bull_signal } else { bear_signal }
}
