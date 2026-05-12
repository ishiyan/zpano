//! Stick Sandwich pattern (3-candle bullish).

use crate::candlestick_patterns::CandlestickPatterns;
use crate::candlestick_patterns::core::{is_black, is_white};

/// Stick Sandwich: a three-candle bullish pattern.
///
/// Must have:
/// - first candle: black,
/// - second candle: white, trades above the first candle's close
/// (low > first close),
/// - third candle: black, close equals the first candle's close.
///
/// The meaning of "equal" is specified with self._equal.
///
/// Category A: always bullish (continuous).
///
/// Returns:
/// Continuous float in [0, 100].  Always bullish.
pub fn stick_sandwich(cp: &CandlestickPatterns) -> f64 {
    if !cp.enough(3, &[&cp.equal]) {
        return 0.0;
    }

    let b1 = cp.bar(3);
    let b2 = cp.bar(2);
    let b3 = cp.bar(1);

    // Crisp gates: colors and gap.
    if !(is_black(b1.o, b1.c) && is_white(b2.o, b2.c) && is_black(b3.o, b3.c) && b2.l > b1.c) {
        return 0.0;
    }

    // Fuzzy: third close equals first close (two-sided band).
    let mu_eq = cp.mu_less((b3.c - b1.c).abs(), &cp.equal, 3);

    let confidence = mu_eq;

    confidence * 100.0
}
