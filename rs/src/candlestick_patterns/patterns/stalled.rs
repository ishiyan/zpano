//! Stalled pattern (3-candle bearish).

use crate::candlestick_patterns::CandlestickPatterns;
use crate::candlestick_patterns::core::{is_white, real_body_len, upper_shadow};
use crate::fuzzy;

/// Stalled (Deliberation): a three-candle bearish pattern.
///
/// Three white candles with progressively higher closes:
/// - first candle: long white body,
/// - second candle: long white body, opens within or near the first
/// candle's body, very short upper shadow,
/// - third candle: small body that rides on the shoulder of the second
/// (opens near the second's close, accounting for its own body size).
///
/// Category A: always bearish (continuous).
///
/// Returns:
/// Continuous float in [-100, 0].  Always bearish.
pub fn stalled(cp: &CandlestickPatterns) -> f64 {
    if !cp.enough(3, &[&cp.long_body, &cp.short_body, &cp.very_short_shadow, &cp.near]) {
        return 0.0;
    }

    let b1 = cp.bar(3);
    let b2 = cp.bar(2);
    let b3 = cp.bar(1);

    // Crisp gates: all white, rising closes.
    if !(is_white(b1.o, b1.c) && is_white(b2.o, b2.c) && is_white(b3.o, b3.c)) {
        return 0.0;
    }
    if !(b3.c > b2.c && b2.c > b1.c) { return 0.0; }
    // Crisp: o2 > o1 (opens above prior open).
    if !(b2.o > b1.o) { return 0.0; }

    let rb3 = real_body_len(b3.o, b3.c);

    // Fuzzy conditions.
    let mu_long1 = cp.mu_greater(real_body_len(b1.o, b1.c), &cp.long_body, 3);
    let mu_long2 = cp.mu_greater(real_body_len(b2.o, b2.c), &cp.long_body, 2);
    let mu_us2 = cp.mu_less(upper_shadow(b2.o, b2.h, b2.c), &cp.very_short_shadow, 2);

    // o2 <= c1 + near_avg (opens within or near prior body).
    let near3 = cp.avg_cs(&cp.near, 3);
    let near3_width = if near3 > 0.0 { cp.fuzz_ratio * near3 } else { 0.0 };
    let mu_o2_near = cp.mu_lt_raw(b2.o, b1.c + near3, near3_width);

    // Third candle: short body.
    let mu_short3 = cp.mu_less(rb3, &cp.short_body, 1);

    // o3 >= c2 - rb3 - near_avg (rides on shoulder).
    let near2 = cp.avg_cs(&cp.near, 2);
    let near2_width = if near2 > 0.0 { cp.fuzz_ratio * near2 } else { 0.0 };
    let mu_o3_shoulder = cp.mu_ge_raw(b3.o, b2.c - rb3 - near2, near2_width);

    let confidence = fuzzy::t_product_all(&[mu_long1, mu_long2, mu_us2, mu_o2_near, mu_short3, mu_o3_shoulder]);

    -1.0 * confidence * 100.0
}
