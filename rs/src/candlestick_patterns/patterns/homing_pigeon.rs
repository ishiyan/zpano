//! Homing Pigeon pattern (2-candle bullish).

use crate::candlestick_patterns::CandlestickPatterns;
use crate::candlestick_patterns::core::{is_black, real_body_len};
use crate::fuzzy;

/// Homing Pigeon: a two-candle bullish pattern.
///
/// Must have:
/// - first candle: long black,
/// - second candle: short black, real body engulfed by first candle's
/// real body.
///
/// The meaning of "long" is specified with self._long_body.
/// The meaning of "short" is specified with self._short_body.
///
/// Category A: always bullish (continuous).
///
/// Returns:
/// Continuous float in [0, 100].  Always bullish.
pub fn homing_pigeon(cp: &CandlestickPatterns) -> f64 {
    if !cp.enough(2, &[&cp.long_body, &cp.short_body]) {
        return 0.0;
    }

    let b1 = cp.bar(2);
    let b2 = cp.bar(1);

    // Crisp gates: both black.
    if !(is_black(b1.o, b1.c) && is_black(b2.o, b2.c)) {
        return 0.0;
    }

    // Fuzzy conditions.
    let mu_long1 = cp.mu_greater(real_body_len(b1.o, b1.c), &cp.long_body, 2);
    let mu_short2 = cp.mu_less(real_body_len(b2.o, b2.c), &cp.short_body, 1);

    // Containment: second body engulfed by first body.
    // For black candles: open > close, so upper = open, lower = close.
    let eq_width = cp.fuzz_ratio * cp.avg_cs(&cp.equal, 2);
    let mu_enc_upper = cp.mu_lt_raw(b2.o, b1.o, eq_width);
    let mu_enc_lower = cp.mu_gt_raw(b2.c, b1.c, eq_width);

    let confidence = fuzzy::t_product_all(&[mu_long1, mu_short2, mu_enc_upper, mu_enc_lower]);
    confidence * 100.0
}
