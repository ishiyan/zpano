//! Upside Gap Two Crows pattern (3-candle bearish).

use crate::candlestick_patterns::CandlestickPatterns;
use crate::candlestick_patterns::core::{is_black, is_real_body_gap_up, is_white, real_body_len};
use crate::fuzzy;

/// Upside Gap Two Crows: a three-candle bearish pattern.
///
/// Must have:
/// - first candle: long white,
/// - second candle: small black that gaps up from the first,
/// - third candle: black that engulfs the second candle's body and
/// closes above the first candle's close (gap not filled).
///
/// The meaning of "long" is specified with self._long_body.
/// The meaning of "short" is specified with self._short_body.
///
/// Category A: always bearish (continuous).
///
/// Returns:
/// Continuous float in [-100, 0].  Always bearish.
pub fn upside_gap_two_crows(cp: &CandlestickPatterns) -> f64 {
    if !cp.enough(3, &[&cp.long_body, &cp.short_body]) {
        return 0.0;
    }

    let b1 = cp.bar(3);
    let b2 = cp.bar(2);
    let b3 = cp.bar(1);

    // Crisp gates: colors.
    if !(is_white(b1.o, b1.c) && is_black(b2.o, b2.c) && is_black(b3.o, b3.c)) { return 0.0; }

    // Crisp: gap up from first to second.
    if !is_real_body_gap_up(b1.o, b1.c, b2.o, b2.c) { return 0.0; }

    // Crisp: third engulfs second (o3 > o2 and c3 < c2) and closes above c1.
    if !(b3.o > b2.o && b3.c < b2.c && b3.c > b1.c) { return 0.0; }

    // Fuzzy: first candle is long.
    let mu_long1 = cp.mu_greater(real_body_len(b1.o, b1.c), &cp.long_body, 3);

    // Fuzzy: second candle is short.
    let mu_short2 = cp.mu_less(real_body_len(b2.o, b2.c), &cp.short_body, 2);

    let confidence = fuzzy::t_product_all(&[mu_long1, mu_short2]);

    -confidence * 100.0
}
