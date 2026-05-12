//! Three White Soldiers pattern (3-candle bullish).

use crate::candlestick_patterns::CandlestickPatterns;
use crate::candlestick_patterns::core::{is_white, real_body_len, upper_shadow};
use crate::fuzzy;

/// Three White Soldiers: a three-candle bullish pattern.
///
/// Must have:
/// - three consecutive white candles with consecutively higher closes,
/// - all three have very short upper shadows,
/// - each opens within or near the prior candle's real body,
/// - none is far shorter than the prior candle,
/// - third candle is not short.
///
/// Category A: always bullish (continuous).
///
/// Returns:
/// Continuous float in [0, 100].  Always bullish.
pub fn three_white_soldiers(cp: &CandlestickPatterns) -> f64 {
    if !cp.enough(3, &[&cp.short_body, &cp.very_short_shadow, &cp.near, &cp.far]) {
        return 0.0;
    }

    let b1 = cp.bar(3);
    let b2 = cp.bar(2);
    let b3 = cp.bar(1);

    // Crisp gates: all white with consecutively higher closes.
    if !(is_white(b1.o, b1.c) && is_white(b2.o, b2.c) && is_white(b3.o, b3.c)
        && b3.c > b2.c && b2.c > b1.c) {
        return 0.0;
    }

    let rb1 = real_body_len(b1.o, b1.c);
    let rb2 = real_body_len(b2.o, b2.c);
    let rb3 = real_body_len(b3.o, b3.c);

    // Crisp: each opens above the prior open (ordering).
    if !(b2.o > b1.o && b3.o > b2.o) { return 0.0; }

    // Fuzzy: very short upper shadows (all three).
    let mu_us1 = cp.mu_less(upper_shadow(b1.o, b1.h, b1.c), &cp.very_short_shadow, 3);
    let mu_us2 = cp.mu_less(upper_shadow(b2.o, b2.h, b2.c), &cp.very_short_shadow, 2);
    let mu_us3 = cp.mu_less(upper_shadow(b3.o, b3.h, b3.c), &cp.very_short_shadow, 1);

    // Fuzzy: each opens within or near the prior body (upper bound).
    let near3 = cp.avg_cs(&cp.near, 3);
    let near3_width = if near3 > 0.0 { cp.fuzz_ratio * near3 } else { 0.0 };
    let mu_o2_near = cp.mu_lt_raw(b2.o, b1.c + near3, near3_width);

    let near2 = cp.avg_cs(&cp.near, 2);
    let near2_width = if near2 > 0.0 { cp.fuzz_ratio * near2 } else { 0.0 };
    let mu_o3_near = cp.mu_lt_raw(b3.o, b2.c + near2, near2_width);

    // Fuzzy: not far shorter than prior candle.
    let far3 = cp.avg_cs(&cp.far, 3);
    let far3_width = if far3 > 0.0 { cp.fuzz_ratio * far3 } else { 0.0 };
    let mu_not_far2 = cp.mu_gt_raw(rb2, rb1 - far3, far3_width);

    let far2 = cp.avg_cs(&cp.far, 2);
    let far2_width = if far2 > 0.0 { cp.fuzz_ratio * far2 } else { 0.0 };
    let mu_not_far3 = cp.mu_gt_raw(rb3, rb2 - far2, far2_width);

    // Fuzzy: third candle is not short.
    let mu_not_short3 = cp.mu_greater(rb3, &cp.short_body, 1);

    let confidence = fuzzy::t_product_all(&[
        mu_us1, mu_us2, mu_us3,
        mu_o2_near, mu_o3_near,
        mu_not_far2, mu_not_far3,
        mu_not_short3,
    ]);

    confidence * 100.0
}
