//! Two Crows pattern (3-candle bearish).

use crate::candlestick_patterns::CandlestickPatterns;
use crate::candlestick_patterns::core::{is_black, is_real_body_gap_up, is_white, real_body_len};

/// Two Crows: a three-candle bearish pattern.
///
/// Must have:
/// - first candle: long white,
/// - second candle: black, gaps up (real body gap up from the first),
/// - third candle: black, opens within the second candle's real body,
/// closes within the first candle's real body.
///
/// The meaning of "long" is specified with self._long_body.
///
/// Category A: always bearish (continuous).
///
/// Returns:
/// Continuous float in [-100, 0].  Always bearish.
pub fn two_crows(cp: &CandlestickPatterns) -> f64 {
    if !cp.enough(3, &[&cp.long_body]) {
        return 0.0;
    }

    let b1 = cp.bar(3);
    let b2 = cp.bar(2);
    let b3 = cp.bar(1);

    // Crisp gates: colors.
    // Crisp gates: color checks.
    if !(is_white(b1.o, b1.c) && is_black(b2.o, b2.c) && is_black(b3.o, b3.c)) { return 0.0; }

    // Crisp: gap up.
    if !is_real_body_gap_up(b1.o, b1.c, b2.o, b2.c) { return 0.0; }

    // Crisp: third opens within second body (o3 < o2 and o3 > c2).
    if !(b3.o < b2.o && b3.o > b2.c) { return 0.0; }

    // Crisp: third closes within first body (c3 > o1 and c3 < c1).
    if !(b3.c > b1.o && b3.c < b1.c) { return 0.0; }

    // Fuzzy: first candle is long.
    let mu_long1 = cp.mu_greater(real_body_len(b1.o, b1.c), &cp.long_body, 3);

    let confidence = mu_long1;

    -confidence * 100.0
}
