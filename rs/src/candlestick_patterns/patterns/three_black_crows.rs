//! Three Black Crows pattern (4-candle bearish).

use crate::candlestick_patterns::CandlestickPatterns;
use crate::candlestick_patterns::core::{is_black, is_white, lower_shadow};
use crate::fuzzy;

/// Three Black Crows: a four-candle bearish reversal pattern.
///
/// Must have:
/// - preceding candle (oldest) is white,
/// - three consecutive black candles with declining closes,
/// - each opens within the prior black candle's real body,
/// - each has a very short lower shadow,
/// - 1st black closes under the prior white candle's high.
///
/// Category A: always bearish (continuous).
///
/// Returns:
/// Continuous float in [-100, 0].  Always bearish.
pub fn three_black_crows(cp: &CandlestickPatterns) -> f64 {
    if !cp.enough(4, &[&cp.very_short_shadow]) {
        return 0.0;
    }

    let b0 = cp.bar(4); // prior white
    let b1 = cp.bar(3); // 1st black
    let b2 = cp.bar(2); // 2nd black
    let b3 = cp.bar(1); // 3rd black

    // Crisp gates: colors, declining closes, opens within prior body.
    if !is_white(b0.o, b0.c) { return 0.0; }
    if !(is_black(b1.o, b1.c) && is_black(b2.o, b2.c) && is_black(b3.o, b3.c)) { return 0.0; }
    if !(b1.c > b2.c && b2.c > b3.c) { return 0.0; }
    // Opens within prior black body (crisp containment for strict ordering).
    if !(b2.o < b1.o && b2.o > b1.c && b3.o < b2.o && b3.o > b2.c) { return 0.0; }
    // Prior white's high > 1st black's close (crisp).
    if !(b0.h > b1.c) { return 0.0; }

    // Fuzzy: very short lower shadows.
    let mu_ls1 = cp.mu_less(lower_shadow(b1.o, b1.l, b1.c), &cp.very_short_shadow, 3);
    let mu_ls2 = cp.mu_less(lower_shadow(b2.o, b2.l, b2.c), &cp.very_short_shadow, 2);
    let mu_ls3 = cp.mu_less(lower_shadow(b3.o, b3.l, b3.c), &cp.very_short_shadow, 1);

    let confidence = fuzzy::t_product_all(&[mu_ls1, mu_ls2, mu_ls3]);

    -1.0 * confidence * 100.0
}
