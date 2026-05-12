//! Tasuki Gap pattern (3-candle continuation).

use crate::candlestick_patterns::CandlestickPatterns;
use crate::candlestick_patterns::core::{is_black, is_real_body_gap_down, is_real_body_gap_up, is_white, real_body_len};
use crate::fuzzy;

/// Tasuki Gap: a three-candle continuation pattern.
///
/// Upside Tasuki Gap (bullish):
/// - real-body gap up between 1st and 2nd candles,
/// - 2nd candle: white,
/// - 3rd candle: black, opens within 2nd white body, closes below 2nd
/// open but above 1st candle's real body top (inside the gap),
/// - 2nd and 3rd have near-equal body sizes.
///
/// Downside Tasuki Gap (bearish):
/// - real-body gap down between 1st and 2nd candles,
/// - 2nd candle: black,
/// - 3rd candle: white, opens within 2nd black body, closes above 2nd
/// open but below 1st candle's real body bottom (inside the gap),
/// - 2nd and 3rd have near-equal body sizes.
///
/// Category C: both branches evaluated, return stronger signal.
///
/// Returns:
/// Continuous float in [-100, +100].
pub fn tasuki_gap(cp: &CandlestickPatterns) -> f64 {
    if !cp.enough(3, &[&cp.near]) {
        return 0.0;
    }

    let b1 = cp.bar(3);
    let b2 = cp.bar(2);
    let b3 = cp.bar(1);

    // Upside Tasuki Gap (bullish).
    let mut bull_signal = 0.0;
    if is_real_body_gap_up(b1.o, b1.c, b2.o, b2.c)
        && is_white(b2.o, b2.c) && is_black(b3.o, b3.c)
    {
        let rb2 = real_body_len(b2.o, b2.c);
        let rb3 = real_body_len(b3.o, b3.c);
        let width = if rb2 > 0.0 { cp.fuzz_ratio * rb2 } else { 0.0 };
        // o3 within 2nd body: o3 < c2 and o3 > o2
        let mu_o3_lt_c2 = cp.mu_lt_raw(b3.o, b2.c, width);
        let mu_o3_gt_o2 = cp.mu_gt_raw(b3.o, b2.o, width);
        // c3 below o2
        let mu_c3_lt_o2 = cp.mu_lt_raw(b3.c, b2.o, width);
        // c3 above 1st body top (inside gap)
        let body1_top = f64::max(b1.c, b1.o);
        let mu_c3_gt_top1 = cp.mu_gt_raw(b3.c, body1_top, width);
        // near-equal bodies
        let mu_near = cp.mu_less((rb2 - rb3).abs(), &cp.near, 2);
        let conf = fuzzy::t_product_all(&[mu_o3_lt_c2, mu_o3_gt_o2, mu_c3_lt_o2, mu_c3_gt_top1, mu_near]);
        bull_signal = conf * 100.0;
    }

    // Downside Tasuki Gap (bearish).
    let mut bear_signal = 0.0;
    if is_real_body_gap_down(b1.o, b1.c, b2.o, b2.c)
        && is_black(b2.o, b2.c) && is_white(b3.o, b3.c)
    {
        let rb2 = real_body_len(b2.o, b2.c);
        let rb3 = real_body_len(b3.o, b3.c);
        let width = if rb2 > 0.0 { cp.fuzz_ratio * rb2 } else { 0.0 };
        // o3 within 2nd body: o3 < o2 and o3 > c2
        let mu_o3_lt_o2 = cp.mu_lt_raw(b3.o, b2.o, width);
        let mu_o3_gt_c2 = cp.mu_gt_raw(b3.o, b2.c, width);
        // c3 above o2
        let mu_c3_gt_o2 = cp.mu_gt_raw(b3.c, b2.o, width);
        // c3 below 1st body bottom (inside gap)
        let body1_bot = f64::min(b1.c, b1.o);
        let mu_c3_lt_bot1 = cp.mu_lt_raw(b3.c, body1_bot, width);
        // near-equal bodies
        let mu_near = cp.mu_less((rb2 - rb3).abs(), &cp.near, 2);
        let conf = fuzzy::t_product_all(&[mu_o3_lt_o2, mu_o3_gt_c2, mu_c3_gt_o2, mu_c3_lt_bot1, mu_near]);
        bear_signal = -conf * 100.0;
    }

    if bull_signal.abs() >= bear_signal.abs() { bull_signal } else { bear_signal }
}
