//! Counterattack pattern (2-candle reversal).

use crate::candlestick_patterns::CandlestickPatterns;
use crate::candlestick_patterns::core::{real_body_len};
use crate::fuzzy;

/// Counterattack: a two-candle reversal pattern.
///
/// Two long candles of opposite color with closes that are equal
/// (or very near equal).
///
/// - bullish: first candle is long black, second is long white,
/// closes are equal,
/// - bearish: first candle is long white, second is long black,
/// closes are equal.
///
/// The meaning of "long" is specified with self._long_body.
/// The meaning of "equal" is specified with self._equal.
///
// Direction from 2nd candle color.
/// Category B: direction from 2nd candle color (continuous).
///
/// Returns:
/// Continuous float in [-100, +100].
pub fn counterattack(cp: &CandlestickPatterns) -> f64 {
    if !cp.enough(2, &[&cp.long_body, &cp.equal]) {
        return 0.0;
    }

    let b1 = cp.bar(2);
    let b2 = cp.bar(1);

    // Opposite colors — crisp gate.
    let color1: i32 = if b1.c < b1.o { -1 } else { 1 };
    let color2: i32 = if b2.c < b2.o { -1 } else { 1 };
    if color1 == color2 { return 0.0; }

    // Fuzzy conditions.
    let mu_long1 = cp.mu_greater(real_body_len(b1.o, b1.c), &cp.long_body, 2);
    let mu_long2 = cp.mu_greater(real_body_len(b2.o, b2.c), &cp.long_body, 1);
    // Closes near equal: crisp was abs(c2-c1) <= eq.
    // Model as mu_less(abs_diff, eq_avg) — crossover at eq boundary.
    let mu_eq = cp.mu_less((b2.c - b1.c).abs(), &cp.equal, 2);

    let confidence = fuzzy::t_product_all(&[mu_long1, mu_long2, mu_eq]);
    // Direction from 2nd candle color.
    let direction: f64 = if b2.c < b2.o { -1.0 } else { 1.0 };
    direction * confidence * 100.0
}
