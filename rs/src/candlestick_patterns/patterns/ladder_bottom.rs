//! Ladder Bottom pattern (5-candle bullish).

use crate::candlestick_patterns::CandlestickPatterns;
use crate::candlestick_patterns::core::{is_black, is_white, upper_shadow};

/// Ladder Bottom: a five-candle bullish pattern.
///
/// Must have:
/// - first three candles: descending black candles (each closes lower),
/// - fourth candle: black with a long upper shadow,
/// - fifth candle: white, opens above the fourth candle's real body,
/// closes above the fourth candle's high.
///
/// The meaning of "long" for shadows is specified with self._long_shadow.
///
/// Category A: always bullish (continuous).
///
/// Returns:
/// Continuous float in [0, 100].  Always bullish.
pub fn ladder_bottom(cp: &CandlestickPatterns) -> f64 {
    if !cp.enough(5, &[&cp.very_short_shadow]) {
        return 0.0;
    }

    let b1 = cp.bar(5);
    let b2 = cp.bar(4);
    let b3 = cp.bar(3);
    let b4 = cp.bar(2);
    let b5 = cp.bar(1);

    // Crisp gates: colors.
    if !(is_black(b1.o, b1.c) && is_black(b2.o, b2.c)
        && is_black(b3.o, b3.c) && is_black(b4.o, b4.c)
        && is_white(b5.o, b5.c)) {
        return 0.0;
    }
    // Crisp: three descending opens and closes.
    if !(b1.o > b2.o && b2.o > b3.o && b1.c > b2.c && b2.c > b3.c) {
        return 0.0;
    }
    // Crisp: fifth opens above fourth's open, closes above fourth's high.
    if !(b5.o > b4.o && b5.c > b4.h) {
        return 0.0;
    }

    // Fuzzy: fourth candle has upper shadow > very short avg.
    let mu_us4 = cp.mu_greater(upper_shadow(b4.o, b4.h, b4.c), &cp.very_short_shadow, 2);
    mu_us4 * 100.0
}
