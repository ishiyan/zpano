//! Matching Low pattern (2-candle bullish).

use crate::candlestick_patterns::CandlestickPatterns;
use crate::candlestick_patterns::core::{is_black};

/// Matching Low: a two-candle bullish pattern.
///
/// Must have:
/// - first candle: black,
/// - second candle: black with close equal to the first candle's close.
///
/// The meaning of "equal" is specified with self._equal.
///
/// Category A: always bullish (continuous).
///
/// Returns:
/// Continuous float in [0, 100].  Always bullish.
pub fn matching_low(cp: &CandlestickPatterns) -> f64 {
    if !cp.enough(2, &[&cp.equal]) {
        return 0.0;
    }

    let b1 = cp.bar(2);
    let b2 = cp.bar(1);

    // Crisp gates: both black.
    if !(is_black(b1.o, b1.c) && is_black(b2.o, b2.c)) {
        return 0.0;
    }

    // Fuzzy: close equal to prior close (two-sided band).
    let mu_eq = cp.mu_less((b2.c - b1.c).abs(), &cp.equal, 2);
    mu_eq * 100.0
}
