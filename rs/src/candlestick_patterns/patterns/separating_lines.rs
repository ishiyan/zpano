//! Separating Lines pattern (2-candle continuation).

use crate::candlestick_patterns::CandlestickPatterns;
use crate::candlestick_patterns::core::{lower_shadow, real_body_len, upper_shadow};
use crate::fuzzy;

/// Separating Lines: a two-candle continuation pattern.
///
/// Opposite colors with the same open. The second candle is a belt hold
/// (long body with no shadow on the opening side).
///
/// - bullish: first candle is black, second is white with same open,
/// long body, very short lower shadow,
/// - bearish: first candle is white, second is black with same open,
/// long body, very short upper shadow.
///
/// The meaning of "long" is specified with self._long_body.
/// The meaning of "very short" for shadows is specified with
/// self._very_short_shadow.
/// The meaning of "equal" is specified with self._equal.
///
/// Category C: both branches evaluated, return stronger signal.
///
/// Returns:
/// Continuous float in [-100, +100].
pub fn separating_lines(cp: &CandlestickPatterns) -> f64 {
    if !cp.enough(2, &[&cp.long_body, &cp.very_short_shadow, &cp.equal]) {
        return 0.0;
    }

    let b1 = cp.bar(2);
    let b2 = cp.bar(1);

    // Opposite colors -- crisp gate.
    let color1: i32 = if b1.c < b1.o { -1 } else { 1 };
    let color2: i32 = if b2.c < b2.o { -1 } else { 1 };
    if color1 == color2 { return 0.0; }

    // Opens near equal -- fuzzy (crisp was abs(o2-o1) <= eq).
    let mu_eq = cp.mu_less((b2.o - b1.o).abs(), &cp.equal, 2);

    // Long body on 2nd candle -- fuzzy.
    let mu_long = cp.mu_greater(real_body_len(b2.o, b2.c), &cp.long_body, 1);

    // Bullish: white belt hold (very short lower shadow).
    let mut bull_signal = 0.0;
    if color2 == 1 {
        let mu_vs = cp.mu_less(lower_shadow(b2.o, b2.l, b2.c), &cp.very_short_shadow, 1);
        let conf = fuzzy::t_product_all(&[mu_eq, mu_long, mu_vs]);
        bull_signal = conf * 100.0;
    }

    // Bearish: black belt hold (very short upper shadow).
    let mut bear_signal = 0.0;
    if color2 == -1 {
        let mu_vs = cp.mu_less(upper_shadow(b2.o, b2.h, b2.c), &cp.very_short_shadow, 1);
        let conf = fuzzy::t_product_all(&[mu_eq, mu_long, mu_vs]);
        bear_signal = -conf * 100.0;
    }

    if bull_signal.abs() >= bear_signal.abs() { bull_signal } else { bear_signal }
}
