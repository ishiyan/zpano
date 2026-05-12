//! Three Outside Up/Down pattern (3-candle reversal).

use crate::candlestick_patterns::CandlestickPatterns;
use crate::candlestick_patterns::core::{is_black, is_white, real_body_len};
use crate::fuzzy;

/// Three Outside Up/Down: a three-candle reversal pattern.
///
/// Must have:
/// - first and second candles form an engulfing pattern,
/// - third candle confirms the direction by closing higher (up) or
/// lower (down).
///
/// Three Outside Up: first candle is black, second is white engulfing
/// the first, third closes higher than the second.
///
/// Three Outside Down: first candle is white, second is black engulfing
/// the first, third closes lower than the second.
///
/// Category C: both branches evaluated, return stronger signal.
///
/// Returns:
/// Continuous float in [-100, +100].
pub fn three_outside(cp: &CandlestickPatterns) -> f64 {
    if !cp.enough(3, &[]) {
        return 0.0;
    }

    let b1 = cp.bar(3);
    let b2 = cp.bar(2);
    let b3 = cp.bar(1);

    // Fuzzy engulfment width.
    let eq_avg = cp.avg_cs(&cp.equal, 1);
    let eq_width = if eq_avg > 0.0 { cp.fuzz_ratio * eq_avg } else { 0.0 };

    // Three Outside Up: black + white engulfing + 3rd closes higher.
    let mut bull_signal = 0.0;
    if is_black(b1.o, b1.c) && is_white(b2.o, b2.c) {
        let mu_enc_upper = cp.mu_ge_raw(f64::max(b2.o, b2.c), f64::max(b1.o, b1.c), eq_width);
        let mu_enc_lower = cp.mu_lt_raw(f64::min(b2.o, b2.c), f64::min(b1.o, b1.c), eq_width);
        let rb2 = real_body_len(b2.o, b2.c);
        let width = if rb2 > 0.0 { cp.fuzz_ratio * rb2 } else { 0.0 };
        let mu_close_higher = cp.mu_gt_raw(b3.c, b2.c, width);
        let conf = fuzzy::t_product_all(&[mu_enc_upper, mu_enc_lower, mu_close_higher]);
        bull_signal = conf * 100.0;
    }

    // Three Outside Down: white + black engulfing + 3rd closes lower.
    let mut bear_signal = 0.0;
    if is_white(b1.o, b1.c) && is_black(b2.o, b2.c) {
        let mu_enc_upper = cp.mu_ge_raw(f64::max(b2.o, b2.c), f64::max(b1.o, b1.c), eq_width);
        let mu_enc_lower = cp.mu_lt_raw(f64::min(b2.o, b2.c), f64::min(b1.o, b1.c), eq_width);
        let rb2 = real_body_len(b2.o, b2.c);
        let width = if rb2 > 0.0 { cp.fuzz_ratio * rb2 } else { 0.0 };
        let mu_close_lower = cp.mu_lt_raw(b3.c, b2.c, width);
        let conf = fuzzy::t_product_all(&[mu_enc_upper, mu_enc_lower, mu_close_lower]);
        bear_signal = -conf * 100.0;
    }

    if bull_signal.abs() >= bear_signal.abs() { bull_signal } else { bear_signal }
}
