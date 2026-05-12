//! Harami pattern (2-candle reversal).

use crate::candlestick_patterns::CandlestickPatterns;
use crate::candlestick_patterns::core::{real_body_len};
use crate::fuzzy;

/// Harami: a two-candle reversal pattern.
///
/// Must have:
/// - first candle: long real body,
/// - second candle: short real body contained within the first candle's
/// real body.
///
// Direction: opposite of 1st candle color.
/// Category B: direction from 1st candle color (continuous).
/// Containment degree is fuzzy.
///
/// Returns:
/// Continuous float in [-100, +100].
pub fn harami(cp: &CandlestickPatterns) -> f64 {
    if !cp.enough(2, &[&cp.long_body, &cp.short_body]) {
        return 0.0;
    }

    let b1 = cp.bar(2);
    let b2 = cp.bar(1);

    // Fuzzy size conditions.
    let mu_long1 = cp.mu_greater(real_body_len(b1.o, b1.c), &cp.long_body, 2);
    let mu_short2 = cp.mu_less(real_body_len(b2.o, b2.c), &cp.short_body, 1);

    // Fuzzy containment: 1st body encloses 2nd body.
    let eq_avg = cp.avg_cs(&cp.equal, 1);
    let eq_width = if eq_avg > 0.0 { cp.fuzz_ratio * eq_avg } else { 0.0 };

    let mu_enc_upper = cp.mu_ge_raw(f64::max(b1.o, b1.c), f64::max(b2.o, b2.c), eq_width);
    let mu_enc_lower = cp.mu_lt_raw(f64::min(b1.o, b1.c), f64::min(b2.o, b2.c), eq_width);

    let confidence = fuzzy::t_product_all(&[mu_long1, mu_short2, mu_enc_upper, mu_enc_lower]);
    // Direction: opposite of 1st candle color.
    let direction: f64 = if b1.c < b1.o { 1.0 } else { -1.0 };
    direction * confidence * 100.0
}
