//! Concealing Baby Swallow pattern (4-candle bullish).

use crate::candlestick_patterns::CandlestickPatterns;
use crate::candlestick_patterns::core::{is_black, is_real_body_gap_down, lower_shadow, upper_shadow};
use crate::fuzzy;

/// Concealing Baby Swallow: a four-candle bullish pattern.
///
/// Must have:
/// - first candle: black marubozu (very short shadows),
/// - second candle: black marubozu (very short shadows),
/// - third candle: black, opens gapping down, upper shadow extends into
/// the prior real body (upper shadow > very-short avg),
/// - fourth candle: black, completely engulfs the third candle including
/// shadows (strict > / <).
///
/// The meaning of "very short" for shadows is specified with
/// self._very_short_shadow.
///
/// Category A: always bullish (continuous).
///
/// Returns:
/// Continuous float in [0, 100].  Always bullish.
pub fn concealing_baby_swallow(cp: &CandlestickPatterns) -> f64 {
    if !cp.enough(4, &[&cp.very_short_shadow]) {
        return 0.0;
    }

    let b1 = cp.bar(4);
    let b2 = cp.bar(3);
    let b3 = cp.bar(2);
    let b4 = cp.bar(1);

    // Crisp gates: all black.
    if !(is_black(b1.o, b1.c) && is_black(b2.o, b2.c)
        && is_black(b3.o, b3.c) && is_black(b4.o, b4.c)) {
        return 0.0;
    }
    // Crisp: gap down and upper shadow extends into prior body.
    if !(is_real_body_gap_down(b2.o, b2.c, b3.o, b3.c) && b3.h > b2.c) {
        return 0.0;
    }
    // Crisp: fourth engulfs third including shadows (strict).
    if !(b4.h > b3.h && b4.l < b3.l) {
        return 0.0;
    }

    // Fuzzy: first and second are marubozu (very short shadows).
    let mu_ls1 = cp.mu_less(lower_shadow(b1.o, b1.l, b1.c), &cp.very_short_shadow, 4);
    let mu_us1 = cp.mu_less(upper_shadow(b1.o, b1.h, b1.c), &cp.very_short_shadow, 4);
    let mu_ls2 = cp.mu_less(lower_shadow(b2.o, b2.l, b2.c), &cp.very_short_shadow, 3);
    let mu_us2 = cp.mu_less(upper_shadow(b2.o, b2.h, b2.c), &cp.very_short_shadow, 3);
    // Fuzzy: third candle upper shadow > very-short avg.
    let mu_us3_long = cp.mu_greater(upper_shadow(b3.o, b3.h, b3.c), &cp.very_short_shadow, 2);

    let confidence = fuzzy::t_product_all(&[mu_ls1, mu_us1, mu_ls2, mu_us2, mu_us3_long]);
    confidence * 100.0
}
