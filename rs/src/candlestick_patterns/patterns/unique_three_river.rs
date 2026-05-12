//! Unique Three River pattern (3-candle bullish).

use crate::candlestick_patterns::CandlestickPatterns;
use crate::candlestick_patterns::core::{is_black, is_white, real_body_len};
use crate::fuzzy;

/// Unique Three River: a three-candle bullish pattern.
///
/// Must have:
/// - first candle: long black,
/// - second candle: black harami (body within first body) with a lower
/// low than the first candle,
/// - third candle: small white, opens not lower than the second candle's
/// low.
///
/// The meaning of "long" is specified with self._long_body.
/// The meaning of "short" is specified with self._short_body.
///
/// Category A: always bullish (continuous).
///
/// Returns:
/// Continuous float in [0, 100].  Always bullish.
pub fn unique_three_river(cp: &CandlestickPatterns) -> f64 {
    if !cp.enough(3, &[&cp.long_body, &cp.short_body]) {
        return 0.0;
    }

    let b1 = cp.bar(3);
    let b2 = cp.bar(2);
    let b3 = cp.bar(1);

    // Crisp gates: colors.
    if !(is_black(b1.o, b1.c) && is_black(b2.o, b2.c) && is_white(b3.o, b3.c)) { return 0.0; }

    // Crisp: harami body containment and lower low.
    if !(b2.c > b1.c && b2.o <= b1.o && b2.l < b1.l) { return 0.0; }

    // Crisp: third opens not lower than second's low.
    if !(b3.o >= b2.l) { return 0.0; }

    // Fuzzy: first candle is long.
    let mu_long1 = cp.mu_greater(real_body_len(b1.o, b1.c), &cp.long_body, 3);

    // Fuzzy: third candle is short.
    let mu_short3 = cp.mu_less(real_body_len(b3.o, b3.c), &cp.short_body, 1);

    let confidence = fuzzy::t_product_all(&[mu_long1, mu_short3]);

    confidence * 100.0
}
