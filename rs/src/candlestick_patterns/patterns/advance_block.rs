//! Advance Block pattern (3-candle bearish).

use crate::candlestick_patterns::CandlestickPatterns;
use crate::candlestick_patterns::core::{is_white, real_body_len, upper_shadow};
use crate::fuzzy;

/// Advance Block: a bearish three-candle pattern.
///
/// Three white candles with consecutively higher closes and opens, but
/// showing signs of weakening (diminishing bodies, growing upper shadows).
///
/// Category A: always bearish (continuous).
///
/// Returns:
/// Continuous float in [-100, 0].  Always bearish.
pub fn advance_block(cp: &CandlestickPatterns) -> f64 {
    if !cp.enough(3, &[&cp.long_body, &cp.short_shadow, &cp.long_shadow, &cp.near, &cp.far]) {
        return 0.0;
    }

    let b1 = cp.bar(3);
    let b2 = cp.bar(2);
    let b3 = cp.bar(1);

    // Crisp gates: all white with rising closes.
    if !(is_white(b1.o, b1.c) && is_white(b2.o, b2.c) && is_white(b3.o, b3.c)
        && b3.c > b2.c && b2.c > b1.c) {
        return 0.0;
    }
    // Crisp: 2nd opens above 1st open.
    if !(b2.o > b1.o) { return 0.0; }
    // Crisp: 3rd opens above 2nd open.
    if !(b3.o > b2.o) { return 0.0; }

    let rb1 = real_body_len(b1.o, b1.c);
    let rb2 = real_body_len(b2.o, b2.c);
    let rb3 = real_body_len(b3.o, b3.c);

    // Fuzzy: 2nd opens within/near 1st body (upper bound).
    let near3 = cp.avg_cs(&cp.near, 3);
    let near3_width = if near3 > 0.0 { cp.fuzz_ratio * near3 } else { 0.0 };
    let mu_o2_near = cp.mu_lt_raw(b2.o, b1.c + near3, near3_width);

    // Fuzzy: 3rd opens within/near 2nd body (upper bound).
    let near2 = cp.avg_cs(&cp.near, 2);
    let near2_width = if near2 > 0.0 { cp.fuzz_ratio * near2 } else { 0.0 };
    let mu_o3_near = cp.mu_lt_raw(b3.o, b2.c + near2, near2_width);

    // Fuzzy: first candle long body.
    let mu_long1 = cp.mu_greater(rb1, &cp.long_body, 3);
    // Fuzzy: first candle short upper shadow.
    let mu_us1 = cp.mu_less(upper_shadow(b1.o, b1.h, b1.c), &cp.short_shadow, 3);

    // At least one weakness condition must hold (OR -> max).
    let far2 = cp.avg_cs(&cp.far, 3);
    let far2_width = if far2 > 0.0 { cp.fuzz_ratio * far2 } else { 0.0 };
    let far1 = cp.avg_cs(&cp.far, 2);
    let far1_width = if far1 > 0.0 { cp.fuzz_ratio * far1 } else { 0.0 };
    let near1 = cp.avg_cs(&cp.near, 2);
    let near1_width = if near1 > 0.0 { cp.fuzz_ratio * near1 } else { 0.0 };

    // Branch 1: 2 far smaller than 1 AND 3 not longer than 2
    let mu_b1a = cp.mu_lt_raw(rb2, rb1 - far2, far2_width);
    let mu_b1b = cp.mu_lt_raw(rb3, rb2 + near1, near1_width);
    let branch1 = fuzzy::t_product_all(&[mu_b1a, mu_b1b]);

    // Branch 2: 3 far smaller than 2
    let branch2 = cp.mu_lt_raw(rb3, rb2 - far1, far1_width);

    // Branch 3: 3 < 2 AND 2 < 1 AND (3 or 2 has non-short upper shadow)
    let rb3_width = if rb2 > 0.0 { cp.fuzz_ratio * rb2 } else { 0.0 };
    let rb2_width = if rb1 > 0.0 { cp.fuzz_ratio * rb1 } else { 0.0 };
    let mu_b3a = cp.mu_lt_raw(rb3, rb2, rb3_width);
    let mu_b3b = cp.mu_lt_raw(rb2, rb1, rb2_width);
    let mu_b3_us3 = cp.mu_greater(upper_shadow(b3.o, b3.h, b3.c), &cp.short_shadow, 1);
    let mu_b3_us2 = cp.mu_greater(upper_shadow(b2.o, b2.h, b2.c), &cp.short_shadow, 2);
    let branch3 = fuzzy::t_product_all(&[mu_b3a, mu_b3b, f64::max(mu_b3_us3, mu_b3_us2)]);

    // Branch 4: 3 < 2 AND 3 has long upper shadow
    let mu_b4a = cp.mu_lt_raw(rb3, rb2, rb3_width);
    let mu_b4b = cp.mu_greater(upper_shadow(b3.o, b3.h, b3.c), &cp.long_shadow, 1);
    let branch4 = fuzzy::t_product_all(&[mu_b4a, mu_b4b]);

    let weakness = f64::max(f64::max(branch1, branch2), f64::max(branch3, branch4));

    let confidence = fuzzy::t_product_all(&[mu_o2_near, mu_o3_near, mu_long1, mu_us1, weakness]);

    -confidence * 100.0
}
