//! Closing Marubozu pattern (1-candle).

use crate::candlestick_patterns::CandlestickPatterns;
use crate::candlestick_patterns::core::{is_white, lower_shadow, real_body_len, upper_shadow};
use crate::fuzzy;

/// Closing Marubozu: a one-candle pattern.
///
/// A long candle with a very short shadow on the closing side:
/// - bullish (white): very short upper shadow,
/// - bearish (black): very short lower shadow.
///
/// The meaning of "long" is specified with self._long_body.
/// The meaning of "very short" for shadows is specified with
/// self._very_short_shadow.
///
/// Category C: both branches evaluated, return stronger signal.
///
/// Returns:
/// Continuous float in [-100, +100].
pub fn closing_marubozu(cp: &CandlestickPatterns) -> f64 {
    if !cp.enough(1, &[&cp.long_body, &cp.very_short_shadow]) {
        return 0.0;
    }

    let b = cp.bar(1);
    let mu_long = cp.mu_greater(real_body_len(b.o, b.c), &cp.long_body, 1);

    // Bullish: white + very short upper shadow.
    let mut bull_signal = 0.0;
    if is_white(b.o, b.c) {
        let mu_vs = cp.mu_less(upper_shadow(b.o, b.h, b.c), &cp.very_short_shadow, 1);
        let conf = fuzzy::t_product_all(&[mu_long, mu_vs]);
        bull_signal = conf * 100.0;
    }

    // Bearish: black (not white) + very short lower shadow.
    let mut bear_signal = 0.0;
    if !is_white(b.o, b.c) {
        let mu_vs = cp.mu_less(lower_shadow(b.o, b.l, b.c), &cp.very_short_shadow, 1);
        let conf = fuzzy::t_product_all(&[mu_long, mu_vs]);
        bear_signal = -conf * 100.0;
    }

    if bull_signal.abs() >= bear_signal.abs() { bull_signal } else { bear_signal }
}
