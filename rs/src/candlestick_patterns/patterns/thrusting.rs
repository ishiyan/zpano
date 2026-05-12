//! Thrusting pattern (2-candle bearish).

use crate::candlestick_patterns::CandlestickPatterns;
use crate::candlestick_patterns::core::{is_black, is_white, real_body_len};
use crate::fuzzy;

/// Thrusting: a two-candle bearish continuation pattern.
///
/// Must have:
/// - first candle: long black,
/// - second candle: white, opens below the prior candle's low, closes
/// into the prior candle's real body but below the midpoint, and the
/// close is not equal to the prior candle's close (to distinguish
/// from in-neck).
///
/// The meaning of "long" is specified with self._long_body.
/// The meaning of "equal" is specified with self._equal.
///
/// Category A: always bearish (continuous).
///
/// Returns:
/// Continuous float in [-100, 0].  Always bearish.
pub fn thrusting(cp: &CandlestickPatterns) -> f64 {
    if !cp.enough(2, &[&cp.long_body, &cp.equal]) {
        return 0.0;
    }

    let b1 = cp.bar(2);
    let b2 = cp.bar(1);

    let rb1 = real_body_len(b1.o, b1.c);

    // Crisp gates: color checks and open below prior low.
    if !(is_black(b1.o, b1.c) && is_white(b2.o, b2.c) && b2.o < b1.l) {
        return 0.0;
    }

    // Fuzzy conditions.
    let mu_long1 = cp.mu_greater(rb1, &cp.long_body, 2);

    // Close above prior close + equal avg (not equal to prior close).
    let eq = cp.avg_cs(&cp.equal, 2);
    let eq_width = if eq > 0.0 { cp.fuzz_ratio * eq } else { 0.0 };
    let mu_above_close = cp.mu_gt_raw(b2.c, b1.c + eq, eq_width);

    // Close at or below midpoint of prior body: c2 <= c1 + rb1 * 0.5
    let mid = b1.c + rb1 * 0.5;
    let mid_width = if rb1 > 0.0 { cp.fuzz_ratio * rb1 * 0.5 } else { 0.0 };
    let mu_below_mid = cp.mu_lt_raw(b2.c, mid, mid_width);

    let confidence = fuzzy::t_product_all(&[mu_long1, mu_above_close, mu_below_mid]);

    -1.0 * confidence * 100.0
}
