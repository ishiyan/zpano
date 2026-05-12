//! Rising/Falling Three Methods pattern (5-candle continuation).

use crate::candlestick_patterns::CandlestickPatterns;
use crate::candlestick_patterns::core::{is_white, real_body_len};
use crate::fuzzy;

/// Rising/Falling Three Methods: a five-candle continuation pattern.
///
/// Uses TA-Lib logic: opposite-color check via color multiplication,
/// real-body overlap (not full candle containment), sequential closes,
/// 5th opens beyond 4th close.
///
/// Category B: direction from 1st candle color (crisp sign).
///
/// Returns:
/// Continuous float in [-100, +100].
pub fn rising_falling_three_methods(cp: &CandlestickPatterns) -> f64 {
    if !cp.enough(5, &[&cp.long_body, &cp.short_body]) {
        return 0.0;
    }

    let b1 = cp.bar(5);
    let b2 = cp.bar(4);
    let b3 = cp.bar(3);
    let b4 = cp.bar(2);
    let b5 = cp.bar(1);

    // Fuzzy: 1st long, 2nd-4th short, 5th long.
    let mu_long1 = cp.mu_greater(real_body_len(b1.o, b1.c), &cp.long_body, 5);
    let mu_short2 = cp.mu_less(real_body_len(b2.o, b2.c), &cp.short_body, 4);
    let mu_short3 = cp.mu_less(real_body_len(b3.o, b3.c), &cp.short_body, 3);
    let mu_short4 = cp.mu_less(real_body_len(b4.o, b4.c), &cp.short_body, 2);
    let mu_long5 = cp.mu_greater(real_body_len(b5.o, b5.c), &cp.long_body, 1);

    // Determine color of 1st candle: +1 white, -1 black -- crisp sign.
    let color1: f64 = if !is_white(b1.o, b1.c) { -1.0 } else { 1.0 };

    // Color check: white, 3 black, white OR black, 3 white, black -- crisp.
    let c2: f64 = if !is_white(b2.o, b2.c) { -1.0 } else { 1.0 };
    let c3: f64 = if !is_white(b3.o, b3.c) { -1.0 } else { 1.0 };
    let c4: f64 = if !is_white(b4.o, b4.c) { -1.0 } else { 1.0 };
    let c5: f64 = if !is_white(b5.o, b5.c) { -1.0 } else { 1.0 };

    if !(c2 == -color1 && c3 == c2 && c4 == c3 && c5 == -c4) {
        return 0.0;
    }

    // 2nd to 4th hold within 1st: a part of the real body overlaps 1st range -- crisp.
    if !(f64::min(b2.o, b2.c) < b1.h && f64::max(b2.o, b2.c) > b1.l
        && f64::min(b3.o, b3.c) < b1.h && f64::max(b3.o, b3.c) > b1.l
        && f64::min(b4.o, b4.c) < b1.h && f64::max(b4.o, b4.c) > b1.l) {
        return 0.0;
    }

    // 2nd to 4th are falling (rising) -- using color multiply trick -- crisp.
    if !(b3.c * color1 < b2.c * color1 && b4.c * color1 < b3.c * color1) {
        return 0.0;
    }

    // 5th opens above (below) the prior close -- crisp.
    if !(b5.o * color1 > b4.c * color1) {
        return 0.0;
    }

    // 5th closes above (below) the 1st close -- crisp.
    if !(b5.c * color1 > b1.c * color1) {
        return 0.0;
    }

    let conf = fuzzy::t_product_all(&[mu_long1, mu_short2, mu_short3, mu_short4, mu_long5]);
    color1 * conf * 100.0
}
