//! Three Inside Up/Down pattern (3-candle reversal).

use crate::candlestick_patterns::CandlestickPatterns;
use crate::candlestick_patterns::core::{is_black, is_white, real_body_len};
use crate::fuzzy;

/// Three Inside Up/Down: a three-candle reversal pattern.
///
/// Three Inside Up (bullish):
/// - first candle: long black,
/// - second candle: short, engulfed by the first candle's real body,
/// - third candle: white, closes above the first candle's open.
///
/// Three Inside Down (bearish):
/// - first candle: long white,
/// - second candle: short, engulfed by the first candle's real body,
/// - third candle: black, closes below the first candle's open.
///
/// The meaning of "long" is specified with self._long_body.
/// The meaning of "short" is specified with self._short_body.
///
/// Category C: both branches evaluated, return stronger signal.
///
/// Returns:
/// Continuous float in [-100, +100].
pub fn three_inside(cp: &CandlestickPatterns) -> f64 {
    if !cp.enough(3, &[&cp.long_body, &cp.short_body]) {
        return 0.0;
    }

    let b1 = cp.bar(3);
    let b2 = cp.bar(2);
    let b3 = cp.bar(1);

    // Shared fuzzy conditions.
    let mu_long1 = cp.mu_greater(real_body_len(b1.o, b1.c), &cp.long_body, 3);
    let mu_short2 = cp.mu_less(real_body_len(b2.o, b2.c), &cp.short_body, 2);

    // Fuzzy containment: 1st body encloses 2nd body.
    let eq_avg = cp.avg_cs(&cp.equal, 2);
    let eq_width = if eq_avg > 0.0 { cp.fuzz_ratio * eq_avg } else { 0.0 };
    let mu_enc_upper = cp.mu_ge_raw(f64::max(b1.o, b1.c), f64::max(b2.o, b2.c), eq_width);
    let mu_enc_lower = cp.mu_lt_raw(f64::min(b1.o, b1.c), f64::min(b2.o, b2.c), eq_width);

    // Three Inside Up: long black, short engulfed, white closes above 1st open.
    let mut bull_signal = 0.0;
    if is_black(b1.o, b1.c) && is_white(b3.o, b3.c) {
        let rb1 = real_body_len(b1.o, b1.c);
        let width = if rb1 > 0.0 { cp.fuzz_ratio * rb1 } else { 0.0 };
        let mu_close_above = cp.mu_gt_raw(b3.c, b1.o, width);
        let conf = fuzzy::t_product_all(&[mu_long1, mu_short2, mu_enc_upper, mu_enc_lower, mu_close_above]);
        bull_signal = conf * 100.0;
    }

    // Three Inside Down: long white, short engulfed, black closes below 1st open.
    let mut bear_signal = 0.0;
    if is_white(b1.o, b1.c) && is_black(b3.o, b3.c) {
        let rb1 = real_body_len(b1.o, b1.c);
        let width = if rb1 > 0.0 { cp.fuzz_ratio * rb1 } else { 0.0 };
        let mu_close_below = cp.mu_lt_raw(b3.c, b1.o, width);
        let conf = fuzzy::t_product_all(&[mu_long1, mu_short2, mu_enc_upper, mu_enc_lower, mu_close_below]);
        bear_signal = -conf * 100.0;
    }

    if bull_signal.abs() >= bear_signal.abs() { bull_signal } else { bear_signal }
}
