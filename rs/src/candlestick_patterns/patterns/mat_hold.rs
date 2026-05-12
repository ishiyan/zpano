//! Mat Hold pattern (5-candle bullish).

use crate::candlestick_patterns::CandlestickPatterns;
use crate::candlestick_patterns::core::{is_black, is_real_body_gap_up, is_white, real_body_len};
use crate::fuzzy;

const MAT_HOLD_PENETRATION_FACTOR: f64 = 0.5;

/// Mat Hold: a five-candle bullish continuation pattern.
///
/// Must have:
/// - first candle: long white,
/// - second candle: small, black, gaps up from first,
/// - third and fourth candles: small,
/// - reaction candles (2-4) are falling, hold within first body
/// (penetration check),
/// - fifth candle: white, opens above prior close, closes above
/// highest high of reaction candles.
///
/// Category A: always bullish (continuous).
///
/// Returns:
/// Continuous float in [0, 100].  Always bullish.
pub fn mat_hold(cp: &CandlestickPatterns) -> f64 {
    if !cp.enough(5, &[&cp.long_body, &cp.short_body]) {
        return 0.0;
    }

    let b1 = cp.bar(5);
    let b2 = cp.bar(4);
    let b3 = cp.bar(3);
    let b4 = cp.bar(2);
    let b5 = cp.bar(1);

    // Crisp gates: colors.
    if !(is_white(b1.o, b1.c) && is_black(b2.o, b2.c) && is_white(b5.o, b5.c)) {
        return 0.0;
    }
    // Crisp: gap up from 1st to 2nd.
    if !is_real_body_gap_up(b1.o, b1.c, b2.o, b2.c) { return 0.0; }
    // Crisp: 3rd to 4th hold within 1st range.
    if !(f64::min(b3.o, b3.c) < b1.c && f64::min(b4.o, b4.c) < b1.c) { return 0.0; }
    // Crisp: reaction days don't penetrate first body too much.
    let rb1 = real_body_len(b1.o, b1.c);
    if !(f64::min(b3.o, b3.c) > b1.c - rb1 * MAT_HOLD_PENETRATION_FACTOR
        && f64::min(b4.o, b4.c) > b1.c - rb1 * MAT_HOLD_PENETRATION_FACTOR) {
        return 0.0;
    }
    // Crisp: 2nd to 4th are falling.
    if !(f64::max(b3.o, b3.c) < b2.o && f64::max(b4.o, b4.c) < f64::max(b3.o, b3.c)) {
        return 0.0;
    }
    // Crisp: 5th opens above prior close.
    if !(b5.o > b4.c) { return 0.0; }
    // Crisp: 5th closes above highest high of reaction candles.
    if !(b5.c > f64::max(b2.h, f64::max(b3.h, b4.h))) { return 0.0; }

    // Fuzzy: first candle long.
    let mu_long1 = cp.mu_greater(rb1, &cp.long_body, 5);
    // Fuzzy: 2nd, 3rd, 4th short.
    let mu_short2 = cp.mu_less(real_body_len(b2.o, b2.c), &cp.short_body, 4);
    let mu_short3 = cp.mu_less(real_body_len(b3.o, b3.c), &cp.short_body, 3);
    let mu_short4 = cp.mu_less(real_body_len(b4.o, b4.c), &cp.short_body, 2);

    let confidence = fuzzy::t_product_all(&[mu_long1, mu_short2, mu_short3, mu_short4]);
    confidence * 100.0
}
