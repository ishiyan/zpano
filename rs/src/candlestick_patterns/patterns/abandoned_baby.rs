//! Abandoned Baby pattern (3-candle reversal).

use crate::candlestick_patterns::CandlestickPatterns;
use crate::candlestick_patterns::core::{is_black, is_high_low_gap_down, is_high_low_gap_up, is_white, real_body_len};
use crate::fuzzy;

const ABANDONED_BABY_PENETRATION_FACTOR: f64 = 0.3;

/// Abandoned Baby: a three-candle reversal pattern.
///
/// Must have:
/// - first candle: long real body,
/// - second candle: doji,
/// - third candle: real body longer than short, opposite color to 1st,
/// closes well within 1st body,
/// - upside/downside gap between 1st and doji (shadows don't touch),
/// - downside/upside gap between doji and 3rd (shadows don't touch).
///
/// Category C: both branches evaluated, return stronger signal.
///
/// Returns:
/// Continuous float in [-100, +100].
pub fn abandoned_baby(cp: &CandlestickPatterns) -> f64 {
    if !cp.enough(3, &[&cp.long_body, &cp.doji_body, &cp.short_body]) {
        return 0.0;
    }

    let b1 = cp.bar(3);
    let b2 = cp.bar(2);
    let b3 = cp.bar(1);

    // Shared fuzzy conditions: 1st long, 2nd doji, 3rd > short.
    let mu_long1 = cp.mu_greater(real_body_len(b1.o, b1.c), &cp.long_body, 3);
    let mu_doji2 = cp.mu_less(real_body_len(b2.o, b2.c), &cp.doji_body, 2);
    let mu_short3 = cp.mu_greater(real_body_len(b3.o, b3.c), &cp.short_body, 1);

    let penetration = ABANDONED_BABY_PENETRATION_FACTOR;

    // Bearish: white-doji-black, gap up then gap down.
    let mut bear_signal = 0.0;
    if is_white(b1.o, b1.c) && is_black(b3.o, b3.c) {
        if is_high_low_gap_up(b1.h, b2.l) && is_high_low_gap_down(b2.l, b3.h) {
            let rb1 = real_body_len(b1.o, b1.c);
            let pen_threshold = b1.c - rb1 * penetration;
            let pen_width = if rb1 > 0.0 { cp.fuzz_ratio * rb1 } else { 0.0 };
            let mu_pen = cp.mu_lt_raw(b3.c, pen_threshold, pen_width);
            let conf_bear = fuzzy::t_product_all(&[mu_long1, mu_doji2, mu_short3, mu_pen]);
            bear_signal = -conf_bear * 100.0;
        }
    }

    // Bullish: black-doji-white, gap down then gap up.
    let mut bull_signal = 0.0;
    if is_black(b1.o, b1.c) && is_white(b3.o, b3.c) {
        if is_high_low_gap_down(b1.l, b2.h) && is_high_low_gap_up(b2.h, b3.l) {
            let rb1 = real_body_len(b1.o, b1.c);
            let pen_threshold = b1.c + rb1 * penetration;
            let pen_width = if rb1 > 0.0 { cp.fuzz_ratio * rb1 } else { 0.0 };
            let mu_pen = cp.mu_gt_raw(b3.c, pen_threshold, pen_width);
            let conf_bull = fuzzy::t_product_all(&[mu_long1, mu_doji2, mu_short3, mu_pen]);
            bull_signal = conf_bull * 100.0;
        }
    }

    if bull_signal.abs() >= bear_signal.abs() { bull_signal } else { bear_signal }
}
