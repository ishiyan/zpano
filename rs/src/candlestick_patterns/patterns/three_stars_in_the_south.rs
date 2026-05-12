//! Three Stars In The South pattern (3-candle bullish).

use crate::candlestick_patterns::CandlestickPatterns;
use crate::candlestick_patterns::core::{is_black, lower_shadow, real_body_len, upper_shadow};
use crate::fuzzy;

/// Three Stars In The South: a three-candle bullish pattern.
///
/// Must have:
/// - all three candles are black,
/// - first candle: long body with long lower shadow,
/// - second candle: smaller body, opens within or above prior range,
/// trades lower but its low does not go below the first candle's low,
/// - third candle: small marubozu (very short shadows) engulfed by the
/// second candle's range.
///
/// The meaning of "long" is specified with self._long_body.
/// The meaning of "short" is specified with self._short_body.
/// The meaning of "long" for shadows is specified with self._long_shadow.
/// The meaning of "very short" for shadows is specified with
/// self._very_short_shadow.
///
/// Category A: always bullish (continuous).
///
/// Returns:
/// Continuous float in [0, 100].  Always bullish.
pub fn three_stars_in_the_south(cp: &CandlestickPatterns) -> f64 {
    if !cp.enough(3, &[&cp.long_body, &cp.short_body, &cp.long_shadow, &cp.very_short_shadow]) {
        return 0.0;
    }

    let b1 = cp.bar(3);
    let b2 = cp.bar(2);
    let b3 = cp.bar(1);

    // Crisp gates: all black.
    if !(is_black(b1.o, b1.c) && is_black(b2.o, b2.c) && is_black(b3.o, b3.c)) {
        return 0.0;
    }

    let rb1 = real_body_len(b1.o, b1.c);
    let rb2 = real_body_len(b2.o, b2.c);

    // Crisp: second body smaller than first.
    if !(rb2 < rb1) { return 0.0; }

    // Crisp: second opens within or above prior range, low not below first's low.
    if !(b2.o <= b1.h && b2.o >= b1.l && b2.l >= b1.l) { return 0.0; }

    // Crisp: third engulfed by second's range.
    if !(b3.h <= b2.h && b3.l >= b2.l) { return 0.0; }

    // Fuzzy: first candle long body.
    let mu_long1 = cp.mu_greater(rb1, &cp.long_body, 3);

    // Fuzzy: first candle long lower shadow.
    let mu_ls1 = cp.mu_greater(lower_shadow(b1.o, b1.l, b1.c), &cp.long_shadow, 3);

    // Fuzzy: third candle short body.
    let mu_short3 = cp.mu_less(real_body_len(b3.o, b3.c), &cp.short_body, 1);

    // Fuzzy: third candle very short shadows (marubozu).
    let mu_vs_us3 = cp.mu_less(upper_shadow(b3.o, b3.h, b3.c), &cp.very_short_shadow, 1);
    let mu_vs_ls3 = cp.mu_less(lower_shadow(b3.o, b3.l, b3.c), &cp.very_short_shadow, 1);

    let confidence = fuzzy::t_product_all(&[mu_long1, mu_ls1, mu_short3, mu_vs_us3, mu_vs_ls3]);

    confidence * 100.0
}
