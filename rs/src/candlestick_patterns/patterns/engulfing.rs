//! Engulfing pattern (2-candle reversal).

use crate::candlestick_patterns::CandlestickPatterns;
use crate::fuzzy;

/// Engulfing: a two-candle reversal pattern.
///
/// Must have:
/// - first candle and second candle have opposite colors,
/// - second candle's real body engulfs the first (at least one end strictly
/// exceeds, the other can match).
///
// Direction sign from 2nd candle (TA-Lib: c >= o is bullish).
/// Category B: direction from 2nd candle color (continuous).
/// Opposite-color check stays crisp (doji edge case).
///
/// Returns:
/// Continuous float in [-100, +100].  Sign from 2nd candle direction.
pub fn engulfing(cp: &CandlestickPatterns) -> f64 {
    if !cp.enough(2, &[]) {
        return 0.0;
    }

    let b1 = cp.bar(2);
    let b2 = cp.bar(1);

    // Opposite colors — crisp gate (TA-Lib convention: c >= o is white).
    let color1: i32 = if b1.c < b1.o { -1 } else { 1 };
    let color2: i32 = if b2.c < b2.o { -1 } else { 1 };
    if color1 == color2 { return 0.0; }

    // Fuzzy engulfment: 2nd body upper >= 1st body upper AND
    // 2nd body lower <= 1st body lower.
    let upper1 = f64::max(b1.o, b1.c);
    let lower1 = f64::min(b1.o, b1.c);
    let upper2 = f64::max(b2.o, b2.c);
    let lower2 = f64::min(b2.o, b2.c);

    // Width based on the equal criterion for tight comparisons.
    let eq_avg = cp.avg_cs(&cp.equal, 1);
    let eq_width = if eq_avg > 0.0 { cp.fuzz_ratio * eq_avg } else { 0.0 };

    let mu_upper = cp.mu_ge_raw(upper2, upper1, eq_width);
    let mu_lower = cp.mu_lt_raw(lower2, lower1, eq_width);

    let confidence = fuzzy::t_product_all(&[mu_upper, mu_lower]);
    // Direction sign from 2nd candle (TA-Lib: c >= o is bullish).
    let direction: f64 = if b2.c < b2.o { -1.0 } else { 1.0 };
    direction * confidence * 100.0
}
