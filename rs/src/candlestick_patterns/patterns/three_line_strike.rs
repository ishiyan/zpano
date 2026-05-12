//! Three-Line Strike pattern (4-candle).

use crate::candlestick_patterns::CandlestickPatterns;
use crate::candlestick_patterns::core::{is_white};
use crate::fuzzy;

/// Three-Line Strike: a four-candle pattern.
///
/// Bullish: three white candles with rising closes, each opening within/near
/// the prior body, 4th black opens above 3rd close and closes below 1st open.
///
/// Bearish: three black candles with falling closes, each opening within/near
/// the prior body, 4th white opens below 3rd close and closes above 1st open.
///
/// Category C: both branches evaluated, return stronger signal.
///
/// Returns:
/// Continuous float in [-100, +100].
pub fn three_line_strike(cp: &CandlestickPatterns) -> f64 {
    if !cp.enough(4, &[&cp.near]) {
        return 0.0;
    }

    let b1 = cp.bar(4);
    let b2 = cp.bar(3);
    let b3 = cp.bar(2);
    let b4 = cp.bar(1);

    // Three same color -- crisp gate.
    let color1: i32 = if !is_white(b1.o, b1.c) { -1 } else { 1 };
    let color2: i32 = if !is_white(b2.o, b2.c) { -1 } else { 1 };
    let color3: i32 = if !is_white(b3.o, b3.c) { -1 } else { 1 };
    let color4: i32 = if !is_white(b4.o, b4.c) { -1 } else { 1 };

    if !(color1 == color2 && color2 == color3 && color4 == -color3) {
        return 0.0;
    }

    // 2nd opens within/near 1st real body -- fuzzy.
    let near4 = cp.avg_cs(&cp.near, 4);
    let near3 = cp.avg_cs(&cp.near, 3);
    let near_width4 = if near4 > 0.0 { cp.fuzz_ratio * near4 } else { 0.0 };
    let near_width3 = if near3 > 0.0 { cp.fuzz_ratio * near3 } else { 0.0 };

    let mu_o2_ge = cp.mu_ge_raw(b2.o, f64::min(b1.o, b1.c) - near4, near_width4);
    let mu_o2_le = cp.mu_lt_raw(b2.o, f64::max(b1.o, b1.c) + near4, near_width4);

    // 3rd opens within/near 2nd real body -- fuzzy.
    let mu_o3_ge = cp.mu_ge_raw(b3.o, f64::min(b2.o, b2.c) - near3, near_width3);
    let mu_o3_le = cp.mu_lt_raw(b3.o, f64::max(b2.o, b2.c) + near3, near_width3);

    // Bullish: three white, rising closes, 4th opens above 3rd close, closes below 1st open.
    let mut bull_signal = 0.0;
    if color3 == 1 && b3.c > b2.c && b2.c > b1.c {
        let rb1 = (b1.c - b1.o).abs();
        let width = if rb1 > 0.0 { cp.fuzz_ratio * rb1 } else { 0.0 };
        let mu_o4_above = cp.mu_gt_raw(b4.o, b3.c, width);
        let mu_c4_below = cp.mu_lt_raw(b4.c, b1.o, width);
        let conf = fuzzy::t_product_all(&[mu_o2_ge, mu_o2_le, mu_o3_ge, mu_o3_le, mu_o4_above, mu_c4_below]);
        bull_signal = conf * 100.0;
    }

    // Bearish: three black, falling closes, 4th opens below 3rd close, closes above 1st open.
    let mut bear_signal = 0.0;
    if color3 == -1 && b3.c < b2.c && b2.c < b1.c {
        let rb1 = (b1.c - b1.o).abs();
        let width = if rb1 > 0.0 { cp.fuzz_ratio * rb1 } else { 0.0 };
        let mu_o4_below = cp.mu_lt_raw(b4.o, b3.c, width);
        let mu_c4_above = cp.mu_gt_raw(b4.c, b1.o, width);
        let conf = fuzzy::t_product_all(&[mu_o2_ge, mu_o2_le, mu_o3_ge, mu_o3_le, mu_o4_below, mu_c4_above]);
        bear_signal = -conf * 100.0;
    }

    if bull_signal.abs() >= bear_signal.abs() { bull_signal } else { bear_signal }
}
